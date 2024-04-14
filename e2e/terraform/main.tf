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

// E2E Tests
resource "google_service_account" "e2e" {
  count      = var.enabled_github_infra ? 1 : 0
  project    = google_project.self.project_id
  account_id = "sa-e2e"
}

locals {
  // Technically, the least privilege role would only be a subset of the CloudSQL Viewer Role with
  // "cloudsql.backupRuns.list", "cloudsql.backupRuns.get", "cloudsql.backupRuns.create", "cloudsql.backupRuns.restoreBackup",
  // but I'm lazy :)
  cloudsql_permissions = [
    "roles/cloudsql.admin"
  ]
}

resource "google_project_iam_member" "e2e" {
  for_each = { for k in local.cloudsql_permissions : k => k if var.enabled_github_infra }
  project  = google_project.self.project_id
  member   = "serviceAccount:${google_service_account.e2e[0].email}"
  role     = each.key
}

resource "google_iam_workload_identity_pool" "github" {
  count                     = var.enabled_github_infra ? 1 : 0
  project                   = google_project_service.self["iamcredentials.googleapis.com"].project
  workload_identity_pool_id = "github-pool"
  display_name              = "Github E2E Tests pipeline"
}

resource "google_iam_workload_identity_pool_provider" "github" {
  count   = var.enabled_github_infra ? 1 : 0
  project = google_project.self.project_id

  workload_identity_pool_id          = google_iam_workload_identity_pool.github[0].workload_identity_pool_id
  workload_identity_pool_provider_id = "github-provider"
  description                        = "OIDC identity pool provider for e2e tests"
  disabled                           = false

  attribute_mapping = {
    "google.subject"             = "assertion.sub"
    "attribute.actor"            = "assertion.actor"
    "attribute.repository_owner" = "assertion.repository_owner"
    "attribute.repository"       = "assertion.repository"
  }

  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }
}

resource "google_service_account_iam_member" "identity_federation_principalset" {
  count              = var.enabled_github_infra ? 1 : 0
  service_account_id = google_service_account.e2e[0].name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github[0].name}/attribute.repository/${var.github_username}/${var.github_repo}"

  depends_on = [
    google_iam_workload_identity_pool_provider.github[0]
  ]
}