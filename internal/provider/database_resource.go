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
	Owner              types.String `tfsdk:"owner"`
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
			"owner": schema.StringAttribute{
				MarkdownDescription: "Database owner",
				Required:            true,
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
	owner := data.Owner.ValueString()

	// Check if owner exists
	var userExists bool
	checkOwnrStmt := "SELECT COUNT(*) FROM sys.syslogins WHERE name = @owner"
	err := r.client.QueryRowContext(ctx, checkOwnrStmt, sql.Named("owner", owner)).Scan(&userExists)
	if err != nil {
		resp.Diagnostics.AddError("Error checking for owner", fmt.Sprintf("Unable to check if '%s' exists: %s", owner, err.Error()))
		return
	}
	if !userExists {
		resp.Diagnostics.AddError("User does not exist", fmt.Sprintf("The '%s' user is required to create the database and assign ownership. Please ensure the '%s' user exists before proceeding.", owner, owner))
		return
	}
	// Create Statement
	createStmt := fmt.Sprintf(`
							CREATE DATABASE %s 
							COLLATE %s
							WITH COMPATIBILITY_LEVEL=%d ;
	`, name, collation, compatibilityLevel)
	_, err = r.client.ExecContext(ctx, createStmt)
	if err != nil {
		resp.Diagnostics.AddError("Error creating database", err.Error())
		return
	}
	// SQL statement to set the owner
	setOwnerStmt := fmt.Sprintf("ALTER AUTHORIZATION ON DATABASE::%s TO %s;", name, owner)
	_, err = r.client.ExecContext(ctx, setOwnerStmt)
	if err != nil {
		resp.Diagnostics.AddError("Error creating database", err.Error())
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
	var data databaseResourceModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// only owner can be changed
	_, err := r.client.ExecContext(ctx, fmt.Sprintf("ALTER AUTHORIZATION on DATABASE::[%s] TO %s", data.Name.ValueString(), data.Owner.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error updating database", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...) // Save state

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
