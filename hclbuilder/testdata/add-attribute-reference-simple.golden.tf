resource "konnect_mesh" "default" {
  type  = "Mesh"
  name  = "default"
  cp_id = konnect_mesh_control_plane.my_cp.id
}
