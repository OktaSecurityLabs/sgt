package deploy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/oktasecuritylabs/sgt/handlers/helpers"
)

// Wizard walks through setting up an environment config and deploys
func Wizard() error {
	config := DeploymentConfig{}
	fmt.Print("Enter new environment name.  This is typically something like" +
		"'Dev' or 'Prod' or 'Testing, but can be anything you want it to be: ")
	//envName, err := reader.ReadString('\n')
	var envName string
	_, err := fmt.Scan(&envName)
	if err != nil {
		return err
	}
	err = CreateDeployDirectory(envName)
	if err != nil {
		return err
	}
	config.Environment = envName
	fmt.Println("Enter the name for the aws profile you'd like to use to deploy this environment \n" +
		"if you've never created a profile before, you can read more about how to do this here\n" +
		"http://docs.aws.amazon.com/cli/latest/userguide/cli-multiple-profiles.html \n" +
		"a 'default' profile is created if you've installed and configured the aws cli: ")
	var profile string
	_, err = fmt.Scan(&profile)
	if err != nil {
		return err
	}
	config.AWSProfile = profile
	fmt.Println("Enter an ipaddress or cidr block for access to your elasticsearch cluster. \n" +
		"Note:  This should probably be your current IP address, as you will need to be able to access \n" +
		"elasticsearch via API to create the proper indices and mappings when deploying: ")
	var ipAddress string
	_, err = fmt.Scan(&ipAddress)
	if err != nil {
		return err
	}
	config.UserIPAddress = ipAddress
	fmt.Println("Enter a name for the s3 bucket that will hold your osquery logs. \n" +
		"Remeber;  S3 bucket names must be globally unique: ")
	var logBucketName string
	_, err = fmt.Scan(&logBucketName)
	if err != nil {
		return err
	}
	config.SgtOsqueryResultsBucketName = logBucketName
	fmt.Println("Enter a name for the s3 bucket that will hold your server configuration \n" +
		"Remeber;  S3 bucket names must be globally unique: ")
	var configBucket string
	_, err = fmt.Scan(&configBucket)
	if err != nil {
		return err
	}
	config.SgtConfigBucketName = configBucket
	fmt.Println("Enter the domain you will be using for your SGT server. \n" +
		"Note:  This MUST be a domain which you have previously registered or are managing through" +
		"aws. \n  This will be used to create a subdomain for the SGT TLS endpoint")
	var domain string
	_, err = fmt.Scan(&domain)
	if err != nil {
		return err
	}
	config.Domain = domain
	fmt.Println("Enter a subdomain to use as the endpoint.  This will be prepended to the \n" +
		"domain you provided as a subdomain")
	var subdomain string
	_, err = fmt.Scan(&subdomain)
	if err != nil {
		return err
	}
	config.Subdomain = subdomain
	fmt.Println("Enter the name of your aws keypair.  This is used to access ec2 instances if" +
		"the need \n should ever arise (it shouldn't).\n" +
		"NOTE:  This is the name of the keypair EXCLUDING the .pem flie name and it must already exist in aws")
	var keypair string
	_, err = fmt.Scan(&keypair)
	if err != nil {
		return err
	}
	config.AwsKeypair = keypair
	fmt.Println("Enter the name of the full ssl certificate chain bundle you will be using for \n" +
		"your SGT server.  EG - full_chain.pem : ")
	var fullChain string
	_, err = fmt.Scan(&fullChain)
	if err != nil {
		return err
	}
	config.FullSslCertchain = fullChain
	fmt.Println("Enter the name of the private key for your ssl certificate.  Eg - privkey.pem: ")
	var privKey string
	_, err = fmt.Scan(&privKey)
	if err != nil {
		return err
	}
	config.SslPrivateKey = privKey
	fmt.Println("Enter the node secret you will use to enroll your endpoints with the SGT server\n" +
		"This secret will be used by each endpoint to authenticate to your server: ")
	var nodeSecret string
	_, err = fmt.Scan(&nodeSecret)
	if err != nil {
		return err
	}
	config.SgtNodeSecret = nodeSecret
	fmt.Println("Enter the app secret key which will be used to generate session tokens when \n" +
		"interacting with the API as an authenticated end-user.  Make this long, random and complex: ")
	var appSecret string
	_, err = fmt.Scan(&appSecret)
	if err != nil {
		return err
	}
	config.SgtAppSecret = appSecret
	fmt.Println("Congratulations, you've successfully configured your SGT deployment! \n" +
		"That wasn't so bad, was it? \n" +
		"You're now ready to do the actual deployment")
	fmt.Println("If you'd like to continue and do the actual deployment, you may continue by\n" +
		"entering 'Y' at the next prompt.  If you'd like to pause, don't worry! \n" +
		"next time you're ready to continue, just run ./sgt -deploy -env $your_deployment_name")
	d, err := json.Marshal(config)
	fn := fmt.Sprintf("terraform/%s/%s.json", envName, envName)
	err = ioutil.WriteFile(fn, d, 0644)
	if err != nil {
		return err
	}

	if helpers.ConfirmAction("Would you like to proceed with deployment?") {
		err = AllComponents(config, envName)
	}

	return err
}
