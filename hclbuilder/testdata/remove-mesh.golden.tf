provider "kong-mesh" {
  server_url = "http://localhost:5681"
}
resource "kong-mesh_mesh" "other" {
  name = "mesh-2"
  type = "Mesh"
}
