provider "kong-mesh" {
  server_url = "http://localhost:5681"
}
resource "kong-mesh_mesh_traffic_permission" "allow_all" {
  mesh     = "kong-mesh_mesh.default.name"
  name     = "allow-all"
  provider = "kong-mesh"
  spec = {
    from = [{
      target_ref = {
        kind = "Mesh"
      }
    }]
  }
  type = "MeshTrafficPermission"
}
