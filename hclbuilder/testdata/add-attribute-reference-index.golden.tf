resource "konnect_mesh" "default" {
  type  = "Mesh"
  name  = "default"
  value = data.aws_instances.example.ids[0]
}
