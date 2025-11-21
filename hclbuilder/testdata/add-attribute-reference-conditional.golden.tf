resource "konnect_mesh" "default" {
  type    = "Mesh"
  name    = "default"
  enabled = var.environment == "production" ? true : false
}
