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

- **Cross-project reuse**: remiges-ion is made a Shared VPC host project and
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
  adapters use. Cache keys are namespaced by transaction/message UUID, so
  collision risk is low but not zero.
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
  those specific env vars are left blank in the initial resource
  declarations and patched in by `cloud-run-wiring.tf` (a `null_resource` +
  `gcloud run services update`) once every service's URL is known. This
  step re-runs on every `terraform apply`.
- **Public access**: all six services grant `roles/run.invoker` to
  `allUsers` — no auth. Documented tradeoff, not an oversight; see
  `iam-invoker.tf`.

## Prerequisites

- `gcloud` CLI, authenticated, with Owner (or equivalent) on both
  `remiges-trade` and `remiges-ion`.
- **Shared VPC Admin** (`roles/compute.xpnAdmin`) at the organization or
  folder level — plain project Owner is not always sufficient for
  `google_compute_shared_vpc_host_project`/`_service_project`. If `terraform
  apply` fails on those two resources with a permission error, this is why.
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
   terraform apply \
     -target=google_project_service.apis \
     -target=google_artifact_registry_repository.repo \
     -target=google_secret_manager_secret.this
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
   The real values already exist in this project's git-ignored
   `config/local-simple-{bap,bpp}.yaml` and root `.env`, if you want this
   deployment to mirror local dev's identity.

5. Build and push all 8 images:
   ```sh
   PROJECT_ID=remiges-trade REGION=asia-southeast2 ../scripts/build-and-push.sh
   ```

6. Set up Shared VPC (must exist before any `vpc_access` block validates):
   ```sh
   terraform apply \
     -target=google_compute_shared_vpc_host_project.host \
     -target=google_compute_shared_vpc_service_project.service \
     -target=google_compute_subnetwork_iam_member.network_user
   ```

7. Full apply — creates the databases/user, all 6 Cloud Run services, the 2
   migrate jobs, the wiring patch, and public IAM invoker bindings:
   ```sh
   terraform apply
   ```

8. Run migrations:
   ```sh
   gcloud run jobs execute migrate-bap --project=remiges-trade --region=asia-southeast2 --wait
   gcloud run jobs execute migrate-bpp --project=remiges-trade --region=asia-southeast2 --wait
   ```

## Verification

- `terraform output bap_frontend_url bpp_frontend_url` — open both, confirm the SPA loads.
- `gcloud run jobs executions list --job=migrate-bap --project=remiges-trade` — confirm `Succeeded` (same for migrate-bpp).
- Repeat the discover → select flow against the new `bap_url`, watching
  `gcloud run services logs read bap|onix-bap|bpp|onix-bpp --project=remiges-trade`
  for signing errors or connectivity failures.
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
image — you must pass a new `image_tag`.)
