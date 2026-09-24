package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/withlogicco/terraform-provider-hri/internal/robot"
)

type HRIProvider struct{ version string }
type ProviderModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	BaseURL  types.String `tfsdk:"base_url"`
}

func New(version string) provider.Provider {
	return &HRIProvider{version: version}
}

func (p *HRIProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "hri"
	resp.Version = p.version
}

func (p *HRIProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The HRI provider reads Hetzner dedicated server inventory from the Robot Webservice.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Robot webservice username. Defaults to HETZNER_ROBOT_USER.",
			},
			"password": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Robot webservice password. Defaults to HETZNER_ROBOT_PASSWORD.",
			},
			"base_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Robot API base URL.",
			},
		}}
}

func (p *HRIProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Username.IsNull() || data.Username.IsUnknown() {
		data.Username = types.StringValue(os.Getenv("HETZNER_ROBOT_USER"))
	}

	if data.Password.IsNull() || data.Password.IsUnknown() {
		data.Password = types.StringValue(os.Getenv("HETZNER_ROBOT_PASSWORD"))
	}

	if data.Username.ValueString() == "" {
		resp.Diagnostics.AddError("Missing Robot username", "Set username or HETZNER_ROBOT_USER.")
	}

	if data.Password.ValueString() == "" {
		resp.Diagnostics.AddError("Missing Robot password", "Set password or HETZNER_ROBOT_PASSWORD.")
	}

	if resp.Diagnostics.HasError() {
		return
	}

	base := data.BaseURL.ValueString()
	if base == "" {
		base = "https://robot-ws.your-server.de"
	}

	client, err := robot.NewClient(data.Username.ValueString(), data.Password.ValueString(), base)
	if err != nil {
		resp.Diagnostics.AddError("Invalid base_url", err.Error())
		return
	}

	index := robot.NewIndex(client)
	resp.ResourceData = index
	resp.DataSourceData = index
}

func (p *HRIProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{NewServerResource}
}

func (p *HRIProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{NewServerDataSource, NewServersDataSource}
}
