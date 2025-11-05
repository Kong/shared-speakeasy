variable "existing" {
  default = "original-value"
}

resource "aws_instance" "old" {
  ami = "ami-old-123"
}

resource "aws_instance" "keep" {
  ami           = "ami-keep-456"
  instance_type = "t2.small"
}
