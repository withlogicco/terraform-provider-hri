terraform {
  required_providers {
    hri = { source = "withlogicco/hri" }
  }
}

provider "hri" {}

data "hri_servers" "all" {
  dc                = "HEL1-DC11"
  product           = "Server Auction"
  name_regex        = "^containership-"
  include_cancelled = false
}
