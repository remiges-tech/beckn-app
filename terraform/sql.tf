# Reuses remiges-ion's existing Cloud SQL instance — does NOT create a new
# instance. Only new databases + a new user are created on it, namespaced to
# this app so they don't collide with the insurance app's insurance_bap/
# insurance_bpp databases already on the same instance.

data "google_sql_database_instance" "existing" {
  provider = google.remiges_ion
  name     = var.sql_instance_name
  project  = var.remiges_ion_project_id
}

resource "google_sql_database" "trade_bap" {
  provider = google.remiges_ion
  name     = var.db_name_bap
  project  = var.remiges_ion_project_id
  instance = data.google_sql_database_instance.existing.name
}

resource "google_sql_database" "trade_bpp" {
  provider = google.remiges_ion
  name     = var.db_name_bpp
  project  = var.remiges_ion_project_id
  instance = data.google_sql_database_instance.existing.name
}

resource "random_password" "postgres" {
  length  = 32
  special = false
}

resource "google_sql_user" "trade" {
  provider = google.remiges_ion
  name     = var.db_user
  project  = var.remiges_ion_project_id
  instance = data.google_sql_database_instance.existing.name
  password = random_password.postgres.result
}
