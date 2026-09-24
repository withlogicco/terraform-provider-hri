terraform {
  required_providers { hri = { source = "withlogicco/hri" } }
}
provider "hri" {}
resource "hri_server" "containership_01" { server_number = 161051 }

data "hri_servers" "all" {}

locals {
  tracked   = toset([hri_server.containership_01.server_number])
  untracked = setsubtract(data.hri_servers.all.server_numbers, local.tracked)
}

check "all_servers_tracked" {
  assert {
    condition     = length(local.untracked) == 0
    error_message = "Untracked Hetzner servers: ${join(", ", local.untracked)}"
  }
}
