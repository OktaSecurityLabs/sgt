output "vpc_id" {
  value = "${module.vpc.WebappVCPID}"
}

output "sgt_private_us_east_1a_id" {
  value = "${module.vpc.sgt_private_us_east_1a_id}"
}

output "sgt_private_us_east_1b_id" {
  value = "${module.vpc.sgt_private_us_east_1b_id}"
}

output "sgt_public_subnet_us_east_1a_id" {
  value = "${module.vpc.sgt_public_subnet_us_east_1a_id}"
}

output "sgt_public_subnet_us_east_1b_id" {
  value = "${module.vpc.sgt_public_subnet_us_east_1b_id}"
}

output "nat_gateway_id" {
  value = "${module.vpc.nat_gateway_id}"
}