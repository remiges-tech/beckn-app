# Reuses remiges-ion's existing Memorystore instance — data source only, no
# new instance created. onix-bap and onix-bpp will both point their
# REDIS_ADDR at this same instance, same as the insurance app's own onix
# adapters already do; cache keys are namespaced by transaction/message UUID
# so collision risk across the two apps sharing one keyspace is low, but not
# zero.

data "google_redis_instance" "existing" {
  provider = google.remiges_ion
  name     = var.redis_instance_name
  region   = var.region
  project  = var.remiges_ion_project_id
}
