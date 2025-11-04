terraform {
  required_providers {
    mssql = {
      source = "hashicorp.com/terrafarmers/mssql"
    }
  }
}

provider "mssql" {
  host = "localhost"
  user= "sa"
  password = "YourStrong!Passw0rd"
}
data "mssql_data" "example" {

}

resource "mssql_login" "login_test123" {
  name             = "testuser1123"
  password         = "SuperSecretPassword123!"
  type             = "sql"    # options: "sql" or "windows"
}

resource "mssql_database" "database_test" {
  name = "testdb"
}

resource "mssql_user" "userexample" {
  name     = "example_user"
  database = mssql_database.database_test.name
  login    = mssql_login.login_test123.name
}




resource "mssql_role" "roletest"{
  name = "app_user1234"
  database = mssql_database.database_test.name
}

resource "mssql_role_assignment" "assignmenttest"{

  member_name=mssql_user.userexample.name
  database=mssql_database.database_test.name
  role_name=mssql_role.roletest.name
}







