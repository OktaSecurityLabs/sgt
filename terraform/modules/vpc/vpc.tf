provider "aws" {
  profile = "${var.aws_profile}"
  region = "us-east-1"
}

resource "aws_vpc" "sgt-vpc" {
  cidr_block = "10.0.0.0/16"
  enable_dns_support = true
  enable_dns_hostnames = true
  tags {
    Name = "${var.vpc_name}"
  }
}

resource "aws_subnet" "sgt-PublicSubnet_us_east_1a" {
  vpc_id = "${aws_vpc.sgt-vpc.id}"
  cidr_block = "10.0.100.0/24"
  map_public_ip_on_launch = true
  availability_zone = "us-east-1a"
  tags {
    Name = "sgt_public_us_east_1a"
  }
}

resource "aws_subnet" "sgt-PublicSubnet_us_east-1b" {
  cidr_block = "10.0.101.0/24"
  map_public_ip_on_launch = true
  vpc_id = "${aws_vpc.sgt-vpc.id}"
  availability_zone = "us-east-1b"
  tags {
    Name = "sgt_public_us_east_1b"
  }
}

resource "aws_subnet" "sgt-Private_us_east_1a" {
  vpc_id = "${aws_vpc.sgt-vpc.id}"
  cidr_block = "10.0.200.0/24"
  map_public_ip_on_launch = false
  availability_zone = "us-east-1a"
  tags {
    Name = "sgt-Private_us_east_1a"
  }
}

resource "aws_subnet" "sgt-Private_us_east_1b" {
  vpc_id = "${aws_vpc.sgt-vpc.id}"
  cidr_block = "10.0.201.0/24"
  map_public_ip_on_launch = false
  availability_zone = "us-east-1b"
  tags {
    Name = "sgt-Private_us_east_1b"
  }
}

resource "aws_internet_gateway" "sgt-VPC_IGW" {
  vpc_id = "${aws_vpc.sgt-vpc.id}"
  tags {
    Name = "sgt-VPC_IGW"
  }
}
#
#resource "aws_route" "WebappExternalRoute" {
  #route_table_id = "${aws_vpc.sgt-vpc.main_route_table_id}"
  #destination_cidr_block = "0.0.0.0/0"
  #gateway_id = "${aws_internet_gateway.sgt-VPC_IGW.id}"
#}

resource "aws_eip" "sgt-IGW_EIP" {
  vpc = true
  depends_on = ["aws_internet_gateway.sgt-VPC_IGW"]
}

resource "aws_nat_gateway" "sgt-Nat" {
  allocation_id = "${aws_eip.sgt-IGW_EIP.id}"
  subnet_id = "${aws_subnet.sgt-PublicSubnet_us_east_1a.id}"
  depends_on = ["aws_internet_gateway.sgt-VPC_IGW"]
}

resource "aws_route_table" "sgt_public_route_table" {
  vpc_id = "${aws_vpc.sgt-vpc.id}"
  tags {
    Name = "sgt-public_route_table"
  }
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.sgt-VPC_IGW.id}"
  }
}

resource "aws_route_table" "sgt-PrivateRouteTable" {
  vpc_id = "${aws_vpc.sgt-vpc.id}"
  tags {
    Name = "$sgt-PrivateRouteTable"
  }
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_nat_gateway.sgt-Nat.id}"
  }
}

#resource "aws_route" "WebappPrivateRoute" {
  #route_table_id = "${aws_route_table.sgt-PrivateRouteTable.id}"
  #destination_cidr_block = "0.0.0.0/0"
  #nat_gateway_id = "${aws_nat_gateway.sgt-Nat.id}"
#}

resource "aws_route_table_association" "sgt-Public_us_east_1a_association" {
  subnet_id = "${aws_subnet.sgt-PublicSubnet_us_east_1a.id}"
  route_table_id = "${aws_route_table.sgt_public_route_table.id}"
}
resource "aws_route_table_association" "sgt_public_us_east_1b_association" {
  subnet_id = "${aws_subnet.sgt-PublicSubnet_us_east-1b.id}"
  route_table_id = "${aws_route_table.sgt_public_route_table.id}"
}

resource "aws_route_table_association" "sgt-Private_us_east_1a_association" {
  subnet_id = "${aws_subnet.sgt-Private_us_east_1a.id}"
  route_table_id = "${aws_route_table.sgt-PrivateRouteTable.id}"
}

resource "aws_route_table_association" "sgt-Private_us_east_1b_association" {
  subnet_id = "${aws_subnet.sgt-Private_us_east_1b.id}"
  route_table_id = "${aws_route_table.sgt-PrivateRouteTable.id}"
}

