terraform{
  required_providers {
  mssql={
       source = "hashicorp.com/terrafarmers/mssql"
    }
  }
}

provider "mssql" {
  host="localhost"
  user="sa"
  password="Coredev0ps"
  
}
data "mssql_data" "example"{
  
}

resource "mssql_database" "appdb"{
  name="appdbasd"
  owner="sa" 
}

resource "mssql_login" "app_login" {
  name             = "app_user1"
  password         = "SuperSecretPassword123!"
  type             = "sql"     # options: "sql" or "windows"
  default_database = "master"  # Optional, defaults to master
}


