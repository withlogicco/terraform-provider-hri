package provider

import (
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceSchemaAttributes map[string]resourceschema.Attribute
type dataSourceSchemaAttributes map[string]datasourceschema.Attribute

func optionalStringDataSourceAttribute(description string) datasourceschema.StringAttribute {
	return datasourceschema.StringAttribute{
		Optional:            true,
		MarkdownDescription: description,
	}
}

type serverAttributeKind int

const (
	serverStringAttribute serverAttributeKind = iota
	serverBoolAttribute
	serverStringListAttribute
)

type serverAttributeDefinition struct {
	name        string
	kind        serverAttributeKind
	description string
}

var computedServerAttributeDefinitions = []serverAttributeDefinition{
	{
		name:        "server_name",
		kind:        serverStringAttribute,
		description: "Server name reported by Robot.",
	},
	{
		name:        "server_ip",
		kind:        serverStringAttribute,
		description: "Main server IPv4 address.",
	},
	{
		name:        "server_ipv6_net",
		kind:        serverStringAttribute,
		description: "Main server IPv6 network.",
	},
	{
		name:        "product",
		kind:        serverStringAttribute,
		description: "Robot product name.",
	},
	{
		name:        "dc",
		kind:        serverStringAttribute,
		description: "Robot data center.",
	},
	{
		name:        "traffic",
		kind:        serverStringAttribute,
		description: "Included traffic quota.",
	},
	{
		name:        "status",
		kind:        serverStringAttribute,
		description: "Server status reported by Robot.",
	},
	{
		name:        "cancelled",
		kind:        serverBoolAttribute,
		description: "Whether cancellation is active.",
	},
	{
		name:        "paid_until",
		kind:        serverStringAttribute,
		description: "Date through which the server is paid.",
	},
	{
		name:        "ips",
		kind:        serverStringListAttribute,
		description: "Single IP addresses assigned to the server.",
	},
	{
		name:        "subnets",
		kind:        serverStringListAttribute,
		description: "Assigned subnets in CIDR notation.",
	},
}

func computedServerResourceAttributes() resourceSchemaAttributes {
	attributes := make(resourceSchemaAttributes, len(computedServerAttributeDefinitions))

	for _, definition := range computedServerAttributeDefinitions {
		switch definition.kind {
		case serverStringAttribute:
			attributes[definition.name] = resourceschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: definition.description,
			}
		case serverBoolAttribute:
			attributes[definition.name] = resourceschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: definition.description,
			}
		case serverStringListAttribute:
			attributes[definition.name] = resourceschema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: definition.description,
			}
		}
	}

	return attributes
}

func computedServerDataSourceAttributes() dataSourceSchemaAttributes {
	attributes := make(dataSourceSchemaAttributes, len(computedServerAttributeDefinitions))

	for _, definition := range computedServerAttributeDefinitions {
		switch definition.kind {
		case serverStringAttribute:
			attributes[definition.name] = datasourceschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: definition.description,
			}
		case serverBoolAttribute:
			attributes[definition.name] = datasourceschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: definition.description,
			}
		case serverStringListAttribute:
			attributes[definition.name] = datasourceschema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: definition.description,
			}
		}
	}

	return attributes
}
