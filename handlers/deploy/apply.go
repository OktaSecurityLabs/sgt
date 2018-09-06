package deploy

import (
        "os/user"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
        "github.com/aws/aws-sdk-go/aws/credentials"
        "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/oktasecuritylabs/sgt/logger"
        "github.com/aws/aws-sdk-go/aws"
        "github.com/aws/aws-sdk-go/aws/session"
        "github.com/aws/aws-sdk-go/service/elasticsearchservice"
        "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

// tfState struct for terraform state
type tfState struct {
	Version          int        `json:"version"`
	TerraformVersion string     `json:"terraform_version"`
	Serial           int        `json:"serial"`
	Lineage          string     `json:"lineage"`
	Modules          []tfModule `json:"modules"`
}

// tfModule struct for terrform modules in tfstates
type tfModule struct {
	Path    []string            `json:"path"`
	Outputs map[string]tfOutput `json:"outputs"`
}

// tfOutput struct for terraform outputs from tfstate files
type tfOutput struct {
	Sensitive bool   `json:"sensitive"`
	Type      string `json:"type"`
	Value     string `json:"value"`
}

// copyComponentTemplates copies the examples to the new deploy env
func copyComponentTemplates(component, envName string) (string, error) {

	files, err := filepath.Glob(fmt.Sprintf("terraform/example/%s/*", component))
	if err != nil {
		return "", err
	}

	componentPath := fmt.Sprintf("terraform/%s/%s", envName, component)

	for _, fn := range files {
		_, filename := filepath.Split(fn)
		err = copyFile(fn, fmt.Sprintf("%s/%s", componentPath, filename))
		if err != nil {
			return "", err
		}
	}

	return componentPath, nil
}

// AllComponents deploys all components
func AllComponents(config DeploymentConfig, environ string) error {
	var DepOrder []string
	//handle teardown other firehose if exists.

	if config.CreateElasticsearch == 1 {
		//if err := destroyAWSComponent(firehose, environ); err != nil {
		//return err
		//}
		DepOrder = ElasticDeployOrder

	} else {
		/*if err := destroyAWSComponent(elasticsearchFirehose, environ); err != nil {
			return err
		}

		if err := destroyAWSComponent(elasticsearch, environ); err != nil {
			return err
		}
		*/
		DepOrder = DeployOrder
	}

	logger.Infof("Deploying: %s", DepOrder)

	for _, name := range DepOrder {
		if err := deployAWSComponent(name, environ, config); err != nil {
			logger.Error(err)
			return err
		}
	}

	for _, fn := range osqueryDeployCommands {
		if err := fn(config, environ); err != nil {
			logger.Error(err)
			return err
		}
	}

	return nil
}

// Component deploys both aws components and osquery components
func Component(config DeploymentConfig, component, envName string) error {
	if fn, ok := osqueryDeployCommands[component]; ok {
		return fn(config, envName)
	}

	return deployAWSComponent(component, envName, config)
}

// deployAWSComponent handles applying the aws components with terraform
// This includes: VPC, Datastore, Firehose, Autoscaling, Secrets
// S3 requires building the binary
// Elasticsearch requires creating the elasticsearch mappings
func deployAWSComponent(component, envName string, config DeploymentConfig) error {

	// Change back to the top level directory after each component deploy
	cachedCurDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(cachedCurDir)

	if component == "s3" {
		logger.Info("Making sure binary is built...")
		cmd := exec.Command("bash", "-c", "go build sgt.go")
		combinedOutput, buildErr := cmd.CombinedOutput()
		if buildErr != nil {
			return buildErr
		}
		logger.Info(string(combinedOutput))
	}

	logger.Infof("Building %s...\n", component)

	spin.Start()
	defer spin.Stop()

	componentPath, err := copyComponentTemplates(component, envName)
	if err != nil {
		return err
	}

	if err = os.Chdir(componentPath); err != nil {
		return err
	}

	if component == "firehose" || component == "elasticsearch_firehose" {
		logger.Info("updating zip file...")
		cmd := exec.Command("bash", "-c", "../../modules/firehose/build_lambda.sh")
		combinedOutput, buildErr := cmd.CombinedOutput()
		if buildErr != nil {
			return buildErr
		}
		logger.Info(string(combinedOutput))
	}

	cmd := exec.Command("terraform", "init")
	_, err = cmd.CombinedOutput()
	if err != nil {
		return err
	}

	args := fmt.Sprintf("terraform apply -auto-approve -var-file=../%s.json", envName)
	logger.Info(args)

	cmd = exec.Command("bash", "-c", args)
	stdoutStderr, err := cmd.CombinedOutput()

	logger.Info(string(stdoutStderr))

	if component == "elasticsearch" {
		time.Sleep(time.Second * 10)

		return createElasticSearchMappings(config)
	}

	return err
}

// createElasticSearchCognitoOptions creates CognitoOption settings for the Elasticsearch domain
func createElasticSearchCognitoOptions(currentRegion string, config DeploymentConfig) error {

        //Get input variables from terraform state
        var ESCognitoRoleArn string
        var CognitoUserPoolId string
        var CognitoIdentityPoolId string
        var ESCognitoDomainName string

        fn := "terraform.tfstate"
        file, err := os.Open(fn)
        if err != nil {
                return err
        }

        decoder := json.NewDecoder(file)

        tfstate := tfState{}

        if err = decoder.Decode(&tfstate); err != nil {
                return err
        }

        for _, i := range tfstate.Modules {
                for k, v := range i.Outputs {
                        if k == "elasticsearch_domain_id" {
                                if !strings.Contains(v.Value, "amazon.com") {
                                        ESCognitoDomainName = v.Value
                                        break
                                }
                        }
                }
        }
        logger.Info(ESCognitoDomainName)

        for _, i := range tfstate.Modules {
                for k, v := range i.Outputs {
                        if k == "elasticsearch_cognito_role_arn" {
                                if !strings.Contains(v.Value, "amazon.com") {
                                        ESCognitoRoleArn = v.Value
                                        break
                                }
                        }
                }
        }
        logger.Info(ESCognitoRoleArn)

        for _, i := range tfstate.Modules {
                for k, v := range i.Outputs {
                        if k == "cognito_user_pool_id" {
                                if !strings.Contains(v.Value, "amazon.com") {
                                        CognitoUserPoolId = v.Value
                                        break
                                }
                        }
                }
        }
        logger.Info(CognitoUserPoolId)

        for _, i := range tfstate.Modules {
                for k, v := range i.Outputs {
                        if k == "cognito_identity_pool_id" {
                                if !strings.Contains(v.Value, "amazon.com") {
                                        CognitoIdentityPoolId = v.Value
                                        break
                                }
                        }
                }
        }
        logger.Info(CognitoIdentityPoolId)

        sess := session.Must(session.NewSessionWithOptions(session.Options{
            Config: aws.Config{
            Region: aws.String(currentRegion),
            },
            Profile: *aws.String(config.AWSProfile),
        }))

        // Create a Elasticsearch service client.
        svc := elasticsearchservice.New(sess)

        //Get the elasticsearch domain config
        result, err := svc.DescribeElasticsearchDomainConfig(&elasticsearchservice.DescribeElasticsearchDomainConfigInput{
                DomainName: aws.String(ESCognitoDomainName),
        })

        if err != nil {
            return err
        }

        ESCognitoOptionsEnabledStatus := *result.DomainConfig.CognitoOptions.Options.Enabled

        logger.Info(ESCognitoOptionsEnabledStatus)

 
        //If cognito options are not set for the domain, set them
        if !ESCognitoOptionsEnabledStatus {
            svc.UpdateElasticsearchDomainConfig(&elasticsearchservice.UpdateElasticsearchDomainConfigInput{
                DomainName: aws.String(ESCognitoDomainName),
                CognitoOptions: &elasticsearchservice.CognitoOptions{
                    Enabled: aws.Bool(true),
                    UserPoolId: aws.String(CognitoUserPoolId),
                    IdentityPoolId: aws.String(CognitoIdentityPoolId),
                    RoleArn: aws.String(ESCognitoRoleArn),
                },
            })
        }

        cognito_svc := cognitoidentityprovider.New(sess)

        UserExists := false
        //List of desired users for Kibana, from the config file, should be the first part of Okta email address to work correctly
        DesiredUsers := config.Users
        MailDomain := config.MailDomain        

        ExistingUsers, err := cognito_svc.ListUsers(&cognitoidentityprovider.ListUsersInput{
                UserPoolId: aws.String(CognitoUserPoolId),
        })

        if err != nil {
            logger.Info(err)
        }        

        //Determine if the user already exists
        for _, DesiredUser := range DesiredUsers {
            for _, ExistingUser := range ExistingUsers.Users{
                UserExists = false
                if DesiredUser == *ExistingUser.Username {
                    UserExists = true
                }               
            }
            if UserExists == false {
                logger.Info(DesiredUser)
                //If not, create it
                createuser, err := cognito_svc.AdminCreateUser(&cognitoidentityprovider.AdminCreateUserInput{
                    UserPoolId: aws.String(CognitoUserPoolId),
                    Username: aws.String(DesiredUser),
                    UserAttributes: []*cognitoidentityprovider.AttributeType{
                        {Name: aws.String("email"),Value : aws.String(DesiredUser+"@"+MailDomain)},
                    },
                })
                if err != nil {
                    logger.Info(err)
                }
                logger.Info(createuser)
            }
        }
        

        if err != nil {
            logger.Info(err)
        }

        return nil
}

// createElasticSearchMappings creates Elasticsearch mappings
func createElasticSearchMappings(config DeploymentConfig) error {
        usr, err := user.Current()
        if err != nil {
                logger.Error(err)
                return err
        }

        credfile := filepath.Join(usr.HomeDir, ".aws", "credentials")
        creds := credentials.NewSharedCredentials(credfile, config.AWSProfile)
        now := time.Now()
        signer := v4.NewSigner(creds)

	fn := "terraform.tfstate"
	file, err := os.Open(fn)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(file)
	tfstate := tfState{}
	if err = decoder.Decode(&tfstate); err != nil {
		return err
	}

	var esEndpoint string
	for _, i := range tfstate.Modules {
		for k, v := range i.Outputs {
			if k == "elasticearch_endpoint" {
				if !strings.Contains(v.Value, "amazon.com") {
					esEndpoint = v.Value
                                        logger.Info(esEndpoint)
					break
				}
			}
		}
	}

        var currentRegion string
        for _, i := range tfstate.Modules {
                for k, v := range i.Outputs {
                        if k == "elasticsearch_region" {
                                if !strings.Contains(v.Value, "amazon.com") {
                                        currentRegion = v.Value
                                        logger.Info(currentRegion)
                                        break
                                }
                        }
                }
        }

	//esEndpoint := "https://search-sgt-osquery-results-r6owrsyarql42ttzy26fz6nf24.us-east-1.es.amazonaws.com"
	path := "_template/template_1"
	rawJSON := json.RawMessage(`{
  "template": "osquery_*",
  "settings": {
    "number_of_shards": 4
  },
  "mappings": {
    "_default_": {
      "properties": {
        "calendarTime": {
          "type": "date",
          "format": "yyyy-MM-dd HH:mm:ss||yyyy-MM-dd||epoch_millis"
        },
        "unixTime": {
          "type": "date"
        }
      }
    }
  }
}`)

	uri := fmt.Sprintf("https://%s/%s", esEndpoint, path)
	logger.Info(uri)

	req, err := http.NewRequest(http.MethodPut, uri, bytes.NewBuffer(rawJSON))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
        signer.Sign(req, bytes.NewReader(rawJSON), "es", currentRegion, now)
	response, err := client.Do(req)
	if err != nil {
		return err
	}

	logger.Info(response.Status)

	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return err
	}

	logger.Info(string(body))
	if response.Status != "200 OK" {
		return fmt.Errorf("Request failed: %s", string(body))
	}

        createElasticSearchCognitoOptions(currentRegion, config)

	return nil
}
