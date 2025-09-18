terraform {
  required_providers {
    mssql = {
      source = "registry.terraform.io/terrafarmers/mssql"
    }
  }
}

provider "mssql" {
  host       = "localhost"
  user       = "sa"
  password   = "YourPassword123!"
  port       = 1433
  default_db = "master"
}

# Create a login first
resource "mssql_login" "example_login" {
  name             = "example_login"
  password         = "ExamplePassword123!"
  type             = "sql"
  default_database = "master"
}

# Create a database user mapped to a login
resource "mssql_user" "example" {
  name     = "example_user"
  database = "master"
  login    = mssql_login.example_login.name
}