resource "yandex_lockbox_secret" "db_credentials" {
  name      = "db-credentials"
  folder_id = var.folder_id
}

resource "yandex_lockbox_secret_version" "db_credentials_v1" {
  secret_id = yandex_lockbox_secret.db_credentials.id
  entries {
    key        = "postgres_password"
    text_value = var.postgres_password
  }
  entries {
    key        = "postgres_user"
    text_value = "app_user"
  }
}