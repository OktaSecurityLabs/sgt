output "elasticsearch_domain_arn" {
  value = "${aws_elasticsearch_domain.sgt-osquery_results.arn}"
}

output "elasticsearch_endpoint" {
  value = "${aws_elasticsearch_domain.sgt-osquery_results.endpoint}"
}

output "elasticsearch_domain_id" {
  value = "${aws_elasticsearch_domain.sgt-osquery_results.domain_name}"
}

output "cognito_user_pool_id" {
  value = "${aws_cognito_user_pool.pool.id}"
}

output "cognito_identity_pool_id" {
  value = "${aws_cognito_identity_pool.main.id}"
}

output "elasticsearch_cognito_role_arn" {
  value = "${aws_iam_role.es_cognito_access_role.arn}"
}

output "elasticsearch_region" {
  value = "us-east-1"
}
