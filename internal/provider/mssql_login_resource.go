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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &MssqlLoginResource{}
var _ resource.ResourceWithImportState = &MssqlLoginResource{}

func NewMssqlLoginResource() resource.Resource {
	return &MssqlLoginResource{}
}

type MssqlLoginResource struct {
	client *sql.DB
}

type MssqlLoginResourceModel struct {
	Name            types.String `tfsdk:"name"`
	Password        types.String `tfsdk:"password"`
	Type            types.String `tfsdk:"type"`
	DefaultDatabase types.String `tfsdk:"default_database"`
	Id              types.String `tfsdk:"id"`
}

func (r *MssqlLoginResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_login"
}

func (r *MssqlLoginResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "MSSQL Login resource.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Login name.",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Login password.",
				Required:            true,
				Sensitive:           true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Login type: sql or windows.",
				Required:            true,
			},
			"default_database": schema.StringAttribute{
				MarkdownDescription: "Default database. Defaults to master.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("master"),
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Login identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *MssqlLoginResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MssqlLoginResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MssqlLoginResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...) // Read plan
	if resp.Diagnostics.HasError() {
		return
	}
	// Create login in MSSQL
	loginType := data.Type.ValueString()
	var createStmt string
	if loginType == "sql" {
		createStmt = fmt.Sprintf("CREATE LOGIN [%s] WITH PASSWORD = '%s', DEFAULT_DATABASE = [%s]", data.Name.ValueString(), data.Password.ValueString(), data.DefaultDatabase.ValueString())
	} else if loginType == "windows" {
		createStmt = fmt.Sprintf("CREATE LOGIN [%s] FROM WINDOWS WITH DEFAULT_DATABASE = [%s]", data.Name.ValueString(), data.DefaultDatabase.ValueString())
	} else {
		resp.Diagnostics.AddError("Invalid login type", "Type must be 'sql' or 'windows'.")
		return
	}
	_, err := r.client.ExecContext(ctx, createStmt)
	if err != nil {
		resp.Diagnostics.AddError("Error creating login", err.Error())
		return
	}
	data.Id = types.StringValue(data.Name.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save state
}

func (r *MssqlLoginResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MssqlLoginResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...) // Read state
	if resp.Diagnostics.HasError() {
		return
	}
	// Check if login exists
	row := r.client.QueryRowContext(ctx, "SELECT name FROM sys.server_principals WHERE name = @p1", data.Name.ValueString())
	var name string
	err := row.Scan(&name)
	if err == sql.ErrNoRows {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError("Error reading login", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save state
}

func (r *MssqlLoginResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MssqlLoginResourceModel
	var state MssqlLoginResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)   // Read plan
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...) // Read state
	if resp.Diagnostics.HasError() {
		return
	}
	// If name changed, rename login
	if plan.Name.ValueString() != state.Name.ValueString() {
		_, err := r.client.ExecContext(ctx, fmt.Sprintf("ALTER LOGIN [%s] WITH NAME = [%s]", state.Name.ValueString(), plan.Name.ValueString()))
		if err != nil {
			resp.Diagnostics.AddError("Error renaming login", err.Error())
			return
		}
		state.Name = plan.Name
	}
	// Update password and default_database if type is sql
	if plan.Type.ValueString() == "sql" {
		_, err := r.client.ExecContext(ctx, fmt.Sprintf("ALTER LOGIN [%s] WITH PASSWORD = '%s', DEFAULT_DATABASE = [%s]", plan.Name.ValueString(), plan.Password.ValueString(), plan.DefaultDatabase.ValueString()))
		if err != nil {
			resp.Diagnostics.AddError("Error updating login", err.Error())
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...) // Save state
}

func (r *MssqlLoginResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MssqlLoginResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...) // Read state
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.ExecContext(ctx, fmt.Sprintf("DROP LOGIN [%s]", data.Name.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error deleting login", err.Error())
		return
	}
}

func (r *MssqlLoginResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
