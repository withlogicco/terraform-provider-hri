terraform {
  required_providers {
    hri = { source = "withlogicco/hri" }
  }
}

provider "hri" {}

data "hri_server" "containership_01" {
  server_number = 161051
}
