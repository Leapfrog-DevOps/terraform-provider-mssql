package provider

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	_ "github.com/microsoft/go-mssqldb"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &roleResource{}
	_ resource.ResourceWithConfigure = &roleResource{}
)

// NewRoleResource a helper function to simplify the provider implementation.
func NewMssqlRoleResource() resource.Resource {
	return &roleResource{}
}

// maps to resource schema table
type roleResourceModel struct {
	Name     types.String `tfsdk:"name"`
	Database types.String `tfsdk:"database"`
	Id       types.String `tfsdk:"id"`
}

// roleResource is the resource implementation.
type roleResource struct {
	client *sql.DB
}

// Metadata returns the resource type name.
func (r *roleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

// Schema defines the schema for the resource.
func (r *roleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "MSSQL Role resource",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Role name",
				Required:            true,
			},
			"database": schema.StringAttribute{
				MarkdownDescription: "Database name",
				Required:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Role identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data roleResourceModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Create database in MSSQL
	role := data.Name.ValueString()
	database := data.Database.ValueString()

	// Create Statement
	createStmt := fmt.Sprintf(`
						-- Switch Context to the db
						Use [%s];
						-- Create custom role
						CREATE ROLE [%s];
	`, database, role)
	_, err := r.client.ExecContext(ctx, createStmt)
	if err != nil {
		resp.Diagnostics.AddError("Error creating role", err.Error())
		return
	}
	data.Id = types.StringValue(data.Name.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save state
}

// Read refreshes the Terraform state with the latest data.
func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state roleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	query := fmt.Sprintf(`
			USE [%s];
			SELECT 1 FROM sys.database_principals WHERE type = 'R' AND name='%s';
	`, state.Database.ValueString(), state.Name.ValueString())
	row := r.client.QueryRowContext(ctx, query)
	var name string
	err := row.Scan(&name)
	if err == sql.ErrNoRows {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError("Error reading roles", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	return
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data roleResourceModel
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.ExecContext(ctx, fmt.Sprintf("USE [%s];DROP ROLE [%s]", data.Database.ValueString(), data.Name.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error deleting role", err.Error())
		return
	}

}

func (r *roleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
