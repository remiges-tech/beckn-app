variable "project_id" {
  description = "GCP project ID to deploy Cloud Run/Artifact Registry/Secret Manager into."
  type        = string
  default     = "remiges-trade"
}

variable "remiges_ion_project_id" {
  description = "GCP project ID that owns the existing Cloud SQL instance, Memorystore instance, and VPC being reused."
  type        = string
  default     = "remiges-ion"
}

variable "region" {
  description = "GCP region for all resources. Must match the region the existing Cloud SQL/Redis/VPC live in for Shared VPC + Direct VPC Egress to work."
  type        = string
  default     = "asia-southeast2"
}

variable "image_tag" {
  description = "Tag applied to all images by scripts/build-and-push.sh (e.g. a git SHA). Terraform ignores image changes after first create — see lifecycle blocks on each Cloud Run resource."
  type        = string
  default     = "latest"
}

variable "artifact_registry_repo_id" {
  description = "Artifact Registry Docker repository ID (created in remiges-trade)."
  type        = string
  default     = "beckn-app"
}

# --- Existing remiges-ion resources being reused (not created here) ---

variable "sql_instance_name" {
  description = "Name of the existing Cloud SQL instance in remiges-ion to reuse."
  type        = string
  default     = "insurance-postgres"
}

variable "redis_instance_name" {
  description = "Name of the existing Memorystore instance in remiges-ion to reuse."
  type        = string
  default     = "insurance-redis"
}

variable "vpc_name" {
  description = "Name of the existing VPC network in remiges-ion to share via Shared VPC."
  type        = string
  default     = "insurance-app-vpc"
}

variable "subnet_name" {
  description = "Name of the existing subnetwork in remiges-ion (within vpc_name) to grant remiges-trade's Cloud Run service account network-user access to."
  type        = string
  default     = "insurance-app-subnet"
}

# --- New databases/user on the existing Cloud SQL instance ---

variable "db_name_bap" {
  type    = string
  default = "trade_bap"
}

variable "db_name_bpp" {
  type    = string
  default = "trade_bpp"
}

variable "db_user" {
  type    = string
  default = "trade"
}

# --- Beckn network identity ---
# Single source of truth per side: this value appears BOTH in context.bapId/
# bppId (every outbound message) AND as onix-bap/onix-bpp's own keyManager
# networkParticipant. These two uses MUST always match — onix signs a
# message using the keyset registered for whatever subscriber_id the message
# itself claims to be from, so any mismatch here fails signing with "keyset
# not found" (hit this for real: bap_id stayed at the old baptest1.remiges.tech
# identity while the onix keyManager was updated to a new one, and every
# outbound select/init/confirm broke). Previously two separate variables;
# consolidated to one after that incident so they can't drift again.

variable "bap_id" {
  type    = string
  default = "onix-bap-x7m2wmy7nq-et.a.run.app"
}

variable "bpp_id" {
  type    = string
  default = "onix-bpp-x7m2wmy7nq-et.a.run.app"
}

variable "network_id" {
  type    = string
  default = "ion.id/ion-launch"
}

variable "cds_discover_url" {
  type    = string
  default = "https://34.47.138.217.sslip.io/beckn/discover"
}

variable "cds_publish_url" {
  type    = string
  default = "https://fabric.nfh.global/beckn/catalog/publish"
}

# --- Secrets ---
# No variables here on purpose. Every secret value (Postgres password aside,
# which Terraform generates itself) is created as an empty Secret Manager
# container by Terraform, then seeded once with its real value via
# scripts/seed-secrets.sh — see terraform/README.md. This keeps plaintext
# secrets out of every Terraform-visible file, including .tfvars and state.
