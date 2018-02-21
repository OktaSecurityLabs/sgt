output "sgt-node-user-access-key-id" {
  value = "${module.firehose.sgt-node-user-access-key-id}"
}

output "sgt-node-user-secret-access-key" {
  value = "${module.firehose.sgt-node-user-secret-access-key}"
}

output "sgt-distributed-firehose-stream-name" {
  value = "${module.firehose.sgt-distributed-firehose-stream-name}"
}

output "sgt-distributed-firehose-stream-arn" {
  value = "${module.firehose.sgt-distributed-firehose-stream-arn}"
}

output "sgt-firehose-stream-name" {
  value = "${module.firehose.sgt-firehose-stream-name}"
}