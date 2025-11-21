resource "kong-mesh_mesh" "default" {
  type                           = "Mesh"
  name                           = "mesh-1"
  skip_creating_initial_policies = ["*"]
}

resource "kong-mesh_mesh_traffic_permission" "existing" {
  type = "MeshTrafficPermission"
  name = "allow-all"
  mesh = "kong-mesh_mesh.default.name"
}
resource "kong-mesh_mesh" {
}
resource "kong-mesh_mesh_traffic_permission" {
}
resource "kong-mesh_mesh_traffic_permission" "new_policy" {
  name = "new-policy"
}
