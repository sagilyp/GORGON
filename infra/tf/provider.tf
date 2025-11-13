terraform {
  required_providers {
    yandex = {
      source = "yandex-cloud/yandex"
    }
  }
  required_version = ">= 1.00"
}

provider "yandex" {
  zone = var.zone
  service_account_key_file = var.yc_sa_key_path
  folder_id = var.folder_id
  cloud_id = var.cloud_id
}

