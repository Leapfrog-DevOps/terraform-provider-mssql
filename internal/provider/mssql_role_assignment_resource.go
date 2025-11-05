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
	_ resource.Resource              = &roleAssignmentResource{}
	_ resource.ResourceWithConfigure = &roleAssignmentResource{}
)

// NewRoleResource a helper function to simplify the provider implementation.
func NewMssqlRoleAssignmentResource() resource.Resource {
	return &roleAssignmentResource{}
}

// maps to resource schema table
type roleAssignmentResourceModel struct {
	RoleName   types.String `tfsdk:"role_name"`
	MemberName types.String `tfsdk:"member_name"`
	Database   types.String `tfsdk:"database"`
	Id         types.String `tfsdk:"id"`
}

// roleResource is the resource implementation.
type roleAssignmentResource struct {
	client *sql.DB
}

// Metadata returns the resource type name.
func (r *roleAssignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_assignment"
}

// Schema defines the schema for the resource.
func (r *roleAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "MSSQL Role resource",
		Attributes: map[string]schema.Attribute{
			"role_name": schema.StringAttribute{
				MarkdownDescription: "Role name",
				Required:            true,
			},
			"member_name": schema.StringAttribute{
				MarkdownDescription: "User that is assigned the role",
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
func (r *roleAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data roleAssignmentResourceModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Create database in MSSQL
	role := data.RoleName.ValueString()
	member := data.MemberName.ValueString()
	database := data.Database.ValueString()

	// Create Statement
	createStmt := fmt.Sprintf(`
						-- Switch Context to the db
						Use [%s];
						-- Create custom role
						ALTER ROLE [%s] ADD MEMBER [%s];
	`, database, role, member)
	_, err := r.client.ExecContext(ctx, createStmt)
	if err != nil {
		resp.Diagnostics.AddError("Error assigning role", err.Error())
		return
	}
	data.Id = types.StringValue(data.RoleName.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save state
}

// Read refreshes the Terraform state with the latest data.
func (r *roleAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state roleAssignmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	query := fmt.Sprintf(`
		USE [%s];
		
		SELECT dp2.name AS member_name
		FROM sys.database_role_members drm
		JOIN sys.database_principals dp1 ON drm.role_principal_id = dp1.principal_id
		JOIN sys.database_principals dp2 ON drm.member_principal_id = dp2.principal_id
		WHERE dp1.name = @p1 AND dp2.name = @p2;
	`, state.Database.ValueString())
	row := r.client.QueryRowContext(ctx, query, state.RoleName.ValueString(), state.MemberName.ValueString())
	var name string
	err := row.Scan(&name)
	if err == sql.ErrNoRows {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError("Error reading role assignment", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *roleAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	return
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *roleAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data roleAssignmentResourceModel
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.ExecContext(ctx, fmt.Sprintf(`
		USE [%s];a
		ALTER ROLE [%s] DROP MEMBER [%s];`,
		data.Database.ValueString(),
		data.RoleName.ValueString(),
		data.MemberName.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error deleting assigned role", err.Error())
		return
	}

}

func (r *roleAssignmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
