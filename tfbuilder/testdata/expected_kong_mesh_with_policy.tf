provider "kong-mesh" {
  server_url = "http://localhost:5681"
}

resource "kong-mesh_mesh" "default" {
  provider = "kong-mesh"
  type = "Mesh"
  name = "mesh-1"

  skip_creating_initial_policies = [ "*" ]
}

resource "kong-mesh_mesh_traffic_permission" "allow_all" {
  provider = "kong-mesh"
  mesh = kong-mesh_mesh.default.name
  depends_on = [kong-mesh_mesh.default]
  labels = {
    "kuma.io/mesh" = "kong-mesh_mesh.default.name"
  }
  type = "MeshTrafficPermission"
  name = "allow-all"

  
  spec = {
    from = [
      {
        target_ref = {
          kind = "MeshSubset"
          proxy_types = ["Sidecar"]
        }
        default = {
        }
      }
    ]
  }
}

