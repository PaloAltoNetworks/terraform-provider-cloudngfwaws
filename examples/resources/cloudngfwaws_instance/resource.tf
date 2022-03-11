resource "cloudngfwaws_instance" "example" {
  name        = "example-instance"
  vpc_id      = aws_vpc.example.id
  account_id  = "12345678"
  description = "Example description"

  endpoint_mode = "ServiceManaged"
  subnet_mapping {
    subnet_id = aws_subnet.subnet1.id
  }

  subnet_mapping {
    subnet_id = aws_subnet.subnet2.id
  }

  rulestack = "example-rulestack"

  tags = {
    Foo = "bar"
  }
}


resource "aws_vpc" "example" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "tf-example"
  }
}

resource "aws_subnet" "subnet1" {
  vpc_id            = aws_vpc.my_vpc.id
  cidr_block        = "172.16.10.0/24"
  availability_zone = "us-west-2a"

  tags = {
    Name = "tf-example"
  }
}

resource "aws_subnet" "subnet2" {
  vpc_id            = aws_vpc.my_vpc.id
  cidr_block        = "172.16.20.0/24"
  availability_zone = "us-west-2b"

  tags = {
    Name = "tf-example"
  }
}
