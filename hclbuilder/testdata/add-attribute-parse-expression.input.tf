resource "kong-mesh_mesh" "default" {
  type = "Mesh"
  name = "default"
  routing = {
    default_forbid_mesh_external_service_access = true
  }
}
