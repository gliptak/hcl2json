provider "aws" {
  profile    = "default"
  region     = "us-east-1"
}

resource "aws_instance" "example1" {
  ami           = "ami-123456"
  instance_type = "t3.micro"
}

resource "aws_instance" "example2" {
  ami           = "ami-123456"
  instance_type = "t3.micro"
}