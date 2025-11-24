resource "kong-mesh_mesh" "default" {
  name = "mesh-1"
  type = "Mesh"
  config = {
    networking = {
      basic = {
        enabled = true
      }
      advanced = {
        retries = 3
        timeout = 30
      }
    }
    security = {
      enabled = true
    }
  }
}
