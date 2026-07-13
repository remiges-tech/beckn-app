# Reuses remiges-ion's existing VPC/subnet — does NOT create a new network.
# Cloud Run in remiges-trade needs Direct VPC Egress into this network to
# reach both the existing Cloud SQL instance (private IP) and the existing
# Memorystore instance (which has no public-IP option at all), so
# remiges-trade must be attached to it via Shared VPC.

data "google_compute_network" "existing_vpc" {
  provider = google.remiges_ion
  name     = var.vpc_name
  project  = var.remiges_ion_project_id
}

data "google_compute_subnetwork" "existing_subnet" {
  provider = google.remiges_ion
  name     = var.subnet_name
  region   = var.region
  project  = var.remiges_ion_project_id
}

# Makes remiges-ion a Shared VPC host project. Requires Shared VPC Admin
# rights (roles/compute.xpnAdmin) at the ORG or FOLDER level to apply — plain
# project Owner on remiges-ion is not always sufficient. Does not disrupt any
# existing resource in remiges-ion; it only enables attaching service
# projects.
resource "google_compute_shared_vpc_host_project" "host" {
  provider = google.remiges_ion
  project  = var.remiges_ion_project_id
}

# Attaches remiges-trade as a service project under that host. Requires
# compute.googleapis.com enabled in remiges-trade first (the attach call
# fails against a project where Compute Engine API isn't on yet), and that
# isn't otherwise inferrable from any attribute this resource reads.
resource "google_compute_shared_vpc_service_project" "service" {
  provider        = google.remiges_ion
  host_project    = var.remiges_ion_project_id
  service_project = var.project_id

  depends_on = [
    google_compute_shared_vpc_host_project.host,
    google_project_service.apis,
  ]
}
