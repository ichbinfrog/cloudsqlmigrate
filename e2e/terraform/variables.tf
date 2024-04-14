variable "billing_account_id" {
  description = "The alphanumeric ID of the billing account this project belongs to"
  type        = string
  sensitive   = true
}

variable "enabled_github_infra" {
  description = "Whether or not to provision infrastructure for e2e tests"
  type        = bool
}

variable "github_username" {
  description = "Github username"
  type        = string
}

variable "github_repo" {
  description = "Github repo name"
  type        = string
}

variable "region" {
  description = "Region to provision resources in"
  type        = string
}

variable "test_matrix" {
  description = "List of CloudSQL database versions to be tested"
  type        = list(string)
}