package deploy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/oktasecuritylabs/sgt/logger"
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

	logger.Info("Deploying: %s", DepOrder)

	for _, name := range DepOrder {
		if err := deployAWSComponent(name, environ); err != nil {
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

	return deployAWSComponent(component, envName)
}

// deployAWSComponent handles applying the aws components with terraform
// This includes: VPC, Datastore, Firehose, Autoscaling, Secrets
// S3 requires building the binary
// Elasticsearch requires creating the elasticsearch mappings
func deployAWSComponent(component, envName string) error {

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
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	args := fmt.Sprintf("terraform apply -auto-approve -var-file=../%s.json", envName)
	logger.Info(args)

	cmd = exec.Command("bash", "-c", args)
	stdoutStderr, err = cmd.CombinedOutput()

	logger.Info(string(stdoutStderr))

	if component == "elasticsearch" {
		time.Sleep(time.Second * 10)

		return createElasticSearchMappings()
	}

	return err
}

// createElasticSearchMappings creates Elasticsearch mappings
func createElasticSearchMappings() error {

	fn := "terraform.tfstate"
	file, err := os.Open(fn)
	if err != nil {
		fmt.Println(err)
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
					break
				}
			}
		}
	}

	logger.Info(esEndpoint)
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

	req, _ := http.NewRequest(http.MethodPut, uri, bytes.NewBuffer(rawJSON))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, _ := client.Do(req)
	fmt.Println(response.Status)
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(body))
	if response.Status != "200 OK" {
		return errors.New(string(body))
	}

	return nil
}
