resource "kong-mesh_mesh" "default" {
  name                           = "mesh-1"
  skip_creating_initial_policies = ["*"]
  type                           = "Mesh"
}
