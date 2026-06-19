// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nfagerlund/terraform-provider-bombnull/internal/planmodifiers"
)

var (
	_ resource.Resource = (*bombNullResource)(nil)
)

func NewBombNullResource() resource.Resource {
	return &bombNullResource{}
}

type bombNullResource struct{}

func (n *bombNullResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource"
}

func (n *bombNullResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The `bombnull_resource` resource implements the standard resource lifecycle, and also raises errors (failing your Terraform run) at configurable points in that lifecycle.\n\n" +
			"The `triggers` argument allows specifying an arbitrary set of values that, when changed, will cause the resource to be replaced. The `notriggers` argument is similar but only causes in-place updates.",
		Attributes: map[string]schema.Attribute{
			"triggers": schema.MapAttribute{
				Description: "A map of arbitrary strings that, when changed, will force the null resource to be replaced, re-running any associated provisioners.",
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.Map{
					planmodifiers.RequiresReplaceIfValuesNotNull(),
				},
			},

			"nontriggers": schema.MapAttribute{
				Description: "A map of arbitrary strings that, when changed, will cause a normal Update of the resource without requiring replacement.",
				ElementType: types.StringType,
				Optional:    true,
			},

			"id": schema.StringAttribute{
				Description: "This is set to a random value at create time.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"bomb_create": schema.BoolAttribute{
				Description: "Whether to error during create",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},

			"bomb_update": schema.BoolAttribute{
				Description: "Whether to error during update",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},

			"bomb_delete": schema.BoolAttribute{
				Description: "Whether to error during delete",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},

			"bomb_every_time": schema.BoolAttribute{
				Description: "Whether to error every time one of the bombed operations is attempted. If false, subsequent operations will succeed. Doesn't apply to bomb_create, which always bombs no matter what.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},

			"tried_update": schema.BoolAttribute{
				Description: "Gets set to true the first time you try to update the resource. Subsequent attempts will succeed.",
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},

			"tried_delete": schema.BoolAttribute{
				Description: "Gets set to true the first time you try to destroy the resource. Subsequent attempts will succeed.",
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (n *bombNullResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model nullModelV0

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// create doesn't get to read state, so it always bombs.
	bomb := model.BombCreate.ValueBool()

	if bomb {
		resp.Diagnostics.AddError("five, four, three, two", "make it boom, make it boom, make it boom, make it boom")
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)

	diags := resp.State.SetAttribute(ctx, path.Root("id"), fmt.Sprintf("%d", rand.Int()))
	resp.Diagnostics.Append(diags...)
}

func (n *bombNullResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (n *bombNullResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model nullModelV0
	var statemodel nullModelV0

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &statemodel)...)

	if resp.Diagnostics.HasError() {
		return
	}

	bomb := model.BombUpdate.ValueBool()
	always := model.BombEveryTime.ValueBool()
	tried := statemodel.TriedUpdate.ValueBool()
	attr := path.Root("tried_update")

	if bomb && (always || !tried) {
		diags := resp.State.SetAttribute(ctx, attr, true)
		resp.Diagnostics.Append(diags...)
		resp.Diagnostics.AddError("five, four, three, two", "make it boom, make it boom, make it boom, make it boom")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (n *bombNullResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model nullModelV0

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	bomb := model.BombDelete.ValueBool()
	always := model.BombEveryTime.ValueBool()
	tried := model.TriedDelete.ValueBool()
	attr := path.Root("tried_delete")

	if bomb && (always || !tried) {
		diags := resp.State.SetAttribute(ctx, attr, true)
		resp.Diagnostics.Append(diags...)
		resp.Diagnostics.AddError("five, four, three, two", "make it boom, make it boom, make it boom, make it boom")
	}
}

type nullModelV0 struct {
	Triggers      types.Map    `tfsdk:"triggers"`
	NonTriggers   types.Map    `tfsdk:"nontriggers"`
	ID            types.String `tfsdk:"id"`
	BombEveryTime types.Bool   `tfsdk:"bomb_every_time"`
	BombCreate    types.Bool   `tfsdk:"bomb_create"`
	BombUpdate    types.Bool   `tfsdk:"bomb_update"`
	BombDelete    types.Bool   `tfsdk:"bomb_delete"`
	TriedUpdate   types.Bool   `tfsdk:"tried_update"`
	TriedDelete   types.Bool   `tfsdk:"tried_delete"`
}
