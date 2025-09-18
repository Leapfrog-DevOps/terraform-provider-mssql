terraform {
  required_providers {
    mssql = {
      source = "hashicorp.com/terrafarmers/mssql"
    }
  }
}

provider "mssql" {
  user = "admin"
}
data "mssql_data" "example" {

}

resource "mssql_login" "login_test123" {
  name             = "testuser1123"
  password         = "SuperSecretPassword123!"
  type             = "sql"    # options: "sql" or "windows"
  default_database = "master" # Optional, defaults to master
}
resource "mssql_database" "database_test" {
  name = "testdb"
}


resource "mssql_role" "roletest"{
  name = "app_user1234"
  database = mssql_database.database_test.name
}

resource "mssql_role_assignment" "assignmenttest"{

  member_name="cena123"
  database=mssql_database.database_test.name
  role_name=mssql_role.roletest.name
}







