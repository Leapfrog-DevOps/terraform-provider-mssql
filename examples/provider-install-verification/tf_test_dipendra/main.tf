terraform{
  required_providers {
  mssql={
       source = "hashicorp.com/terrafarmers/mssql"
    }
  }
}

provider "mssql" {
  user="admin"  
}
data "mssql_data" "example"{
  
}

resource "mssql_login" "login_test" {
  name             = "testuser"
  password         = "SuperSecretPassword123!"
  type             = "sql"     # options: "sql" or "windows"
  default_database = "master"  # Optional, defaults to master
}
resource "mssql_database" "database_test"{
  name="testdb"
}



