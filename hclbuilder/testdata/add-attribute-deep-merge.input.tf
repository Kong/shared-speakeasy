resource "kong-mesh_mesh" "default" {
  type = "Mesh"
  name = "default"
  constraints = {
    dataplane_proxy = {
      requirements = [{
        tags = {
          key = "a"
        }
      }]
    }
  }
}
