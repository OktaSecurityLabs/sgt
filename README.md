## SGT: OSQuery Management Server Built Entirely on AWS!
[![Build Status](https://travis-ci.org/OktaSecurityLabs/sgt.svg?branch=master)](https://travis-ci.org/OktaSecurityLabs/sgt)
[![Go Report](https://goreportcard.com/badge/github.com/OktaSecurityLabs/sgt)](https://goreportcard.com/badge/github.com/OktaSecurityLabs/sgt)

SGT is an osquery management server written in Golang and built in aws.  Sgt (Simple Go TLS)
is backed entirely by AWS services, making its infrastructure requirements extremely
simple, robust and scalable.

SGT is managed entirely through terraform

###### NOTE: SGT is under active development.  Please help us improve by submitting issues!


### Getting started.

Getting started with sgt is designed to be very simple with minimal setup required.  To get started, however, you will need a FEW things first.


##### prereqs:
1. An [AWS account](https://aws.amazon.com/free/) with admin access to DynamoDB, EC2, ES (ElastisearchService), Kinesis/Firehose and IAM. (note, this must be programatic access, so you can have an access key and secret to use)
2. [Golang 1.8.2+]((https://golang.org/doc/install))
3. [Terraform 11.0+](https://www.terraform.io/intro/getting-started/install.html)
4. A domain with DNS [managed via Route53](http://docs.aws.amazon.com/Route53/latest/DeveloperGuide/MigratingDNS.html) (Note: This does not mean you need to buy a domain, you can use an existing domain and just  manage DNS on Route53)
5. An SSL cert with public and private keypair. This will be used to terminate TLS connections to our server see [Obtaining a free ssl cert for SGT with Letsencrypt for one method of aquiring a certificate](https://github.com/OktaSecurityLabs/sgt/blob/master/docs/letsencrypt_cert_instructions.md)
6. An aws [profile configured](https://docs.aws.amazon.com/cli/latest/userguide/cli-multiple-profiles.html).


## Installation

1.  Clone the repo
    ```commandline

    go get github.com/OktaSecurityLabs/sgt
    ```
2. change into the downloaded directory
    ```commandline
    cd $GOPATH/src/github.com/OktaSecurityLabs/sgt
    ```

3.  Build the project
    ```commandline
    go build
    ```
4.  Copy your ssl certs to the proper directory. For this example, I'm using a subdomain of example.com
with a letsencrypt certificate, sgt-demo.example.com.  Lets encrypt certs live in `/etc/letsencrypt/live/<site>`
so I'm copying them from there into the cert directory for SGT.
    ```commandline
    sudo cp /etc/letsencrypt/live/sgt-demo.example.com/fullchain.pem certs/fullchain.pem
    sudo cp /etc/letsencrypt/live/sgt-demo.example.com/privkey.pem certs/privkey.pem
    ```

5. Rename your certs to reflect which site they belong to.  I recommend following the example format of
    ```commandline
    example.domain.com.fullchain.pem
    ```
   moving...
   ```commandline
    cd certs
    mv fullchain.pem sgt-demo.example.com.fullchain.pem
    mv privkey.pem sgt-demo.example.com.privkey.pem
    cd ..
    ```

6. Create a new environment by following the prompts
    ```commandline
    ./sgt wizard
    ```
    6a. Enter a name for your environment (I'm calling my demo one sgt-demo)
    ```commandline
    Enter new environment name.  This is typically something like'Dev' or 'Prod' or 'Testing, but can be anything you want it to be: sgt-demo
    ```
    6b. Choose the AWS profile to use (Mine is again called sgt-demo)
    ```commandline
    Enter the name for the aws profile you'd like to use to deploy this environment
    if you've never created a profile before, you can read more about how to do this here
    http://docs.aws.amazon.com/cli/latest/userguide/cli-multiple-profiles.html
    a 'default' profile is created if you've installed and configured the aws cli:
    sgt-demo
    ```
    6c. Enter the IP address that you are currently deploying from.
    ```commandline
    Enter an ipaddress or cidr block for access to your elasticsearch cluster.
    Note:  This should probably be your current IP address, as you will need to be able to access
    elasticsearch via API to create the proper indices and mappings when deploying: xxx.xxx.xxx.xxx/24
    ```
    6d. Name your log bucket.  I recommend something easily identified for your domain.
    ```commandline
    Enter a name for the s3 bucket that will hold your osquery logs.
    Remeber, S3 bucket names must be globally unique: sgt-demo.log.bucket
    ```
    6e. And your config bucket...
    ```commandline
    Enter a name for the s3 bucket that will hold your server configuration
    Remember, S3 bucket names must be globally unique:
    sgt-demo.configuration.bucket
    ```
    6f. Enter your root domain
    ```commandline
    Enter the domain you will be using for your SGT server.
    Note:  This MUST be a domain which you have previously registered or are managing throughaws.
    This will be used to create a subdomain for the SGT TLS endpoint
    example.com
    ```
    6g. Enter the subdomain (sgt-demo in my case)
    ```commandline
    Enter a subdomain to use as the endpoint.  This will be prepended to the
    domain you provided as a subdomain
    sgt-demo
    ```
    6h. Enter your aws keypair name
    ```commandline
    Enter the name of your aws keypair.  This is used to access ec2 instances ifthe need
    should ever arise (it shouldn't).
    NOTE:  This is the name of the keypair EXCLUDING the .pem flie name and it must already exist in aws
    my-secret-key-name
    ```
    6i. Enter the name of your keypair and priv key, as you named them above.
    ```commandline
    Enter the name of the full ssl certificate chain bundle you will be using for
    your SGT server.  EG - full_chain.pem :
    sgt-demo.example.com.fullchain.pem
    Enter the name of the private key for your ssl certificate.  Eg - privkey.pem:
    sgt-demo.example.com.privkey.pem
    ```
    6j. Enter the node secret
    ```commandline
    Enter the node secret you will use to enroll your endpoints with the SGT server
    This secret will be used by each endpoint to authenticate to your server:
    my-super-secret-node-secret
    ```
    6k. Enter the app secret
    ```commandline
    Enter the app secret key which will be used to generate session tokens when
    interacting with the API as an authenticated end-user.  Make this long, random and complex:
    diu3piqeujr302348u33rqwu934r1@#)(*@3
    ```
    Select __N__ when prompted to continue.  Because this is a demo environment, we're going to make a small change to our configuration.

7. Edit the environment config file found in `/terraform/<environment/environment.json`
with your favorite editor and change the value for create_elasticsearch to `0`.  This will disable the creation of elasticsearch,
which we will not be using for this demo.  In a production environment, Elasticsearch would be a large part of your
process, but it adds significant cost and it's not needed for this demo.
    ```json
    {
      "environment": "example_environment",
      "aws_profile": "default",
      "user_ip_address": "127.0.0.1",
      "sgt_osquery_results_bucket_name": "example_log_bucket_name",
      "sgt_config_bucket_name": "example_config_bucket_name",
      "domain": "somedomain.com",
      "subdomain": "mysubdomain",
      "aws_keypair": "my_aws_ec2_keypair_name",
      "full_ssl_certchain": "full_cert_chain.pem",
      "ssl_private_key": "privkey.pem",
      "sgt_node_secret": "super_sekret_node_enrollment_key",
      "sgt_app_secret": "ultra_mega_sekret_key_you'll_never_give_to_anyone_not_even_your_mother",
      "create_elasticsearch": 0
    }
    ```

### Deploy!!

Its finally time to deploy, although hopefully that wasn't too painful.  Deployment is by far the easiest part.

```commandline
./sgt deploy -env <your environment name> -all
```

This will stand up the entire environment, including endpoint configuration scripts which we will use to set up some osquery nodes later.
The entire process should take about 5-10 minutes depending on your internet connection, at which point you should be ready to install osquery on
an endpoint and start receiving logs!


-Note:  This getting started guide originally appeared on blog.securelyinsecure.com, but I'm appropriating it for the docs as well, due to it being better than the last readme I wrote.



### Manual deployment

SGT can be deployed as a full environment, or individual pieces(Note that the components
still requires their dependencies to be built, they may just be updated individually to save time)

To deploy SGT...

```commandline
./sgt deploy -env <environment> -all
```

To deploy/update an individual component..

```commandline
./sgt deploy -env <environment> -components elasticsearch,firehose
```

for a full list of commands, issue the -h flag

If terraform fails at any point during this process, cancel the installation `ctrl+c` and review
your errors.  SGT depends on all previous deploy steps completing successfully, so it is important
to make sure this occurs before moving on to next steps

### Creating your first user.

To create a user to interact with SGT, pass the `-create_user` flag with the requisite options

```commandline
./sgt create-user -credentials-file <cred_file> -profile <profile> -username <username> -role <"Admin"|"User"|"Read-only">
```

### Getting an Authentication Token

Using any portion of the End-user facing API requires an Authorization token.  To get an auth token, send a post request to
`/api/v1/get-token` supplying your username and password in the post body
```json
{"username": "my_username", "password": "my_password"}
```

If your credentials are valid, you will recieve a json response back
```json
{"Authorization": "<long jtw">}
```

Provide this token in any subsequent requests in the Authorization header


