# MSSQL Provider Complete Flow Test

## Prerequisites

### 1. Start SQL Server
```bash
docker run -e "ACCEPT_EULA=Y" -e "SA_PASSWORD=YourPassword123!" -p 1433:1433 --name sqlserver -d mcr.microsoft.com/mssql/server:2019-latest
sleep 15
```
2. Build Provider
```bash
make install
```

3. Configure Terraform
```bash
cat > ~/.terraformrc << 'EOL'
provider_installation {
  dev_overrides {
    "hashicorp.com/terrafarmers/mssql" = "/home/leapfrog/go/bin"
  }
  direct {}
}
EOL
```


Run Test
```
./complete_flow_test.sh
```

CREATE - Database, login, user

LOGIN - User authentication

CHECK - Database connection

UPDATE - User modification

LOGIN - Post-update authentication

DELETE - Resource cleanup

VERIFY - Deletion confirmation

Expected Output
```
🎉 COMPLETE FLOW TEST FINISHED SUCCESSFULLY!
🎯 MSSQL Terraform Provider: FULLY FUNCTIONAL!
```