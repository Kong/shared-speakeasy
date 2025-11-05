variable "existing" {
  default = "original-value"
}


resource "aws_instance" "keep" {
  ami           = "ami-keep-456"
  instance_type = "t2.small"
}
variable "new_var" {
  default = "added"
}
