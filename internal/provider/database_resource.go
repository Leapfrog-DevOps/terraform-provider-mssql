// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	_ "github.com/microsoft/go-mssqldb"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &databaseResource{}
	_ resource.ResourceWithConfigure = &databaseResource{}
)

// NewDatabaseResource a helper function to simplify the provider implementation.
func NewDatabaseResource() resource.Resource {
	return &databaseResource{}
}

// maps to resource schema table
type databaseResourceModel struct {
	Name               types.String `tfsdk:"name"`
	Collation          types.String `tfsdk:"collation"`
	CompatibilityLevel types.Int32  `tfsdk:"compatibility_level"`
	Id                 types.String `tfsdk:"id"`
}

// databaseResource is the resource implementation.
type databaseResource struct {
	client *sql.DB
}

// Metadata returns the resource type name.
func (r *databaseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

// Schema defines the schema for the resource.
func (r *databaseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "MSSQL Database resource",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Database name",
				Required:            true,
			},
			"collation": schema.StringAttribute{
				MarkdownDescription: "Database collation",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("SQL_Latin1_General_CP1_CI_AS"),
			},
			"compatibility_level": schema.Int32Attribute{
				MarkdownDescription: "Database compatibility level",
				Optional:            true,
				Computed:            true,
				Default:             int32default.StaticInt32(150),
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Database identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *databaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data databaseResourceModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Create database in MSSQL
	name := data.Name.ValueString()
	collation := data.Collation.ValueString()
	compatibilityLevel := data.CompatibilityLevel.ValueInt32()

	// Create Statement
	createStmt := fmt.Sprintf(`
							CREATE DATABASE %s 
							COLLATE %s
							WITH COMPATIBILITY_LEVEL=%d ;
	`, name, collation, compatibilityLevel)
	_, err := r.client.ExecContext(ctx, createStmt)
	if err != nil {
		resp.Diagnostics.AddError("Error creating database", err.Error())
		return
	}
	data.Id = types.StringValue(data.Name.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save state
}

// Read refreshes the Terraform state with the latest data.
func (r *databaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state databaseResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	row := r.client.QueryRowContext(ctx, "SELECT name FROM sys.databases where name=@db", sql.Named("db", state.Name.ValueString()))
	var name string
	err := row.Scan(&name)
	if err == sql.ErrNoRows {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError("Error reading database", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *databaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan databaseResourceModel
	var state databaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)   // Read plan
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...) // Read state
	if resp.Diagnostics.HasError() {
		return
	}
	// If name changed, rename database
	if plan.Name.ValueString() != state.Name.ValueString() {
		_, err := r.client.ExecContext(ctx, fmt.Sprintf("ALTER DATABASE [%s] MODIFY NAME = [%s]", state.Name.ValueString(), plan.Name.ValueString()))
		if err != nil {
			resp.Diagnostics.AddError("Error renaming database", err.Error())
			return
		}
		state.Name = plan.Name
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...) // Save state

}

// Delete deletes the resource and removes the Terraform state on success.
func (r *databaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data databaseResourceModel
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.ExecContext(ctx, fmt.Sprintf("DROP DATABASE [%s]", data.Name.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error deleting database", err.Error())
		return
	}

}

func (r *databaseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*sql.DB)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sql.DB, got :%T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}
