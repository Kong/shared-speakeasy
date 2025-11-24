resource "konnect_mesh" "default" {
  type   = "Mesh"
  name   = "default"
  config = module.networking.vpc.subnet.cidr_block
}
