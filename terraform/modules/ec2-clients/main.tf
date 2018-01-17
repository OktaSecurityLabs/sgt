provider "aws" {
  region = "us-east-1"
}

resource "aws_security_group" "osquery_client_sg" {
  name = "osquery-client-sg"
}

resource "aws_security_group_rule" "osquery_client_rules_ingress" {
  type = "ingress"
  from_port = 22
  to_port = 22
  cidr_blocks = ["0.0.0.0/0"]
  protocol = "tcp"
  security_group_id = "${aws_security_group.osquery_client_sg.id}"
}

resource "aws_security_group_rule" "osquery_client_rules_egress" {
  type = "egress"
  from_port = 0
  to_port = 65535 protocol = "all"
  cidr_blocks = ["0.0.0.0/0"]
  security_group_id = "${aws_security_group.osquery_client_sg.id}"
}

data template_file "osquery_client_userdata" {
  template = "${file("userdata.sh")}"
}


resource "aws_instance" "osquery_client_test" {
  count = "${var.num_clients}"
  ami = "ami-d15a75c7"
  instance_type = "t2.micro"
  key_name = "matt-terraform-key"
  associate_public_ip_address = true
  user_data = "${data.template_file.osquery_client_userdata.rendered}"
  security_groups = ["${aws_security_group.osquery_client_sg.name}"]
  provisioner "file" {
    source = "bundle.pem"
    destination = "cert_bundle.pem"
    connection {
      type = "ssh"
      user = "ubuntu"
      private_key = "${file("matt-terraform-key.pem")}"
    }
  }
  provisioner "file" {
    source = "osquery.secret"
    destination = "osquery.secret"
    connection {
      type = "ssh"
      user = "ubuntu"
      private_key = "${file("matt-terraform-key.pem")}"
    }
  }
  provisioner "file" {
    source = "osquery.flags.default"
    destination = "osquery.flags.default"
    connection {
      type = "ssh"
      user = "ubuntu"
      private_key = "${file("matt-terraform-key.pem")}"
    }
  }
  tags {
    Name = "oquery_client_${count.index}"
  }
  depends_on = ["aws_security_group.osquery_client_sg"]
}
