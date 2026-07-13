# GCP deployment (Cloud Run, scale-to-zero)

Deploys beckn-app's six services (bap, bpp, bap-frontend, bpp-frontend,
onix-bap, onix-bpp) to Cloud Run in the `remiges-trade` project, reusing the
**existing** Cloud SQL instance (`insurance-postgres`) and Memorystore
instance (`insurance-redis`) that live in the separate `remiges-ion` project
— no new database/cache instances are created.

This directory replaces `docker-compose.yml`'s Docker-network hostnames
(`http://bap:8083`, `redis-bap:6379`, ...) with Cloud Run's per-service HTTPS
URLs and the reused instances' addresses.

## Architecture notes

- **Cross-project reuse**: remiges-ion is a Shared VPC host project and
  remiges-trade is attached as a service project, so Cloud Run in
  remiges-trade can reach `insurance-app-vpc` via Direct VPC Egress. This is
  required because Memorystore has no public-IP option at all. See
  `network.tf`.
- **New databases, not a new instance**: `trade_bap`/`trade_bpp` databases +
  a `trade` SQL user are created on the *existing* `insurance-postgres`
  instance, namespaced separately from the insurance app's own
  `insurance_bap`/`insurance_bpp` databases. See `sql.tf`.
- **Shared Redis**: onix-bap and onix-bpp both point at the one existing
  `insurance-redis` instance — same instance the insurance app's own onix
  adapters use. This has actually caused a stale/foreign cached routing
  decision to surface once in practice (see Gotchas below) — not just a
  theoretical risk.
- **Custom onix images**: `onix-bap/` and `onix-bpp/` wrap the vendor's
  `fidedocker/onix-adapter` image with an `entrypoint.sh` that runs
  `envsubst` over `.yaml.tmpl` templates (the real `keyManager`/`router`/
  `schemaValidator` config) at container startup, substituting only the
  small set of env vars Cloud Run actually varies at runtime (Redis address,
  cross-service URLs, signing keys). Everything else is baked into the image
  at build time.
- **Cloud SQL Auth Proxy sidecar**: bap/bpp/the migrate jobs reach Postgres
  via a `cloudsql-proxy` sidecar container on `127.0.0.1:5432`, not the
  native `/cloudsql` socket — zero app-code changes needed (`DB_HOST` is
  just `127.0.0.1`).
- **Circular URL dependencies**: bap↔onix-bap and bpp↔onix-bpp each
  reference the other's Cloud Run URL. Terraform can't express a cycle, so
  those specific env vars are set to a bootstrap placeholder in the initial
  resource declarations and patched to the real value by
  `cloud-run-wiring.tf` (a `null_resource` + `gcloud run services update`)
  once every service's URL is known. This step re-runs on every
  `terraform apply` (deliberately — see Gotchas).
- **Public access**: all six services grant `roles/run.invoker` to
  `allUsers` — no auth. Documented tradeoff, not an oversight; see
  `iam-invoker.tf`.
- **Network identity**: `BAP_ID`/`BPP_ID` (context.bapId/bppId in every
  outbound message) and the onix keyManager's `NETWORK_PARTICIPANT` share
  the *same* Terraform variable (`var.bap_id`/`var.bpp_id` — see
  `variables.tf`). These two uses must always be the same value, and
  `BAP_PRIVATE_KEY`/`BAP_KEY_ID` (the app's own key for signing direct CDS
  calls) must also be the *same keypair* as the onix keyManager's signing
  key — there is only one registered identity per side, not two. See the
  first Gotcha below for what happens if this drifts.

## Prerequisites

- `gcloud` CLI, authenticated, with Owner (or equivalent) on both
  `remiges-trade` and `remiges-ion`.
- **`remiges-ion` must belong to a GCP organization.** Shared VPC host
  project enablement fails with `Error 400: Invalid resource usage: 'The
  project has no organization.'` on a fully standalone project. If it's
  standalone, move it first:
  ```sh
  gcloud beta projects move remiges-ion --organization=<ORG_ID>
  ```
- **Shared VPC Admin** (`roles/compute.xpnAdmin`) at the organization level —
  plain project Owner is NOT sufficient for
  `google_compute_shared_vpc_host_project`. Grant it to yourself if you have
  `roles/resourcemanager.organizationAdmin` (or ask someone who does):
  ```sh
  gcloud organizations add-iam-policy-binding <ORG_ID> \
    --member="user:you@example.com" --role="roles/compute.xpnAdmin"
  ```
  This only needs to exist for the one-time Shared VPC setup step below —
  safe to revoke afterward (`remove-iam-policy-binding` with the same args),
  since it isn't needed again unless Shared VPC is destroyed and recreated.
- Application Default Credentials set up: `gcloud auth application-default
  login` and `gcloud auth application-default set-quota-project
  remiges-trade` (separate from `gcloud auth login` — Terraform's provider
  reads a different credential store).
- `terraform` >= 1.5, `docker`.
- The wiring step shells out to `gcloud run services update` during
  `terraform apply` — make sure `gcloud` is authenticated against
  `remiges-trade` on the machine running `terraform apply`.

## Bootstrap sequence

1. One-time state bucket (chicken-and-egg — Terraform can't create the
   bucket it stores its own state in):
   ```sh
   gcloud storage buckets create gs://remiges-trade-tfstate \
     --project=remiges-trade --location=asia-southeast2 \
     --uniform-bucket-level-access
   gcloud storage buckets update gs://remiges-trade-tfstate --versioning
   ```

2. Create `terraform/terraform.tfvars` (not committed) with just your project ID:
   ```hcl
   project_id = "remiges-trade"
   ```

3. Bootstrap the things later steps depend on:
   ```sh
   cd terraform
   terraform init
   terraform plan -target=google_project_service.apis \
     -target=google_artifact_registry_repository.repo \
     -target=google_secret_manager_secret.this -out=bootstrap1.tfplan
   terraform apply bootstrap1.tfplan
   ```

4. Seed the 14 externally-provided secrets (everything except
   `postgres-password`, which Terraform generates itself):
   ```sh
   PROJECT_ID=remiges-trade \
     BAP_PRIVATE_KEY=... BAP_KEY_ID=... BPP_PRIVATE_KEY=... BPP_KEY_ID=... \
     BAP_ONIX_KEY_ID=... BAP_ONIX_SIGNING_PRIVATE_KEY=... BAP_ONIX_SIGNING_PUBLIC_KEY=... \
     BAP_ONIX_ENCR_PRIVATE_KEY=... BAP_ONIX_ENCR_PUBLIC_KEY=... \
     BPP_ONIX_KEY_ID=... BPP_ONIX_SIGNING_PRIVATE_KEY=... BPP_ONIX_SIGNING_PUBLIC_KEY=... \
     BPP_ONIX_ENCR_PRIVATE_KEY=... BPP_ONIX_ENCR_PUBLIC_KEY=... \
     ../scripts/seed-secrets.sh
   ```
   `BAP_PRIVATE_KEY`/`BAP_KEY_ID` MUST be the same keypair/identity as
   `BAP_ONIX_SIGNING_PRIVATE_KEY`/`BAP_ONIX_KEY_ID` — see the network identity
   note above. The real values live in this project's git-ignored
   `config/local-simple-{bap,bpp}.yaml` and root `.env`, if you want this
   deployment to mirror local dev's identity.

5. Build and push all 8 images:
   ```sh
   PROJECT_ID=remiges-trade REGION=asia-southeast2 ../scripts/build-and-push.sh
   ```

6. Set up Shared VPC (must exist before any `vpc_access` block validates):
   ```sh
   terraform plan -target=google_compute_shared_vpc_host_project.host \
     -target=google_compute_shared_vpc_service_project.service \
     -target=google_compute_subnetwork_iam_member.network_user -out=bootstrap2.tfplan
   terraform apply bootstrap2.tfplan
   ```

7. First full apply — **this will fail on `bap`/`bpp` the first time**, see
   Gotchas below for why. That's expected:
   ```sh
   terraform plan -out=full.tfplan
   terraform apply full.tfplan
   ```

8. Run migrations (the databases are empty until this runs — `bap`/`bpp`
   crash-loop on startup without it):
   ```sh
   gcloud run jobs execute migrate-bap --project=remiges-trade --region=asia-southeast2 --wait
   gcloud run jobs execute migrate-bpp --project=remiges-trade --region=asia-southeast2 --wait
   ```

9. Re-apply — this time `bap`/`bpp` start successfully against the now-real schema:
   ```sh
   terraform plan -out=full2.tfplan
   terraform apply full2.tfplan
   ```

## Gotchas actually hit doing this (not theoretical — each one broke a real deploy)

1. **`BAP_ID`/`NETWORK_PARTICIPANT`/`BAP_KEY_ID` must all agree.** onix signs
   a message using the keyset registered for whatever `subscriber_id` the
   message *itself* claims to be from (`context.bapId`), not just whatever
   identity the onix pod happens to be configured as. If `BAP_ID` says one
   subscriber but the onix keyManager only has keys for a different one:
   `"Keyset not found for keyID: ..."` → 500. Same rule applies to
   `BAP_PRIVATE_KEY`/`BAP_KEY_ID` (used for direct-to-CDS signing, bypassing
   onix) — the CDS itself will reject with `"subscriber identity mismatch"`
   if that keypair's embedded subscriber doesn't match `context.bapId`
   either. There is exactly one registered identity per side; every signing
   path must use it.
2. **`sqladmin.googleapis.com` must be enabled in the *Cloud Run* project
   (remiges-trade), not just wherever the Cloud SQL instance lives.** The
   Cloud SQL Auth Proxy sidecar calls the Cloud SQL Admin API using the
   calling project's own service account/quota context. Symptom:
   `cloudsql-proxy` container fails with `"Cloud SQL Admin API has not been
   used in project <number> before or it is disabled."` even though the
   instance's own project has it enabled.
3. **`tern.conf` renders as a Go template — literal values are never read
   from the environment.** `jackc/tern` parses `tern.conf` through
   `text/template` with Masterminds/sprig functions. A plain
   `host = localhost` line is never substituted; it takes `{{env "PGHOST" |
   default "localhost"}}` syntax to actually pick up `PGHOST` at runtime.
   Without this, the migrate job silently connects to whatever
   host/port/user/db is hardcoded in the file, not the Cloud SQL Auth Proxy
   sidecar — and fails with a plain connection-refused error that doesn't
   obviously point at the real cause.
4. **`roles/compute.networkUser` for Direct VPC Egress with Shared VPC must
   go to the Cloud Run *Service Agent*
   (`service-<PROJECT_NUMBER>@serverless-robot-prod.iam.gserviceaccount.com`),
   not the workload's runtime service account.** Granting it to the runtime
   SA instead looks plausible (it's the identity the *rest* of the app's
   permissions go to) but every service with a `vpc_access` block will fail
   to start with a Direct VPC Egress permission error.
5. **`gcloud run services update`/`jobs update`'s container-scoped flags
   (`--update-env-vars`, `--image`, etc.) need an explicit `--container=NAME`
   once a service/job has more than one container** (e.g. the
   `cloudsql-proxy` sidecar). Omitting it is either ambiguous or targets the
   wrong container depending on gcloud version.
6. **Cloud Run rejects `memory < 512Mi` when CPU is "always allocated"**
   (the default). The frontend nginx containers hit this at 256Mi even
   though nginx barely needs it — bumped to 512Mi.
7. **First `terraform apply` will fail on `bap`/`bpp` even with everything
   else correct**, because the databases have no schema yet and both apps
   `os.Exit(1)` if a required startup query fails against a table that
   doesn't exist. The migrate jobs (Cloud Run *jobs*, not services) get
   created fine in that same apply since jobs don't block on a health
   check — run them, then re-apply. This is why the bootstrap sequence
   above has an apply → migrate → apply-again shape instead of one apply.
8. **The registered Subscriber URL and the onix module's listening path
   must agree.** If a participant's registry entry is a bare URL (e.g.
   `https://onix-bpp-xxx.a.run.app`, no path), other participants doing a
   registry-based lookup will call `<that-bare-url>/<action-name>` directly
   — e.g. `.../select` — not `<url>/bpp/receiver/select`. If your onix
   adapter's receiver module is only configured to listen at `/bpp/receiver/`
   (the vendor's default), that request 404s. This is a known, currently
   **unresolved** issue on this deployment (`select` fails this way) — the
   real fix is either moving the receiver module to listen at `/`, or
   re-registering with the `/receiver` suffix in the URL, and hasn't been
   decided yet.
9. **Shared Redis across tenants can serve stale/foreign cached routing
   decisions.** Once, `onix-bap`'s router resolved `select` to a completely
   unrelated (and long-dead) external URL that had nothing to do with the
   current request — almost certainly a cache key collision on the shared
   `insurance-redis` instance. It self-corrected on a later attempt (cache
   entry expired), but there's no dedicated Redis/DB-index isolation between
   this app and whatever else uses that instance.
10. **Cloud Run pins an image tag to a specific digest at deploy time and
    never re-resolves it.** Pushing a new image under the same `:latest` tag
    does nothing to an already-running service/job — `lifecycle {
    ignore_changes = [...] }` on every Cloud Run resource means Terraform
    won't touch it either. Force it with `gcloud run services update
    --container=<name> --image=<new-ref>` (services) or `gcloud run jobs
    update --container=<name> --image=<new-ref>` (jobs).
11. **Secret Manager's `version = "latest"` doesn't propagate to already-running
    instances.** Adding a new secret version doesn't do anything until a
    *new revision* starts (secrets are resolved at container start, not
    live). Any change to a secret value needs a forced redeploy of whatever
    service/job reads it — a `terraform apply` that touches an unrelated
    field on that resource is enough, since it creates a new revision.

## Verification

- `terraform output bap_frontend_url bpp_frontend_url` — open both, confirm the SPA loads.
- `gcloud run jobs executions list --job=migrate-bap --project=remiges-trade` — confirm `Succeeded` (same for migrate-bpp).
- Repeat the discover → select flow against the new `bap_url`, watching
  `gcloud logging read 'resource.type="cloud_run_revision" AND resource.labels.service_name="bap"' --project=remiges-trade --freshness=5m`
  (swap the service name for `onix-bap`/`bpp`/`onix-bpp` as needed) for
  signing errors or connectivity failures.
- A cloudsql-proxy sidecar failing its startup probe means the
  `roles/cloudsql.client` grant in remiges-ion isn't working; onix logs
  timing out reaching Redis means the `compute.networkUser` subnet grant
  isn't working.
- Seed a catalog via `catalog-seed/` against the new `bpp_url`.

## Redeploying after a code change

```sh
PROJECT_ID=remiges-trade REGION=asia-southeast2 ../scripts/build-and-push.sh "$(git rev-parse --short HEAD)"
terraform apply -var="image_tag=$(git rev-parse --short HEAD)"
```
(The `lifecycle { ignore_changes = [...] }` block on each Cloud Run resource
means a plain `terraform apply` with no tag change won't redeploy a new
image — you must pass a new `image_tag`. For the two migrate jobs
specifically, this doesn't apply the new image either — see Gotcha #10.)

## Inspecting the reused Cloud SQL instance

It has no public IP and isn't reachable by `cloud-sql-proxy` run outside a
VPC-attached environment (your laptop won't reach the private IP even with
`--private-ip` — only Cloud Run's Direct VPC Egress can). Use **Cloud SQL
Studio** in the console instead — no network path needed, goes through the
Cloud SQL Admin API:
```
https://console.cloud.google.com/sql/instances/insurance-postgres/studio?project=remiges-ion
```
Log in inside Studio with the `trade` user and the `postgres-password`
secret's value (`gcloud secrets versions access latest --secret=postgres-password --project=remiges-trade`).
