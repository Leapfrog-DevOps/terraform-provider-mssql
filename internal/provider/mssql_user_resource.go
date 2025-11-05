// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &MssqlUserResource{}
var _ resource.ResourceWithImportState = &MssqlUserResource{}

func NewMssqlUserResource() resource.Resource {
	return &MssqlUserResource{}
}

type MssqlUserResource struct {
	client *sql.DB
}

type MssqlUserResourceModel struct {
	Name     types.String `tfsdk:"name"`
	Database types.String `tfsdk:"database"`
	Login    types.String `tfsdk:"login"`
	Id       types.String `tfsdk:"id"`
}

func (r *MssqlUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *MssqlUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "MSSQL User resource.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "User name.",
				Required:            true,
			},
			"database": schema.StringAttribute{
				MarkdownDescription: "Database name where the user will be created.",
				Required:            true,
			},
			"login": schema.StringAttribute{
				MarkdownDescription: "Login name to map the user to.",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "User identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *MssqlUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*sql.DB)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sql.DB, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *MssqlUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MssqlUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Switch to the target database
	_, err := r.client.ExecContext(ctx, fmt.Sprintf("USE [%s]", data.Database.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error switching to database", err.Error())
		return
	}

	// Create user in MSSQL
	var createStmt string
	if !data.Login.IsNull() && data.Login.ValueString() != "" {
		createStmt = fmt.Sprintf("USE [%s];CREATE USER [%s] FOR LOGIN [%s]", data.Database.ValueString(), data.Name.ValueString(), data.Login.ValueString())
	} else {
		createStmt = fmt.Sprintf("Use [%s];CREATE USER [%s] WITHOUT LOGIN", data.Database.ValueString(), data.Name.ValueString())
	}

	_, err = r.client.ExecContext(ctx, createStmt)
	if err != nil {
		resp.Diagnostics.AddError("Error creating user", err.Error())
		return
	}

	data.Id = types.StringValue(fmt.Sprintf("%s.%s", data.Database.ValueString(), data.Name.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MssqlUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MssqlUserResourceModel

	// Load the current state into `data`
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the combined SQL query (USE + SELECT)
	query := fmt.Sprintf(`
		USE [%s];
		SELECT name 
		FROM sys.database_principals 
		WHERE name = @p1 AND type = 'S';`, data.Database.ValueString())

	// Run the query
	row := r.client.QueryRowContext(ctx, query, data.Name.ValueString())

	// Read the result
	var name string
	err := row.Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			// User doesn't exist — remove from state
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Error reading user", err.Error())
		return
	}

	// Update state (no change needed if user exists)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MssqlUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MssqlUserResourceModel
	var state MssqlUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Switch to the target database
	_, err := r.client.ExecContext(ctx, fmt.Sprintf("USE [%s]", plan.Database.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error switching to database", err.Error())
		return
	}

	// If name changed, rename user
	if plan.Name.ValueString() != state.Name.ValueString() {
		_, err := r.client.ExecContext(ctx, fmt.Sprintf("ALTER USER [%s] WITH NAME = [%s]", state.Name.ValueString(), plan.Name.ValueString()))
		if err != nil {
			resp.Diagnostics.AddError("Error renaming user", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *MssqlUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MssqlUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Switch to the target database
	_, err := r.client.ExecContext(ctx, fmt.Sprintf("USE [%s]", data.Database.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error switching to database", err.Error())
		return
	}

	_, err = r.client.ExecContext(ctx, fmt.Sprintf("DROP USER [%s]", data.Name.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error deleting user", err.Error())
		return
	}
}

func (r *MssqlUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
