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
  password="YourStrong!Passw0rd"
  
}
data "mssql_data" "example"{
  
}

resource "mssql_database" "appdb"{
  name="appdb"
  owner="sa" 
}



