# Default provider — everything this module actually CREATES (Cloud Run,
# Artifact Registry, Secret Manager, service accounts) lives in remiges-trade.
provider "google" {
  project = var.project_id
  region  = var.region
}

# Aliased provider — used only for the handful of things that must reference
# or modify remiges-ion: the existing Cloud SQL instance, the existing
# Memorystore instance, the existing VPC/subnet, Shared VPC host/service
# project attachment, and the cross-project IAM grants those require.
provider "google" {
  alias   = "remiges_ion"
  project = var.remiges_ion_project_id
  region  = var.region
}

locals {
  # APIs needed in remiges-trade. sqladmin IS needed here even though no SQL
  # instance is created in this project — the Cloud SQL Auth Proxy sidecar
  # (bap/bpp/migrate jobs) calls the Cloud SQL Admin API using remiges-trade's
  # own service account/quota context to resolve the (remiges-ion-owned)
  # instance's connection metadata, so it must be enabled on the CALLING
  # project, not just wherever the instance lives. Confirmed by a real
  # deploy failure: "Cloud SQL Admin API has not been used in project
  # <remiges-trade-number> before or it is disabled." redis/servicenetworking
  # remain excluded — nothing in this project talks to those APIs directly.
  required_apis = [
    "run.googleapis.com",
    "artifactregistry.googleapis.com",
    "secretmanager.googleapis.com",
    "iam.googleapis.com",
    "compute.googleapis.com",
    "sqladmin.googleapis.com",
  ]
}

resource "google_project_service" "apis" {
  for_each                   = toset(local.required_apis)
  project                    = var.project_id
  service                    = each.key
  disable_dependent_services = false
  disable_on_destroy         = false
}
