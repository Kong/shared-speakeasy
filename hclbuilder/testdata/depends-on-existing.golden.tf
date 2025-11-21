resource "kong-mesh_mesh" "default" {
  type       = "Mesh"
  name       = "default"
  depends_on = [konnect_mesh_control_plane.my_meshcontrolplane]
}
