# Terraform Provider: Hetzner Robot Inventory

The `hri` provider inventories existing Hetzner dedicated servers through the Robot API. It tracks each server by its Robot server number and exposes server details through Terraform resources and data sources.

## Requirements

- Terraform 1.9.8 or later
- A Hetzner Robot webservice user with access to the servers being inventoried

> [!IMPORTANT]
> 
> A Hetzner Robot webservice user is different than the user you are logging into Hetzner.
> 
> To create one, navigate to https://robot.hetzner.com and follow the process in the Settings page, accessible from the top-right user avatar icon.

Set the Robot webservice credentials in the environment:

```sh
export HETZNER_ROBOT_USER="your-webservice-user"
export HETZNER_ROBOT_PASSWORD="your-webservice-password"
```

## Installation

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_version = ">= 1.9.8"

  required_providers {
    hri = {
      source  = "withlogicco/hri"
      version = "~> 0.1.0"
    }
  }
}

provider "hri" {}
```

Run `terraform init` to install the provider.

## Usage

Inventory an existing Robot server into Terraform state by its server number:

```hcl
resource "hri_server" "containership_01" {
  server_number = 161051
}
```

The resource reads server inventory into state. Before removing the resource block, delete the server in Robot; Terraform validates that Robot reports the server as absent before removing the resource from state.

Read one server or list the account inventory with data sources:

```hcl
data "hri_server" "containership_01" {
  server_number = 161051
}

data "hri_servers" "all" {}
```

Import a server that is already represented by a resource block:

```sh
terraform import hri_server.containership_01 161051
```

See the [resource documentation](docs/resources/server.md), [single-server data source](docs/data-sources/server.md), and [servers data source](docs/data-sources/servers.md) for the available attributes and filters.

## Development and local testing

Run the provider's automated checks from the repository root:

```sh
go test ./...
go vet ./...
go build ./...
terraform fmt -check -recursive
```

The tests use a fake Robot API and do not require Robot credentials or a live account. To exercise the provider CLI locally before a Registry release, build the binary and install it in a Terraform filesystem mirror for your platform. For example, on macOS with Apple Silicon:

```sh
mkdir -p /tmp/hri-mirror/registry.terraform.io/withlogicco/hri/0.1.0/darwin_arm64
go build -o /tmp/hri-mirror/registry.terraform.io/withlogicco/hri/0.1.0/darwin_arm64/terraform-provider-hri_v0.1.0 .
```

Create a Terraform CLI config file at `/tmp/hri.tfrc`:

```hcl
provider_installation {
  filesystem_mirror {
    path    = "/tmp/hri-mirror"
    include = ["registry.terraform.io/withlogicco/hri"]
  }

  direct {
    exclude = ["registry.terraform.io/withlogicco/hri"]
  }
}
```

In a Terraform configuration that requires `withlogicco/hri` version `0.1.0`, run:

```sh
TF_CLI_CONFIG_FILE=/tmp/hri.tfrc terraform init
TF_CLI_CONFIG_FILE=/tmp/hri.tfrc terraform plan
```

The example above targets `darwin_arm64`. Use the matching `GOOS_GOARCH` directory for other platforms. For live API smoke testing, set the Robot environment variables above and use the examples under [`examples/`](examples/).

## Documentation

- [Provider configuration](docs/index.md)
- [`hri_server` resource](docs/resources/server.md)
- [`hri_server` data source](docs/data-sources/server.md)
- [`hri_servers` data source](docs/data-sources/servers.md)
