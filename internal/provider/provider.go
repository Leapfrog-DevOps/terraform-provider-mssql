package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	Host     types.String `tfsdk:"host"`
	User     types.String `tfsdk:"user"`
	Password types.String `tfsdk:"password"`
}

// mssqlProvider is the provider implementation.
type mssqlProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
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
				Required:  true,
				Sensitive: true,
			},
			"user": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"password": schema.StringAttribute{
				Required:  true,
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

}

// DataSources defines the data sources implemented in the provider.
func (p *mssqlProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

// Resources defines the resources implemented in the provider.
func (p *mssqlProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}
