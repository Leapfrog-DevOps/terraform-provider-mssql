# mssql_user (Resource)

Manages an MSSQL database user.

## Example Usage

```terraform
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
```

## Schema

### Required

- `name` (String) User name.
- `database` (String) Database name where the user will be created.

### Optional

- `login` (String) Login name to map the user to.

### Read-Only

- `id` (String) User identifier.

## Import

Import is supported using the following syntax:

```shell
terraform import mssql_user.example database.username
```