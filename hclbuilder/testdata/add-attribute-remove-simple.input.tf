resource "kong-mesh_mesh" "default" {
  type = "Mesh"
  name = "default"
  skip_creating_initial_policies = ["*"]
}
