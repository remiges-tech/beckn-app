resource "google_service_account" "cloud_run_sa" {
  account_id   = "beckn-app-run-sa"
  project      = var.project_id
  display_name = "beckn-app Cloud Run runtime service account"
}

# Needed to construct the Cloud Run Service Agent's identity below
# (service-<PROJECT_NUMBER>@serverless-robot-prod.iam.gserviceaccount.com).
data "google_project" "trade" {
  project_id = var.project_id
}

resource "google_project_iam_member" "artifact_reader" {
  project = var.project_id
  role    = "roles/artifactregistry.reader"
  member  = "serviceAccount:${google_service_account.cloud_run_sa.email}"
}

# Cross-project — the grant has to live in the project that owns the Cloud
# SQL instance (remiges-ion), not in remiges-trade.
resource "google_project_iam_member" "cloud_sql_client" {
  provider = google.remiges_ion
  project  = var.remiges_ion_project_id
  role     = "roles/cloudsql.client"
  member   = "serviceAccount:${google_service_account.cloud_run_sa.email}"
}

# Grants Direct VPC Egress rights scoped to just the one existing subnet
# being shared (not project-wide network access in remiges-ion).
#
# IMPORTANT: this must go to the Cloud Run SERVICE AGENT
# (service-<PROJECT_NUMBER>@serverless-robot-prod.iam.gserviceaccount.com),
# not the workload runtime service account (cloud_run_sa) — Direct VPC
# Egress attachment to a Shared VPC subnet is performed by the platform's own
# service agent, not the container's runtime identity. Confirmed against
# https://docs.cloud.google.com/run/docs/configuring/shared-vpc-direct-vpc.
resource "google_compute_subnetwork_iam_member" "network_user" {
  provider   = google.remiges_ion
  project    = var.remiges_ion_project_id
  region     = var.region
  subnetwork = data.google_compute_subnetwork.existing_subnet.name
  role       = "roles/compute.networkUser"
  member     = "serviceAccount:service-${data.google_project.trade.number}@serverless-robot-prod.iam.gserviceaccount.com"

  depends_on = [google_project_service.apis]
}
