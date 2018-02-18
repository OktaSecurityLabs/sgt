output "sgt-node-user-access-key-id" {
  value = "${aws_iam_access_key.node-firehose-access-key.id}"
}

output "sgt-node-user-secret-access-key" {
  value = "${aws_iam_access_key.node-firehose-access-key.secret}"
}

output "sgt-distributed-firehose-stream-name" {
  value = "${aws_kinesis_firehose_delivery_stream.sgt-firehose-distributed-osquery_results.name}"
}

output "sgt-distributed-firehose-stream-arn" {
  value = "${aws_kinesis_firehose_delivery_stream.sgt-firehose-distributed-osquery_results.arn}"
}

output "sgt-firehose-stream-name" {
  value = "${aws_kinesis_firehose_delivery_stream.sgt-firehose-osquery_results.name}"
}