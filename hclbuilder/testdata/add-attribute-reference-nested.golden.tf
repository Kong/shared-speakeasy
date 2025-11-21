resource "konnect_mesh_traffic_permission" "allow_all" {
  type     = "MeshTrafficPermission"
  name     = "allow-all"
  mesh_ref = konnect_mesh.default.name
}
