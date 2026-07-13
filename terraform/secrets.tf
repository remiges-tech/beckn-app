# Every secret Cloud Run reads lives in Secret Manager, never as a plaintext
# env var or in any Terraform-visible file. Terraform creates the containers
# for ALL of them (so IAM access and references are declarative), but only
# ever writes a VALUE into the one it can generate itself (the Postgres
# password). The rest are empty containers until seeded once via
# scripts/seed-secrets.sh — see terraform/README.md.
#
# IMPORTANT ORDERING: a Cloud Run revision fails to deploy if it references a
# secret version that doesn't exist yet. Run scripts/seed-secrets.sh after
# `terraform apply -target=google_secret_manager_secret.this` and BEFORE the
# full `terraform apply` that creates the Cloud Run services.
#
# There are two independent key pairs per side (unlike a single shared key):
# the Go app's own CDS-signing key (bap-private-key/bap-key-id,
# bpp-private-key/bpp-key-id) and the onix adapter's keyManager block
# (bap-onix-*/bpp-onix-*, 5 fields each).

locals {
  externally_seeded_secret_ids = [
    "bap-private-key",
    "bap-key-id",
    "bpp-private-key",
    "bpp-key-id",
    "bap-onix-key-id",
    "bap-onix-signing-private-key",
    "bap-onix-signing-public-key",
    "bap-onix-encr-private-key",
    "bap-onix-encr-public-key",
    "bpp-onix-key-id",
    "bpp-onix-signing-private-key",
    "bpp-onix-signing-public-key",
    "bpp-onix-encr-private-key",
    "bpp-onix-encr-public-key",
  ]

  all_secret_ids = concat(["postgres-password"], local.externally_seeded_secret_ids)
}

resource "google_secret_manager_secret" "this" {
  for_each   = toset(local.all_secret_ids)
  secret_id  = each.key
  project    = var.project_id
  depends_on = [google_project_service.apis]

  replication {
    auto {}
  }
}

# The only secret value Terraform ever generates or sees itself.
resource "google_secret_manager_secret_version" "postgres_password" {
  secret      = google_secret_manager_secret.this["postgres-password"].id
  secret_data = random_password.postgres.result
}

resource "google_secret_manager_secret_iam_member" "accessor" {
  for_each  = google_secret_manager_secret.this
  secret_id = each.value.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.cloud_run_sa.email}"
}
