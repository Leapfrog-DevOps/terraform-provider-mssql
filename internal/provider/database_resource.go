package provider

import (
	"context"
	"database/sql"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	_ "github.com/microsoft/go-mssqldb"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &databaseResource{}
)

// NewOrderResource is a helper function to simplify the provider implementation.
func NewDatabaseResource() resource.Resource {
	return &databaseResource{}
}

// orderResource is the resource implementation.
type databaseResource struct {
	client *sql.DB
}

// Metadata returns the resource type name.
func (r *databaseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

// Schema defines the schema for the resource.
func (r *databaseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{}
}

// Create creates the resource and sets the initial Terraform state.
func (r *databaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
}

// Read refreshes the Terraform state with the latest data.
func (r *databaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *databaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *databaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
