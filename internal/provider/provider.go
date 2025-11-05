// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	_ "github.com/microsoft/go-mssqldb"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &mssqlProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &mssqlProvider{
			version: version,
		}
	}
}

type mssqlProviderModel struct {
	Host      types.String `tfsdk:"host"`
	User      types.String `tfsdk:"user"`
	Password  types.String `tfsdk:"password"`
	Port      types.Int32  `tfsdk:"port"`
	DefaultDb types.String `tfsdk:"default_db"`
}

// mssqlProvider is the provider implementation.
type mssqlProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	client  *sql.DB
	version string
}

// Metadata returns the provider type name.
func (p *mssqlProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mssql"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *mssqlProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"user": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"port": schema.Int32Attribute{
				Optional: true,
			},
			"default_db": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

// Configure prepares a mssql API client for data sources and resources.
func (p *mssqlProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config mssqlProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown mssql host",
			"The provider cannot create connection to the server as there is unknown configuration value from mssql provider host",
		)
	}
	if config.User.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown mssql user",
			"The provider cannot create connection to the server as there is unknown configuration value from mssql provider user",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown mssql password",
			"The provider cannot create connection to the server as there is unknown configuration value from mssql provider user",
		)
	}

	if config.DefaultDb.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown mssql default_db",
			"The provider cannot create connection to the server as there is unknown configuration value from mssql provider default_db",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values set in the environment variables
	// override with Terraform configuration if set

	host := os.Getenv("MSSQL_HOST")
	user := os.Getenv("MSSQL_USER")
	password := os.Getenv("MSSQL_PASSWORD")
	defaultDb := os.Getenv("MSSQL_DEFAULT_DB")
	var port int32 = 0

	if portStr := os.Getenv("MSSQL_PORT"); portStr != "" {
		parsedPort, err := strconv.Atoi(portStr)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid MSSQL_PORT Environment Variable",
				fmt.Sprintf("Expected MSSQL_PORT to be an integer, but got: %q. Error: %s", portStr, err),
			)
			return
		}
		port = int32(parsedPort)
	}
	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}
	if !config.User.IsNull() {
		user = config.User.ValueString()
	}
	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}
	if !config.Port.IsNull() {
		port = config.Port.ValueInt32()
	}
	if !config.DefaultDb.IsNull() {
		defaultDb = config.DefaultDb.ValueString()
	}
	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.
	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing mssql Host",
			"The provider cannot create mssql client as their is a missing or empty value for mssql host",
		)
	}

	if user == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("user"),
			"Missing mssql User",
			"The provider cannot create mssql client as their is a missing or empty value for mssql User",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing mssql Password",
			"The provider cannot create mssql client as their is a missing or empty value for mssql password",
		)
	}

	if defaultDb == "" {
		defaultDb = "master"
	}

	if port == 0 {
		port = 1433
	}

	connString := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=master",
		user,
		password,
		host,
		port,
	)

	client, err := sql.Open("sqlserver", connString)
	if err != nil {
		resp.Diagnostics.AddError("Failed to ping DB", err.Error())
		return
	}
	if err := client.PingContext(ctx); err != nil {
		resp.Diagnostics.AddError(
			"Unable to connecte to SQL Server",
			fmt.Sprintf("Ping failed: %s", err.Error()),
		)
	}
	resp.DataSourceData = client
	resp.ResourceData = client
}

// DataSources defines the data sources implemented in the provider.
func (p *mssqlProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewMssqlDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *mssqlProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewMssqlLoginResource,
		NewDatabaseResource,
		NewMssqlUserResource,
		NewMssqlRoleResource,
		NewMssqlRoleAssignmentResource,
	}
}
