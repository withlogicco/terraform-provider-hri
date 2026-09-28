package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/withlogicco/terraform-provider-hri/internal/robot"
)

type ServerDataSource struct{ index *robot.Index }
type ServerDataSourceModel struct {
	ServerNumber  types.Int64  `tfsdk:"server_number"`
	ServerName    types.String `tfsdk:"server_name"`
	ServerIP      types.String `tfsdk:"server_ip"`
	ServerIPv6Net types.String `tfsdk:"server_ipv6_net"`
	Product       types.String `tfsdk:"product"`
	DC            types.String `tfsdk:"dc"`
	Traffic       types.String `tfsdk:"traffic"`
	Status        types.String `tfsdk:"status"`
	Cancelled     types.Bool   `tfsdk:"cancelled"`
	PaidUntil     types.String `tfsdk:"paid_until"`
	IPs           types.List   `tfsdk:"ips"`
	Subnets       types.List   `tfsdk:"subnets"`
}

func NewServerDataSource() datasource.DataSource {
	return &ServerDataSource{}
}

func (d *ServerDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (d *ServerDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	attributes := computedServerDataSourceAttributes()
	attributes["server_number"] = schema.Int64Attribute{
		Required:            true,
		MarkdownDescription: "Robot server number to read.",
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads one dedicated server from the Robot account inventory.",
		Attributes:          attributes,
	}
}

func (d *ServerDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	i, ok := req.ProviderData.(*robot.Index)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected Robot inventory index.")
		return
	}

	d.index = i
}

func (d *ServerDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var m ServerDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, ok, e := d.index.Get(int(m.ServerNumber.ValueInt64()))
	if e != nil {
		resp.Diagnostics.AddError("Robot inventory failed", e.Error())
		return
	}

	if !ok {
		resp.Diagnostics.AddAttributeError(path.Root("server_number"), "Server not found", fmt.Sprintf("Robot does not report server %d.", m.ServerNumber.ValueInt64()))
		return
	}

	x := ServerModel{ServerNumber: m.ServerNumber}
	setServer(ctx, &x, s, &resp.Diagnostics)
	m.ServerName = x.ServerName
	m.ServerIP = x.ServerIP
	m.ServerIPv6Net = x.ServerIPv6Net
	m.Product = x.Product
	m.DC = x.DC
	m.Traffic = x.Traffic
	m.Status = x.Status
	m.Cancelled = x.Cancelled
	m.PaidUntil = x.PaidUntil
	m.IPs = x.IPs
	m.Subnets = x.Subnets
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}
