package provider

import (
	"context"
	"regexp"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/withlogicco/terraform-provider-hri/internal/robot"
)

type ServersDataSource struct{ index *robot.Index }
type ServersDataSourceModel struct {
	DC               types.String `tfsdk:"dc"`
	Product          types.String `tfsdk:"product"`
	NameRegex        types.String `tfsdk:"name_regex"`
	IncludeCancelled types.Bool   `tfsdk:"include_cancelled"`
	ServerNumbers    types.Set    `tfsdk:"server_numbers"`
	ID               types.String `tfsdk:"id"`
}

func NewServersDataSource() datasource.DataSource {
	return &ServersDataSource{}
}

func (d *ServersDataSource) Metadata(
	_ context.Context,
	r datasource.MetadataRequest,
	p *datasource.MetadataResponse,
) {
	p.TypeName = r.ProviderTypeName + "_servers"
}

func (d *ServersDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	r *datasource.SchemaResponse,
) {
	attributes := dataSourceSchemaAttributes{
		"dc": optionalStringDataSourceAttribute(
			"Filter by Robot data center.",
		),
		"product": optionalStringDataSourceAttribute(
			"Filter by product name.",
		),
		"name_regex": optionalStringDataSourceAttribute(
			"Filter server names using a Go regular expression.",
		),
		"include_cancelled": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "Include cancelled servers. Defaults to false.",
		},
		"server_numbers": schema.SetAttribute{
			Computed:            true,
			ElementType:         types.Int64Type,
			MarkdownDescription: "Server numbers matching the filters.",
		},
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Inventory snapshot identifier.",
		},
	}

	r.Schema = schema.Schema{
		MarkdownDescription: "Reads the account inventory and returns server numbers, with optional filters.",
		Attributes:          attributes,
	}
}
func (d *ServersDataSource) Configure(
	_ context.Context,
	r datasource.ConfigureRequest,
	p *datasource.ConfigureResponse,
) {
	if r.ProviderData == nil {
		return
	}

	i, ok := r.ProviderData.(*robot.Index)
	if !ok {
		p.Diagnostics.AddError("Unexpected Provider Data", "Expected Robot inventory index.")
		return
	}

	d.index = i
}

func (d *ServersDataSource) Read(
	ctx context.Context,
	r datasource.ReadRequest,
	p *datasource.ReadResponse,
) {
	var m ServersDataSourceModel

	p.Diagnostics.Append(r.Config.Get(ctx, &m)...)
	if p.Diagnostics.HasError() {
		return
	}

	all, e := d.index.All()
	if e != nil {
		p.Diagnostics.AddError("Robot inventory failed", e.Error())
		return
	}

	var re *regexp.Regexp
	if !m.NameRegex.IsNull() && !m.NameRegex.IsUnknown() {
		re, e = regexp.Compile(m.NameRegex.ValueString())
		if e != nil {
			p.Diagnostics.AddAttributeError(path.Root("name_regex"), "Invalid regular expression", e.Error())
			return
		}
	}

	nums := []int64{}
	for n, s := range all {
		if !m.DC.IsNull() && s.DC != m.DC.ValueString() {
			continue
		}
		if !m.Product.IsNull() && s.Product != m.Product.ValueString() {
			continue
		}
		if s.Cancelled && (m.IncludeCancelled.IsNull() || !m.IncludeCancelled.ValueBool()) {
			continue
		}
		if re != nil && !re.MatchString(s.ServerName) {
			continue
		}
		nums = append(nums, int64(n))
	}

	sort.Slice(nums, func(i, j int) bool { return nums[i] < nums[j] })

	var ds diag.Diagnostics

	m.ServerNumbers, ds = types.SetValueFrom(ctx, types.Int64Type, nums)
	p.Diagnostics.Append(ds...)
	if p.Diagnostics.HasError() {
		return
	}

	m.ID = types.StringValue("inventory")
	p.Diagnostics.Append(p.State.Set(ctx, &m)...)
}
