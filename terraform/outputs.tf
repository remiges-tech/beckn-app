output "bap_frontend_url" {
  value       = google_cloud_run_v2_service.bap_frontend.uri
  description = "Public URL — open this to use the BAP (buyer) app."
}

output "bpp_frontend_url" {
  value       = google_cloud_run_v2_service.bpp_frontend.uri
  description = "Public URL — open this to use the BPP (seller/admin) app."
}

output "bap_url" {
  value = google_cloud_run_v2_service.bap.uri
}

output "bpp_url" {
  value = google_cloud_run_v2_service.bpp.uri
}

output "onix_bap_url" {
  value = google_cloud_run_v2_service.onix_bap.uri
}

output "onix_bpp_url" {
  value = google_cloud_run_v2_service.onix_bpp.uri
}

output "artifact_registry_repo" {
  value       = "${var.region}-docker.pkg.dev/${var.project_id}/${var.artifact_registry_repo_id}"
  description = "Push images here — see scripts/build-and-push.sh."
}

output "cloud_sql_connection_name" {
  value = data.google_sql_database_instance.existing.connection_name
}

output "redis_host" {
  value = data.google_redis_instance.existing.host
}
