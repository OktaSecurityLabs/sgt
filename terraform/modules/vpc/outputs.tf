output "WebappVCPID" {
  value = "${aws_vpc.sgt-vpc.id}"
}

output "sgt_private_us_east_1a_id" {
  value = "${aws_subnet.sgt-Private_us_east_1a.id}"
}

output "sgt_private_us_east_1b_id" {
  value = "${aws_subnet.sgt-Private_us_east_1b.id}"
}

output "sgt_public_subnet_us_east_1a_id" {
  value = "${aws_subnet.sgt-PublicSubnet_us_east_1a.id}"
}

output "sgt_public_subnet_us_east_1b_id" {
  value = "${aws_subnet.sgt-PublicSubnet_us_east-1b.id}"
}

output "nat_gateway_id" {
  value = "${aws_nat_gateway.sgt-Nat.id}"
}
