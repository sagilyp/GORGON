variable "yc_sa_key_path" {
  type        = string
  description = "Path to the Yandex.Cloud service account key file (JSON)."
}

variable "cloud_id" {
  type        = string
  description = "Yandex Cloud ID"
}

variable "folder_id" {
  type        = string
  description = "Yandex Cloud Folder ID"
}

variable "zone" {
  type        = string
  description = "(Optional) - Availability zone"
  default     = "ru-central1-d"
}

variable "ssh_key_public_path" {
  type        = string
  description = "Path to the public SSH key file for VM access (e.g., ~/.ssh/id_rsa.pub)."
}

variable "postgres_password" {
  type        = string
  description = "Password for the PostgreSQL database user. Will be stored in Lockbox."
  sensitive   = true
}