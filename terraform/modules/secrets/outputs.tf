output "sgt_app_secret_id" {
  value = "${aws_ssm_parameter.sgt_app_secret.name}"
}
output "sgt_node_secret_id" {
  value = "${aws_ssm_parameter.sgt_node_secret.name}"
}