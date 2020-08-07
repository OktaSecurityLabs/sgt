# Welcome to version 0.2.0!

0.2.0 brings a large amount of changes from [0.1.6](https://github.com/OktaSecurityLabs/sgt/tree/0.1.6)

### Highlights include:

- Cognito Auth for Kibana
- Support for the osquery filecarver
- A new S3 backend for terraform state.
- An option to auto-approve new nodes with the correct node secret

And more!


## Upgrading:

Upgrading your deployment from pre - 0.2.0 should be relatively painless.  Everything
is still deployed as normal, via `./sgt deploy -env <your env> -all`.

Note, however, that your your env.json config file will need some new additions to support
the new features.  Additionally, to support the new s3 backend, you will need to create
an s3 bucket to keep your terraform state files in.

### Config file - env.json:

Found in the `sgt/terraform/<env>/ folder, your `env.json` file controls the configuration
for your environemnt.  Now that Cognito auth has been added, you will need to add a few new fields
to the config to support cognito.

##### cognito

```json
  "users":
    [
    "my.user"
    ],
  "mail_domain":"example.com",
```

these additional fields tell cognito what domain to user for authentication and which users to add.

__NOTE__: Make sure these users actually are valid email addresses, `user.name@domain.com`.  Cognito will automatically
send these users a temporary password which needs to be changed on first login.  If they do not
exist, they will not be able to receive the password.

##### terraform backend

Create an s3 bucket and enable versioning on it. This will allow you to automatically back up your
configuration should anything ever happen to it.

This requires on additional field be added to the config:
```json
    "terraform_backend_bucket_name":"my-new-fancy-backend-bucket",
  }

```

##### auto approving nodes:

Lastly, the field to tell SGT whether or not to auto-approve new nodes (with the correct secret)

```json
  "auto_approve_nodes":"true"
}
```

the full config should look something like this:

```json
{
  "environment":"my-example-env",
  "aws_profile":"my-aws-profile",
  "user_ip_address":"127.0.0.1",
  "sgt_osquery_results_bucket_name":"my-osquery-log-bucket-name",
  "sgt_config_bucket_name":"my-osquery-config-bucket-name",
  "domain":"example.com",
  "subdomain":"subdomain",
  "aws_keypair":"my-aws-keypair",
  "full_ssl_certchain":"subdomain.example.com.fullchain.pem",
  "ssl_private_key":"subdomain.example.com.privkey.pem",
  "sgt_node_secret":"my-node-enroll-secret",
  "sgt_app_secret":"super-ultra-mega-app-secret",
  "create_elasticsearch":1,
  "asg_desired_size":1,
  "asg_min_size": 1,
  "aws_region":"us-east-1",
  "users":
  [
    "user.name"
  ],
  "mail_domain":"example.com",
  "terraform_backend_bucket_name":"my-new-fancy-backend-bucket",
  "auto_approve_nodes":"true"
}
```


Once these fields are set, you should be able to run a normal deploy to be
upgraded to the new version


### Bug fixes
:bug: :bug: :bug:
##### Duplicate nodes.

Default configs have been updated to use the uuid field for host_identifier in osquery,
making these fields truly unique.  This was previously an issue due to hostname being
derived from mutable data on the host.

Additionally, an added check for node_keys and host_identifiers should resolve an instance when
a single host could recieve multiple node_keys on a failed enrollment attempt.
