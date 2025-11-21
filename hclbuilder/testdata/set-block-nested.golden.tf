resource "kong-mesh_mesh_traffic_permission" "allow_all" {
  mesh = "kong-mesh_mesh.default.name"
  name = "allow-all"
  spec = {
    from = [{
      target_ref = {
        kind = "Mesh"
      }
    }]
  }
  type = "MeshTrafficPermission"
}
