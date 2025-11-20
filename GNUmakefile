# ---------------------------
# Settings
# ---------------------------

PROVIDER_NAME := mssql
NAMESPACE     := terrafarmers
VERSION       := 0.1.0

TERRAFORMRC   := $(HOME)/.terraformrc

# Provider binary lives here after `go install`
PLUGIN_DIR    := $(HOME)/go/bin

SQL_CONTAINER := local-mssql
SA_PASSWORD   := YourStrong!Passw0rd

TF_TEST_DIR   := ./test

# ---------------------------
# DEFAULT target: create terraformrc
# ---------------------------
default: terraformrc

# ---------------------------
# Run full test sequence
# ---------------------------
test: start-mssql install-provider clean-tfstate tf-apply

# ---------------------------
# Create terraformrc
# ---------------------------
terraformrc:
	@echo "==> Creating $(TERRAFORMRC)"
	@mkdir -p $$(dirname $(TERRAFORMRC))
	@printf "%s\n" "\
provider_installation {\n\
  dev_overrides {\n\
    \"$(NAMESPACE)/$(PROVIDER_NAME)\" = \"$(PLUGIN_DIR)\"\n\
  }\n\
  direct {}\n\
}" > $(TERRAFORMRC)
	@echo "terraformrc created. Terraform will load provider from: $(PLUGIN_DIR)"

# ---------------------------
# Start SQL Server
# ---------------------------
start-mssql:
	@echo "==> Starting SQL Server Docker container"
	-@docker rm -f $(SQL_CONTAINER) >/dev/null 2>&1 || true
	docker run -e 'ACCEPT_EULA=Y' \
	           -e 'SA_PASSWORD=$(SA_PASSWORD)' \
	           -p 1433:1433 \
	           --name $(SQL_CONTAINER) \
	           -d mcr.microsoft.com/mssql/server:2019-latest

	@echo "Waiting for SQL Server to become ready..."
	@until docker logs $(SQL_CONTAINER) 2>&1 | grep -q "SQL Server is now ready"; do \
		printf "."; \
		sleep 2; \
	done
	@echo "\nSQL Server ready!"

# ---------------------------
# Build provider into $HOME/go/bin
# ---------------------------
install-provider:
	@echo "==> Building provider (go install)"
	go install .
	@echo "Provider installed at: $(HOME)/go/bin/terraform-provider-$(PROVIDER_NAME)"

# ---------------------------
# Clean TF state in test/
# ---------------------------
clean-tfstate:
	@echo "==> Removing Terraform state files"
	rm -f  $(TF_TEST_DIR)/terraform.tfstate
	rm -f  $(TF_TEST_DIR)/terraform.tfstate.backup
	rm -f  $(TF_TEST_DIR)/.terraform.lock.hcl
	rm -rf $(TF_TEST_DIR)/.terraform
	@echo "TF state cleaned."

# ---------------------------
# Run terraform init + apply
# ---------------------------
tf-apply:
	@echo "==> Running Terraform using test/main.tf"
	cd $(TF_TEST_DIR) && terraform apply -auto-approve

# ---------------------------
# Cleanup
# ---------------------------
clean:
	-@docker rm -f $(SQL_CONTAINER) >/dev/null 2>&1 || true
	@echo "Clean complete."
