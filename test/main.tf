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