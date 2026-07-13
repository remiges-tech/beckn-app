# Infra setup — from scratch

This is the master checklist for standing up beckn-app's infrastructure,
local and cloud, from a completely fresh clone. It's split into two
independent tracks:

- **[Local development](#1-local-development-docker-compose)** — everything
  runs in Docker Compose on your machine. Start here; it's the fastest way to
  get a working end-to-end system and to generate/verify the signing keys
  you'll reuse in the cloud track.
- **[Cloud deployment](#2-cloud-deployment-cloud-run)** — deploys the same
  six services to Google Cloud Run. Full detail lives in
  [`terraform/README.md`](terraform/README.md); this section is the
  condensed version plus how it fits together with the local track.

## The six services

| Service | What it is | Local port | Local dev source |
|---|---|---|---|
| `bap` | Buyer-side app (Go) | 8083 | `bap-application/` |
| `bpp` | Seller-side app (Go) | 8080 | `bpp-application/` |
| `bap-frontend` | Buyer web UI | 3000 | `bap-frontend/` |
| `bpp-frontend` | Seller/admin web UI | 3001 | `bpp-frontend/` |
| `onix-bap` | Beckn network adapter for bap | 8081 | `onix-bap/` |
| `onix-bpp` | Beckn network adapter for bpp | 8082 | `onix-bpp/` |

Plus Postgres and two Redis instances, all defined in `docker-compose.yml`
for local dev (in the cloud, these are *reused* from an existing project's
Cloud SQL/Memorystore — see the cloud section).

## 1. Local development (Docker Compose)

### 1a. Prerequisites

- Docker + Docker Compose.
- A registered Beckn network identity for both `bap` and `bpp` — a
  subscriber ID, a `keyId`, and an Ed25519 keypair for signing (`signingPrivateKey`/
  `signingPublicKey`) plus an encryption keypair (`encrPrivateKey`/`encrPublicKey`),
  each registered on the network registry you're testing against
  (`fabric.nfh.global` in this codebase's current setup). **You cannot
  generate these yourself and expect them to work** — they have to actually
  be registered so other participants can verify your signatures. Ask
  whoever administers the network registry for a subscriber slot.

### 1b. Config files you must create (all git-ignored, never committed)

1. **`.env`** (repo root) — copy `.env.example`, fill in your registered
   keys:
   ```
   BAP_PRIVATE_KEY=<your bap signingPrivateKey>
   BAP_KEY_ID=<bap subscriber id>|<bap keyId>|ed25519
   BPP_PRIVATE_KEY=<your bpp signingPrivateKey>
   BPP_KEY_ID=<bpp subscriber id>|<bpp keyId>|ed25519
   ```
   **Important:** `BAP_PRIVATE_KEY`/`BAP_KEY_ID` must be the *same keypair
   and subscriber identity* as whatever you put in `config/local-simple-bap.yaml`'s
   `keyManager` block below — there's only one registered identity per side,
   used for both the onix adapter's own signing and the app's direct CDS
   calls. Same rule for bpp. (See `terraform/README.md`'s Gotcha #1 if you
   want the full "why.")

2. **`config/local-simple-bap.yaml`** and **`config/local-simple-bpp.yaml`**
   — copy the `.example` versions next to them and fill in the same 5
   `keyManager` fields (`networkParticipant`, `keyId`, `signingPrivateKey`,
   `signingPublicKey`, `encrPrivateKey`, `encrPublicKey`) with your
   registered identity, **twice each** — both the receiver and caller
   modules in the file need the identical block.

3. **`docker-compose.yml`**'s `BAP_ID`/`BPP_ID` env vars — these must equal
   the same subscriber ID you used above. Edit them directly if your
   registered identity isn't already what's there.

### 1c. Bring it up

```sh
docker compose up -d --build
```

This builds and starts all 6 app services + Postgres + 2 Redis instances.
First run also needs the database schema created — migrations aren't run
automatically. Requires the `tern` CLI (`go install github.com/jackc/tern/v2@latest`);
each `tern.conf` already defaults to `docker-compose.yml`'s exposed Postgres
port (`5435`), so no extra config is needed:

```sh
cd bap-application/db/migrations && tern migrate && cd ../../..
cd bpp-application/db/migrations && tern migrate && cd ../../..
```

### 1d. Verify

```sh
curl http://localhost:8083/health   # bap
curl http://localhost:8080/health   # bpp
curl "http://localhost:3000/api/v1/discover?textSearch=coffee"   # full discover flow through the frontend proxy
```

A `200` with real catalog data back means signing, routing, and the CDS
connection are all working.

### 1e. Known local-dev gotcha

If you rebuild/recreate `bap`, `bpp`, `onix-bap`, or `onix-bpp` without also
restarting `frontend`/`bpp-frontend`, the frontends' nginx will keep talking
to the *old* container IP and you'll get `502 Bad Gateway`. Either
`docker compose up -d` the whole stack together, or explicitly
`docker compose restart frontend bpp-frontend` after rebuilding a backend.

## 2. Cloud deployment (Cloud Run)

Full step-by-step is in [`terraform/README.md`](terraform/README.md) —
including the org/IAM prerequisites, the exact bootstrap command sequence,
and 11 specific, real gotchas encountered getting this working (not
hypothetical ones). Condensed version:

1. Own (or get access to) a GCP project to deploy into, and — if you're
   reusing an existing Cloud SQL/Memorystore from a *different* project like
   this deployment does — confirm that other project belongs to a GCP
   organization (Shared VPC requires it) and you can get
   `roles/compute.xpnAdmin` granted at the org level, at least temporarily.
2. Create the Terraform state bucket, `terraform.tfvars` with your project
   ID, and bootstrap APIs + Artifact Registry + empty secret containers.
3. Seed secrets — same identity/key material as local dev's `.env` and
   `config/local-simple-{bap,bpp}.yaml`, if you want cloud and local to
   represent the same registered participant (recommended, see below).
4. Build and push all 8 images (6 services + 2 migrate-job images).
5. Set up Shared VPC, then apply everything else. The very first full apply
   will fail on `bap`/`bpp` — that's expected, run the migrate jobs and
   apply again (see `terraform/README.md`'s bootstrap sequence for why).

## 3. Keeping local and cloud in sync

Whenever the registered network identity or signing keys change (e.g. a new
subscriber registration, a key rotation), the same values need updating in
**four** places, or the two environments will drift and start failing with
subscriber-identity-mismatch errors:

| What | Local dev | Cloud |
|---|---|---|
| App-level signing key (`BAP_PRIVATE_KEY`/`BAP_KEY_ID`, `BPP_PRIVATE_KEY`/`BPP_KEY_ID`) | root `.env` | `bap-private-key`/`bap-key-id`/`bpp-private-key`/`bpp-key-id` secrets in Secret Manager |
| Onix keyManager identity (`networkParticipant`, `keyId`, signing/encr keys) | `config/local-simple-{bap,bpp}.yaml` | `bap-onix-*`/`bpp-onix-*` secrets in Secret Manager |
| Subscriber ID used in `context.bapId`/`bppId` | `docker-compose.yml`'s `BAP_ID`/`BPP_ID` | `terraform/variables.tf`'s `bap_id`/`bpp_id` |
| Network ID | `docker-compose.yml`'s `NETWORK_ID` (+ both apps' `cmd/server/.env`) | `terraform/variables.tf`'s `network_id` |

After any cloud secret update, the affected Cloud Run service needs a forced
new revision to actually pick up the change — see `terraform/README.md`'s
Gotcha #11.

## 4. Known unresolved issue

`select`/`init`/`confirm` currently fail in the cloud deployment (and would
hit the same issue locally, if a registry-based routing lookup for `bpp_id`
ever resolved to a bare, no-path URL): the network registry stores a bare
Subscriber URL (no path), but other participants doing a registry lookup
append the action name directly (`<url>/select`), which doesn't match the
onix adapter's actual listening path (`/bpp/receiver/select`). Not yet
decided how to resolve — options are documented in `terraform/README.md`'s
Gotcha #8.
