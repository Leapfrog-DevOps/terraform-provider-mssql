terraform {
  required_providers {
    mssql = {
      source = "hashicorp.com/terrafarmers/mssql"
    }
  }
}

provider "mssql" {
  host        = "localhost"
  user        = "sa"
  password    = "YourPassword123!"
  port        = 1433
  default_db  = "master"
}

resource "mssql_database" "flow_db" {
  name = "flow_test_database"
}

resource "mssql_login" "flow_login" {
  name             = "flow_test_user"
  password         = "FlowTest123!"
  type             = "sql"
  default_database = "master"
}

resource "mssql_user" "flow_user_updated" {
  name     = "flow_user_updated"
  database = "master"
  login    = mssql_login.flow_login.name
}
