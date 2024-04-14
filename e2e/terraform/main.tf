// Base structure
resource "random_bytes" "self" {
  length = 2
}

resource "google_project" "self" {
  name            = "CloudSQL Migrate E2E tests"
  project_id      = "prj-migrate-${random_bytes.self.hex}"
  billing_account = var.billing_account_id
}

resource "google_project_service" "self" {
  for_each = toset([
    "sqladmin.googleapis.com",
    "iamcredentials.googleapis.com",
  ])
  project                    = google_project.self.id
  service                    = each.key
  disable_dependent_services = false
}

resource "random_bytes" "cloudsql" {
  length = 2
}

resource "google_sql_database_instance" "source" {
  for_each = toset(var.test_matrix)
  project  = google_project.self.project_id
  name     = "sql-e2e-src-${lower(replace(each.key, "_", "-"))}-${random_bytes.cloudsql.hex}"
  region   = var.region

  database_version = each.key
  settings {
    tier      = "db-f1-micro"
    disk_type = "PD_HDD"
    disk_size = 10
  }
}

resource "google_sql_database_instance" "target" {
  for_each = toset(var.test_matrix)
  project  = google_project.self.project_id
  name     = "sql-e2e-dst-${lower(replace(each.key, "_", "-"))}-${random_bytes.cloudsql.hex}"
  region   = var.region

  database_version = each.key
  settings {
    tier      = "db-f1-micro"
    disk_type = "PD_HDD"
    disk_size = 10
  }
}