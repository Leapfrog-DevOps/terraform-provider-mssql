// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &mssqlDataSource{}
var _ datasource.DataSourceWithConfigure = &mssqlDataSource{}

func NewMssqlDataSource() datasource.DataSource {
	return &mssqlDataSource{}
}

type mssqlDataSource struct {
	client *sql.DB
}

type mssqlDataSourceModel struct {
	Id      types.String `tfsdk:"id"`
	Version types.String `tfsdk:"version"`
}

func (d *mssqlDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data"
}

func (d *mssqlDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "MSSQL server information data source.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier.",
				Computed:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "MSSQL server version.",
				Computed:            true,
			},
		},
	}
}

func (d *mssqlDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*sql.DB)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *sql.DB, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *mssqlDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data mssqlDataSourceModel

	// Get SQL Server version
	row := d.client.QueryRowContext(ctx, "SELECT @@VERSION")
	var version string
	err := row.Scan(&version)
	if err != nil {
		resp.Diagnostics.AddError("Error reading SQL Server version", err.Error())
		return
	}

	data.Id = types.StringValue("mssql_server")
	data.Version = types.StringValue(version)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
