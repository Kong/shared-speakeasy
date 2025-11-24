resource "kong-mesh_mesh" "default" {
  name = "mesh-1"
  type = "Mesh"
  routing = {
    default_forbid_mesh_external_service_access = true
    locality_aware_load_balancing                = false
  }
}
