resource "kong-mesh_mesh" "default" {
  name = "mesh-1"
  type = "Mesh"
  routing = {
    locality_aware_load_balancing = false
  }
}
