resource "cloudngfwaws_ngfw" "example" {
  name        = "example-instance"
  description = "Example description"
  az_list     = ["use1-az1"]

  rulestack = cloudngfwaws_commit_rulestack.rs.rulestack

  tags = {
    Foo = "bar"
  }
}

resource "cloudngfwaws_commit_rulestack" "rs" {
  rulestack = "my-rulestack"
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
