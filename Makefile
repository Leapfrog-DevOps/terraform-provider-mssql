GOPATH := $(shell go env GOPATH)

terraformrc:
	@echo 'Creating terraform.rc in home directory'
	@touch ~/.terraformrc
	@echo 'provider_installation {' > ~/.terraformrc
	@echo '  dev_overrides {' >> ~/.terraformrc
	@echo '    "hashicorp.com/terrafarmers/mssql" = "$(GOPATH)/bin"' >> ~/.terraformrc
	@echo '  }' >> ~/.terraformrc
	@echo '  direct {}' >> ~/.terraformrc
	@echo '}' >> ~/.terraformrc
	@echo "Created: .terraformrc file"

