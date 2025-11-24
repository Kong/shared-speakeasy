provider "kong-mesh" {
  server_url = "http://localhost:5681"
}
resource "kong-mesh_mesh" "default" {
  type                           = "Mesh"
  name                           = "default"
  skip_creating_initial_policies = ["*"]
}
