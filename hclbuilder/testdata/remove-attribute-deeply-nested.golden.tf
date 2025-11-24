resource "kong-mesh_mesh" "default" {
  name = "mesh-1"
  type = "Mesh"
  config = {
    networking = {
      advanced = {
        retries = 3
      }
      basic = {
        enabled = true
      }
    }
    security = {
      enabled = true
    }
  }
}
