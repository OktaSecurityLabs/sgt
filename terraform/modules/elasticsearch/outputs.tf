output "elasticsearch_domain_arn" {
  value = "${aws_elasticsearch_domain.sgt-osquery_results.arn}"
}

output "elasticsearch_endpoint" {
  value = "${aws_elasticsearch_domain.sgt-osquery_results.endpoint}"
}