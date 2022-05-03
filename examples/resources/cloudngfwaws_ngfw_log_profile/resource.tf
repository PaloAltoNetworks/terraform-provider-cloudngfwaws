resource "cloudngfwaws_ngfw_log_profile" "example" {
  ngfw       = cloudngfwaws_ngfw.x.name
  account_id = cloudngfwaws_ngfw.x.account_id
  log_destination {
    destination_type = "S3"
    destination      = "my-s3-bucket"
    log_type         = "TRAFFIC"
  }
  log_destination {
    destination_type = "CloudWatchLogs"
    destination      = "panw-log-group"
    log_type         = "THREAT"
  }
}

resource "cloudngfwaws_ngfw" "x" {
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
