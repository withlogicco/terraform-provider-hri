package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/withlogicco/terraform-provider-hri/internal/robot"
)

type ServerResource struct{ index *robot.Index }
type ServerModel struct {
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

func NewServerResource() resource.Resource {
	return &ServerResource{}
}

func (r *ServerResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (r *ServerResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	attributes := computedServerResourceAttributes()
	attributes["server_number"] = schema.Int64Attribute{
		Required:            true,
		MarkdownDescription: "Robot server number. Changing this value replaces the resource.",
		PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Inventories an existing dedicated server into Terraform state. This resource does not create or modify Robot servers.",
		Attributes:          attributes,
	}
}

func (r *ServerResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	idx, ok := req.ProviderData.(*robot.Index)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected Robot inventory index.")
		return
	}

	r.index = idx
}

func (r *ServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var m ServerModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, ok, e := r.index.Get(int(m.ServerNumber.ValueInt64()))
	if e != nil {
		resp.Diagnostics.AddError("Robot inventory failed", e.Error())
		return
	}

	if !ok {
		resp.Diagnostics.AddAttributeError(path.Root("server_number"), "Server not found", fmt.Sprintf("Robot does not report server %d.", m.ServerNumber.ValueInt64()))
		return
	}

	setServer(ctx, &m, s, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *ServerResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var m ServerModel

	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, ok, e := r.index.Get(int(m.ServerNumber.ValueInt64()))
	if e != nil {
		resp.Diagnostics.AddError("Robot inventory failed", e.Error())
		return
	}

	if !ok {
		resp.State.RemoveResource(ctx)
		return
	}

	setServer(ctx, &m, s, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *ServerResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var m ServerModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, ok, e := r.index.Get(int(m.ServerNumber.ValueInt64()))
	if e != nil {
		resp.Diagnostics.AddError("Robot inventory failed", e.Error())
		return
	}

	if !ok {
		resp.Diagnostics.AddError("Server not found", fmt.Sprintf("Robot does not report server %d.", m.ServerNumber.ValueInt64()))
		return
	}

	setServer(ctx, &m, s, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *ServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var m ServerModel

	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, ok, e := r.index.Get(int(m.ServerNumber.ValueInt64()))
	if e != nil {
		resp.Diagnostics.AddError("Robot inventory failed", e.Error())
		return
	}
	if ok {
		resp.Diagnostics.AddError("Server still exists", fmt.Sprintf("Robot still reports server %d; Terraform state is retained.", m.ServerNumber.ValueInt64()))
	}
}

func (r *ServerResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	n, e := strconv.ParseInt(req.ID, 10, 64)
	if e != nil || n <= 0 {
		resp.Diagnostics.AddError("Invalid import ID", "Use the positive Robot server number as the import ID.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("server_number"), n)...)
}

func setServer(ctx context.Context, m *ServerModel, s robot.Server, d *diag.Diagnostics) {
	m.ServerName = types.StringValue(s.ServerName)
	m.ServerIP = types.StringValue(s.ServerIP)
	m.ServerIPv6Net = types.StringValue(s.ServerIPv6Net)
	m.Product = types.StringValue(s.Product)
	m.DC = types.StringValue(s.DC)
	m.Traffic = types.StringValue(s.Traffic)
	m.Status = types.StringValue(s.Status)
	m.Cancelled = types.BoolValue(s.Cancelled)
	m.PaidUntil = types.StringValue(s.PaidUntil)
	m.IPs, _ = types.ListValueFrom(ctx, types.StringType, s.IPs)
	m.Subnets, _ = types.ListValueFrom(ctx, types.StringType, s.Subnets)
}

func (r *ServerResource) ModifyPlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
) {
	if r.index == nil {
		return
	}

	var prior, planned ServerModel
	priorExists := !req.State.Raw.IsNull()
	plannedExists := !req.Plan.Raw.IsNull()

	if priorExists {
		resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if plannedExists {
		resp.Diagnostics.Append(req.Plan.Get(ctx, &planned)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !priorExists && !plannedExists {
		return
	}

	n := planned.ServerNumber
	deleting := !plannedExists || n.IsNull()

	if deleting {
		n = prior.ServerNumber
	}

	if n.IsUnknown() || n.IsNull() {
		return
	}

	_, present, err := r.index.Get(int(n.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Robot inventory failed", err.Error())
		return
	}

	if deleting && present {
		resp.Diagnostics.AddError(
			"Cannot destroy server",
			fmt.Sprintf(
				"Robot still reports server %d; remove the server from Robot before removing this resource.",
				n.ValueInt64(),
			),
		)
		return
	}

	isServerNumberBeingReplaced :=
		!deleting &&
			priorExists &&
			!prior.ServerNumber.IsNull() &&
			!prior.ServerNumber.IsUnknown() &&
			prior.ServerNumber.ValueInt64() != planned.ServerNumber.ValueInt64()

	if isServerNumberBeingReplaced {
		_, oldPresent, oldErr := r.index.Get(int(prior.ServerNumber.ValueInt64()))

		if oldErr != nil {
			resp.Diagnostics.AddError("Robot inventory failed", oldErr.Error())
			return
		}

		if oldPresent {
			resp.Diagnostics.AddError("Cannot replace server", fmt.Sprintf("Robot still reports server %d; remove it before replacing this resource.", prior.ServerNumber.ValueInt64()))
			return
		}
	}

	if !deleting && !present {
		resp.Diagnostics.AddAttributeError(path.Root("server_number"), "Server not found", fmt.Sprintf("Robot does not report server %d.", n.ValueInt64()))
	}
}
