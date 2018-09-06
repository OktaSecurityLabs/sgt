output "elasticsearch_domain_arn" {
  value = "${module.elasticsearch.elasticsearch_domain_arn}"
}

output "elasticsearch_domain_id" {
  value = "${module.elasticsearch.elasticsearch_domain_id}"
}

output "elasticearch_endpoint" {
  value = "${module.elasticsearch.elasticsearch_endpoint}"
}

output "cognito_identity_pool_id" {
  value = "${module.elasticsearch.cognito_identity_pool_id}"
}

output "cognito_user_pool_id" {
  value = "${module.elasticsearch.cognito_user_pool_id}"
}

output "elasticsearch_cognito_role_arn" {
  value = "${module.elasticsearch.elasticsearch_cognito_role_arn}"
}

output "elasticsearch_region" {
  value = "us-east-1"
}
