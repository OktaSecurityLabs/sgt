package deploy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/oktasecuritylabs/sgt/dyndb"
	"github.com/oktasecuritylabs/sgt/handlers/auth"
	"github.com/oktasecuritylabs/sgt/handlers/helpers"
	"github.com/oktasecuritylabs/sgt/logger"
	osq_types "github.com/oktasecuritylabs/sgt/osquery_types"
)

type DeploymentConfig struct {
	Environment                 string `json:"environment"`
	AWSProfile                  string `json:"aws_profile"`
	UserIPAddress               string `json:"user_ip_address"`
	SgtOsqueryResultsBucketName string `json:"sgt_osquery_results_bucket_name"`
	SgtConfigBucketName         string `json:"sgt_config_bucket_name"`
	Domain                      string `json:"domain"`
	Subdomain                   string `json:"subdomain"`
	AwsKeypair                  string `json:"aws_keypair"`
	FullSslCertchain            string `json:"full_ssl_certchain"`
	SslPrivateKey               string `json:"ssl_private_key"`
	SgtNodeSecret               string `json:"sgt_node_secret"`
	SgtAppSecret                string `json:"sgt_app_secret"`
}

func CopyFile(src, dst string) error {
	// ripped from https://stackoverflow.com/questions/21060945/simple-way-to-copy-a-file-in-golang
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func ErrorCheck(err error) error {
	if err != nil {
		logger.Error(err)
		logger.Fatal(err)
		return err
	}
	return nil
}

func ParseDeploymentConfig(config_file string) (DeploymentConfig, error) {
	dep_conf := DeploymentConfig{}
	file, err := os.Open(config_file)
	if err != nil {
		logger.Warn(err)
		logger.Fatal(err)
		return dep_conf, err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&dep_conf)
	if err != nil {
		logger.Warn(err)
		logger.Fatal(err)
		return dep_conf, err
	}
	return dep_conf, nil
}

func CreateDeployDirectory(environ string) error {
	path := fmt.Sprintf("terraform/%s", environ)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logger.Info(fmt.Sprintf("creating new deployment environment: %s", environ))
		os.Mkdir(path, 0755)
	} else {
		logger.Info("environment already exists, are you sure you meant to to use deploy to\n")
		logger.Info(environ)
		os.Exit(0)
	}
	dirs := []string{"vpc", "datastore", "firehose", "elasticsearch", "s3", "autoscaling", "secrets"}
	for _, p := range dirs {
		dir := filepath.Join(path, p)
		//logger.Info(dir)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			logger.Info(fmt.Sprintf("Creating %s directory", dir))
			os.Mkdir(dir, 0755)
		}
	}
	return nil
}

func CheckEnvironMatchConfig(environ, config_file string) error {
	dep_conf, err := ParseDeploymentConfig(config_file)
	if err != nil {
		logger.Fatal(err)
		return err
	}
	if dep_conf.Environment != environ {
		return errors.New("config environment and passed environment variable do not match")
	}
	return nil
}

func VPC(top_level_dir, environ string) error {
	config_file := fmt.Sprintf("terraform/%s/%s.json", environ, environ)
	logger.Info("Building VPC....")
	//check to make sure terraform files are in place
	_, err := ParseDeploymentConfig(config_file)
	files, err := filepath.Glob("terraform/example/vpc/*")
	for _, fn := range files {
		_, filename := filepath.Split(fn)
		//err = CopyFile(fn, fmt.Sprintf("terraform/%s/vpc/%", environ, filename))
		err = CopyFile(fn, fmt.Sprintf("terraform/%s/vpc/%s", environ, filename))
		if err != nil {
			logger.Error(err)
		}
	}
	err = CheckEnvironMatchConfig(environ, config_file)
	ErrorCheck(err)
	err = os.Chdir(fmt.Sprintf("terraform/%s/vpc", environ))
	ErrorCheck(err)
	cmd := exec.Command("terraform", "init")
	stdoutStderr, err := cmd.CombinedOutput()
	if err = ErrorCheck(err); err != nil {
		return err
	}
	//args := fmt.Sprintf("terraform apply -var aws_profile=%s", config.AWSProfile)
	args := fmt.Sprintf("terraform apply -var-file=../%s.json", environ)
	logger.Info(args)
	s := spinner.New(spinner.CharSets[43], time.Millisecond*500)
	s.Start()
	cmd = exec.Command("bash", "-c", args)
	stdoutStderr, err = cmd.CombinedOutput()
	logger.Info(string(stdoutStderr))
	s.Stop()
	err = os.Chdir(top_level_dir)
	ErrorCheck(err)
	s.Stop()
	return nil
}

func Datastore(top_level_dir, environ string) error {
	config_file := fmt.Sprintf("terraform/%s/%s.json", environ, environ)
	logger.Info("building Datastore...")
	//s := spinner.New(spinner.CharSets[0], 500*time.Millisecond)
	//s.Start()
	//check to make sure terraform files are in place
	_, err := ParseDeploymentConfig(config_file)
	files, err := filepath.Glob("terraform/example/datastore/*")
	for _, fn := range files {
		_, filename := filepath.Split(fn)
		//err = CopyFile(fn, fmt.Sprintf("terraform/%s/vpc/%", environ, filename))
		err = CopyFile(fn, fmt.Sprintf("terraform/%s/datastore/%s", environ, filename))
		if err != nil {
			logger.Error(err)
		}
	}
	err = CheckEnvironMatchConfig(environ, config_file)
	ErrorCheck(err)
	err = os.Chdir(fmt.Sprintf("terraform/%s/datastore", environ))
	ErrorCheck(err)
	cmd := exec.Command("terraform", "init")
	stdoutStderr, err := cmd.CombinedOutput()
	//args := fmt.Sprintf("terraform apply -var aws_profile=%s", config.AWSProfile)
	args := fmt.Sprintf("terraform apply -var-file=../%s.json", environ)
	logger.Info(args)
	cmd = exec.Command("bash", "-c", args)
	stdoutStderr, err = cmd.CombinedOutput()
	logger.Info(string(stdoutStderr))
	err = os.Chdir(top_level_dir)
	ErrorCheck(err)
	//s.Stop()
	return nil
}

func ElasticSearchMappings(top_level_dir, environ string) error {
	err := os.Chdir(fmt.Sprintf("terraform/%s/elasticsearch", environ))
	ErrorCheck(err)
	fn := "terraform.tfstate"
	file, err := os.Open(fn)
	if err != nil {
		fmt.Println(err)
	}
	decoder := json.NewDecoder(file)
	tfstate := TState{}
	err = decoder.Decode(&tfstate)
	if err != nil {
		logger.Error(err)
	}
	//fmt.Printf("%+v", tfstate)
	//fmt.Printf("%+v", tfstate.Modules[0].Outputs.(map))
	//es_endpoint := tfstate.Modules[0].Outputs["elasticsearch_endpoint"]
	es_endpoint := ""
	for _, i := range tfstate.Modules {
		//fmt.Printf("%+v", i)
		for k, v := range i.Outputs {
			if k == "elasticearch_endpoint" {
				if !strings.Contains(v.Value, "amazon.com") {
					//fmt.Println(v.Value)
					es_endpoint = v.Value
				}
			}
		}
	}
	logger.Info(es_endpoint)
	//es_endpoint := "https://search-sgt-osquery-results-r6owrsyarql42ttzy26fz6nf24.us-east-1.es.amazonaws.com"
	path := "_template/template_1"
	raw_json := json.RawMessage(`{
	  "template": "osquery_*",
	  "settings": {
		"number_of_shards": 4},
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

	uri := fmt.Sprintf("https://%s/%s", es_endpoint, path)
	logger.Info(uri)
	//js, err := json.Marshal(raw_json)
	req, _ := http.NewRequest("PUT", uri, bytes.NewBuffer(raw_json))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, _ := client.Do(req)
	fmt.Println(response.Status)
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(body))
	if response.Status != "200 OK" {
		return errors.New(string(body))
	}
	err = os.Chdir(top_level_dir)
	return nil
}

func Elasticsearch(top_level_dir, environ string) error {
	config_file := fmt.Sprintf("terraform/%s/%s.json", environ, environ)
	logger.Info("building Elasticsearch...")
	logger.Info("Note:  Due to the way Amazon's elasticsearch service is built, this may take up to 30 minutes or more to complete")
	//check to make sure terraform files are in place
	_, err := ParseDeploymentConfig(config_file)
	files, err := filepath.Glob("terraform/example/elasticsearch/*")
	for _, fn := range files {
		_, filename := filepath.Split(fn)
		//err = CopyFile(fn, fmt.Sprintf("terraform/%s/vpc/%", environ, filename))
		err = CopyFile(fn, fmt.Sprintf("terraform/%s/elasticsearch/%s", environ, filename))
		if err != nil {
			logger.Error(err)
		}
	}
	err = CheckEnvironMatchConfig(environ, config_file)
	ErrorCheck(err)
	err = os.Chdir(fmt.Sprintf("terraform/%s/elasticsearch", environ))
	ErrorCheck(err)
	cmd := exec.Command("terraform", "init")
	stdoutStderr, err := cmd.CombinedOutput()
	if err = ErrorCheck(err); err != nil {
		return err
	}
	//args := fmt.Sprintf("terraform apply -var aws_profile=%s -var user_ip_address=%s", config.AWSProfile, config.UserIPAddress)
	args := fmt.Sprintf("terraform apply -var-file=../%s.json", environ)
	logger.Info(args)
	cmd = exec.Command("bash", "-c", args)
	s := spinner.New(spinner.CharSets[43], 500*time.Millisecond)
	s.Start()
	stdoutStderr, err = cmd.CombinedOutput()
	logger.Info(string(stdoutStderr))
	s.Stop()
	err = os.Chdir(top_level_dir)
	ErrorCheck(err)
	time.Sleep(time.Second * 10)
	err = ElasticSearchMappings(top_level_dir, environ)
	if err != nil {
		logger.Error(err)
		return err
	}
	//ElasticSearchMappings("https://search-sgt-osquery-results-r6owrsyarql42ttzy26fz6nf24.us-east-1.es.amazonaws.com")
	return nil
}

func Firehose(top_level_dir, environ string) error {
	config_file := fmt.Sprintf("terraform/%s/%s.json", environ, environ)
	logger.Info("building Firehose(n)...")
	//check to make sure terraform files are in place
	_, err := ParseDeploymentConfig(config_file)
	files, err := filepath.Glob("terraform/example/firehose/*")
	for _, fn := range files {
		_, filename := filepath.Split(fn)
		//err = CopyFile(fn, fmt.Sprintf("terraform/%s/vpc/%", environ, filename))
		err = CopyFile(fn, fmt.Sprintf("terraform/%s/firehose/%s", environ, filename))
		if err != nil {
			logger.Error(err)
		}
	}
	err = CheckEnvironMatchConfig(environ, config_file)
	ErrorCheck(err)
	err = os.Chdir(fmt.Sprintf("terraform/%s/firehose", environ))
	ErrorCheck(err)
	cmd := exec.Command("terraform", "init")
	stdoutStderr, err := cmd.CombinedOutput()
	if err = ErrorCheck(err); err != nil {
		return err
	}
	//args := fmt.Sprintf("terraform apply -var aws_profile=%s -var s3_bucket_name=%s", config.AWSProfile, config.LogBucketName)
	args := fmt.Sprintf("terraform apply -var-file=../%s.json", environ)
	logger.Info(args)
	s := spinner.New(spinner.CharSets[43], time.Millisecond*500)
	s.Start()
	cmd = exec.Command("bash", "-c", args)
	stdoutStderr, err = cmd.CombinedOutput()
	s.Stop()
	logger.Info(string(stdoutStderr))
	err = os.Chdir(top_level_dir)
	ErrorCheck(err)
	return nil
}

func S3(top_level_dir, environ string) error {
	config_file := fmt.Sprintf("terraform/%s/%s.json", environ, environ)
	logger.Info("Making sure binary is built...")
	cmd := exec.Command("bash", "-c", "go build sgt.go")
	combinedoutput, err := cmd.CombinedOutput()
	ErrorCheck(err)
	logger.Info(combinedoutput)
	logger.Info("building S3...")
	//check to make sure terraform files are in place
	_, err = ParseDeploymentConfig(config_file)
	files, err := filepath.Glob("terraform/example/s3/*")
	for _, fn := range files {
		_, filename := filepath.Split(fn)
		err = CopyFile(fn, fmt.Sprintf("terraform/%s/s3/%s", environ, filename))
		if err != nil {
			logger.Error(err)
		}
	}
	err = CheckEnvironMatchConfig(environ, config_file)
	ErrorCheck(err)
	err = os.Chdir(fmt.Sprintf("terraform/%s/s3", environ))
	ErrorCheck(err)
	cmd = exec.Command("bash", "-c", "terraform init")
	stdoutStderr, err := cmd.CombinedOutput()
	//args := fmt.Sprintf("terraform apply -var aws_profile=%s -var sgt_config_bucket=%s -var full_cert_chain=%s -var priv_key=%s",
	//config.AWSProfile, config.ConfigBucketName, config.SslFullKeychain, config.SslPrivateKey)
	args := fmt.Sprintf("terraform apply -var-file=../%s.json", environ)
	logger.Info(args)
	cmd = exec.Command("bash", "-c", args)
	stdoutStderr, err = cmd.CombinedOutput()
	logger.Info(string(stdoutStderr))
	err = os.Chdir(top_level_dir)
	ErrorCheck(err)
	return nil
}

func Secrets(top_level_dir, environ string) error {
	config_file := fmt.Sprintf("terraform/%s/%s.json", environ, environ)
	logger.Info("Uploading secrets...")
	//check to make sure terraform files are in place
	//_, err := ParseDeploymentConfig(config_file)
	files, err := filepath.Glob("terraform/example/secrets/*")
	for _, fn := range files {
		_, filename := filepath.Split(fn)
		err = CopyFile(fn, fmt.Sprintf("terraform/%s/secrets/%s", environ, filename))
		if err != nil {
			logger.Error(err)
		}
	}
	err = CheckEnvironMatchConfig(environ, config_file)
	ErrorCheck(err)
	err = os.Chdir(fmt.Sprintf("terraform/%s/secrets", environ))
	ErrorCheck(err)
	cmd := exec.Command("bash", "-c", "terraform init")
	stdoutStderr, err := cmd.CombinedOutput()
	args := fmt.Sprintf("terraform apply -var-file=../%s.json", environ)
	fmt.Println(args)
	//config.AWSProfile, config.NodeSecret, config.AppSecret)
	logger.Info("Args hidden due to sensitive nature...")
	cmd = exec.Command("bash", "-c", args)
	stdoutStderr, err = cmd.CombinedOutput()
	logger.Info(string(stdoutStderr))
	err = os.Chdir(top_level_dir)
	ErrorCheck(err)
	return nil
}

func Autoscaling(top_level_dir, environ string) error {
	config_file := fmt.Sprintf("terraform/%s/%s.json", environ, environ)
	logger.Info("building Autoscaling...")
	//check to make sure terraform files are in place
	//_, err := ParseDeploymentConfig(config_file)
	files, err := filepath.Glob("terraform/example/autoscaling/*")
	for _, fn := range files {
		_, filename := filepath.Split(fn)
		//err = CopyFile(fn, fmt.Sprintf("terraform/%s/vpc/%", environ, filename))
		err = CopyFile(fn, fmt.Sprintf("terraform/%s/autoscaling/%s", environ, filename))
		if err != nil {
			logger.Error(err)
		}
	}
	err = CheckEnvironMatchConfig(environ, config_file)
	ErrorCheck(err)
	err = os.Chdir(fmt.Sprintf("terraform/%s/autoscaling", environ))
	ErrorCheck(err)
	cmd := exec.Command("terraform", "init")
	stdoutStderr, err := cmd.CombinedOutput()
	if err = ErrorCheck(err); err != nil {
		return err
	}
	//args := fmt.Sprintf("terraform apply -var aws_profile=%s -var domain=%s -var subdomain=%s -var keypair=%s", config.AWSProfile, config.DomainName, config.Subdomain, config.AwsKeypair)
	args := fmt.Sprintf("terraform apply -var-file=../%s.json", environ)
	logger.Info(args)
	cmd = exec.Command("bash", "-c", args)
	s := spinner.New(spinner.CharSets[43], time.Millisecond*500)
	s.Start()
	stdoutStderr, err = cmd.CombinedOutput()
	s.Stop()
	logger.Info(string(stdoutStderr))
	err = os.Chdir(top_level_dir)
	ErrorCheck(err)
	return nil
}

func DestroyAutoscaling(top_level_dir, environ string) error {
	config_file := fmt.Sprintf("terraform/%s/%s.json", environ, environ)
	logger.Info("Destroying Autoscaling...")
	//check to make sure terraform files are in place
	_, err := ParseDeploymentConfig(config_file)
	err = CheckEnvironMatchConfig(environ, config_file)
	ErrorCheck(err)
	err = os.Chdir(fmt.Sprintf("terraform/%s/autoscaling", environ))
	ErrorCheck(err)
	if err = ErrorCheck(err); err != nil {
		return err
	}
	//args := fmt.Sprintf("terraform destroy -force -var aws_profile=%s -var domain=%s -var subdomain=%s -var keypair=%s", config.AWSProfile, config.DomainName, config.Subdomain, config.AwsKeypair)
	args := fmt.Sprintf("terraform destroy -force -var-file=../%s.json", environ)
	logger.Info(args)
	cmd := exec.Command("bash", "-c", args)
	s := spinner.New(spinner.CharSets[43], time.Millisecond*500)
	s.Start()
	stdoutStderr, err := cmd.CombinedOutput()
	s.Stop()
	logger.Info(string(stdoutStderr))
	err = os.Chdir(top_level_dir)
	ErrorCheck(err)
	return nil
}

func DestroySecrets(top_level_dir, environ string) error {
	config_file := fmt.Sprintf("terraform/%s/%s.json", environ, environ)
	logger.Info("Destroying secrets...")
	//check to make sure terraform files are in place
	_, err := ParseDeploymentConfig(config_file)
	err = CheckEnvironMatchConfig(environ, config_file)
	ErrorCheck(err)
	err = os.Chdir(fmt.Sprintf("terraform/%s/secrets", environ))
	ErrorCheck(err)
	//args := fmt.Sprintf("terraform destroy -var aws_profile=%s -var sgt_node_secret=%s -var sgt_app_secret=%s",
	//config.AWSProfile, config.NodeSecret, config.AppSecret)
	args := fmt.Sprintf("terraform destroy -force -var-file=../%s.json", environ)
	logger.Info("Args hidden due to sensitive nature...")
	cmd := exec.Command("bash", "-c", args)
	stdoutStderr, err := cmd.CombinedOutput()
	logger.Info(string(stdoutStderr))
	err = os.Chdir(top_level_dir)
	ErrorCheck(err)
	return nil
}

func DestroyS3(top_level_dir, environ string) error {
	config_file := fmt.Sprintf("terraform/%s/%s.json", environ, environ)
	logger.Info("Destroy S3...")
	//check to make sure terraform files are in place
	_, err := ParseDeploymentConfig(config_file)
	err = CheckEnvironMatchConfig(environ, config_file)
	ErrorCheck(err)
	err = os.Chdir(fmt.Sprintf("terraform/%s/s3", environ))
	ErrorCheck(err)
	//args := fmt.Sprintf("terraform destroy -force -var aws_profile=%s -var sgt_config_bucket=%s -var full_cert_chain=%s -var priv_key=%s",
	//config.AWSProfile, config.ConfigBucketName, config.SslFullKeychain, config.SslPrivateKey)
	args := fmt.Sprintf("terraform destroy -force -var-file=../%s.json", environ)
	logger.Info(args)
	cmd := exec.Command("bash", "-c", args)
	stdoutStderr, err := cmd.CombinedOutput()
	logger.Info(string(stdoutStderr))
	err = os.Chdir(top_level_dir)
	ErrorCheck(err)
	return nil
}

func DestroyFirehose(top_level_dir, environ string) error {
	config_file := fmt.Sprintf("terraform/%s/%s.json", environ, environ)
	logger.Info("Destrying Firehose(n)...")
	//check to make sure terraform files are in place
	_, err := ParseDeploymentConfig(config_file)
	err = CheckEnvironMatchConfig(environ, config_file)
	ErrorCheck(err)
	err = os.Chdir(fmt.Sprintf("terraform/%s/firehose", environ))
	ErrorCheck(err)
	//args := fmt.Sprintf("terraform destroy -force -var aws_profile=%s -var s3_bucket_name=%s", config.AWSProfile, config.LogBucketName)
	args := fmt.Sprintf("terraform destroy -force -var-file=../%s.json", environ)
	logger.Info(args)
	cmd := exec.Command("bash", "-c", args)
	s := spinner.New(spinner.CharSets[43], time.Millisecond*500)
	stdoutStderr, err := cmd.CombinedOutput()
	s.Stop()
	logger.Info(string(stdoutStderr))
	err = os.Chdir(top_level_dir)
	ErrorCheck(err)
	return nil
}

func DestroyElasticsearch(top_level_dir, environ string) error {
	config_file := fmt.Sprintf("terraform/%s/%s.json", environ, environ)
	logger.Info("Destroying Elasticsearch...")
	logger.Info("Note:  Due to the way Amazon's elasticsearch service is built, this may take up to 30 minutes or more to complete")
	logger.Info("PS.  Now is probably a good time for some coffee...mmm, coffee")
	//check to make sure terraform files are in place
	_, err := ParseDeploymentConfig(config_file)
	err = CheckEnvironMatchConfig(environ, config_file)
	ErrorCheck(err)
	err = os.Chdir(fmt.Sprintf("terraform/%s/elasticsearch", environ))
	ErrorCheck(err)
	//args := fmt.Sprintf("terraform destroy -force -var aws_profile=%s -var user_ip_address=%s", config.AWSProfile, config.UserIPAddress)
	args := fmt.Sprintf("terraform destroy -force -var-file=../%s.json", environ)
	logger.Info(args)
	cmd := exec.Command("bash", "-c", args)
	s := spinner.New(spinner.CharSets[43], 500*time.Millisecond)
	s.Start()
	stdoutStderr, err := cmd.CombinedOutput()
	s.Stop()
	logger.Info(string(stdoutStderr))
	err = os.Chdir(top_level_dir)
	ErrorCheck(err)
	return nil
}

func DestroyDatastore(top_level_dir, environ string) error {
	config_file := fmt.Sprintf("terraform/%s/%s.json", environ, environ)
	logger.Info("Destroying Datastore...")
	//s := spinner.New(spinner.CharSets[0], 500*time.Millisecond)
	//s.Start()
	//check to make sure terraform files are in place
	_, err := ParseDeploymentConfig(config_file)
	err = CheckEnvironMatchConfig(environ, config_file)
	ErrorCheck(err)
	err = os.Chdir(fmt.Sprintf("terraform/%s/datastore", environ))
	ErrorCheck(err)
	//args := fmt.Sprintf("terraform destroy -force -var aws_profile=%s", config.AWSProfile)
	args := fmt.Sprintf("terraform destroy -force -var-file=../%s.json", environ)
	logger.Info(args)
	cmd := exec.Command("bash", "-c", args)
	stdoutStderr, err := cmd.CombinedOutput()
	logger.Info(string(stdoutStderr))
	err = os.Chdir(top_level_dir)
	ErrorCheck(err)
	//s.Stop()
	return nil
}

func DestroyVPC(top_level_dir, environ string) error {
	config_file := fmt.Sprintf("terraform/%s/%s.json", environ, environ)
	logger.Info("Destroying VPC....")
	//check to make sure terraform files are in place
	_, err := ParseDeploymentConfig(config_file)
	err = CheckEnvironMatchConfig(environ, config_file)
	ErrorCheck(err)
	err = os.Chdir(fmt.Sprintf("terraform/%s/vpc", environ))
	ErrorCheck(err)
	//args := fmt.Sprintf("terraform destroy -force -var aws_profile=%s", config.AWSProfile)
	args := fmt.Sprintf("terraform destroy -force -var-file=../%s.json", environ)
	logger.Info(args)
	cmd := exec.Command("bash", "-c", args)
	s := spinner.New(spinner.CharSets[43], time.Millisecond*500)
	s.Start()
	stdoutStderr, err := cmd.CombinedOutput()
	s.Stop()
	logger.Info(string(stdoutStderr))
	err = os.Chdir(top_level_dir)
	ErrorCheck(err)
	return nil
}

func DeployAll(top_level_dir, environ string) error {
	err := VPC(top_level_dir, environ)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	err = Datastore(top_level_dir, environ)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	err = Elasticsearch(top_level_dir, environ)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	err = Firehose(top_level_dir, environ)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	err = S3(top_level_dir, environ)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	err = Secrets(top_level_dir, environ)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	err = Autoscaling(top_level_dir, environ)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	err = DeployDefaultPacks(environ)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	err = DeployDefaultConfigs(environ)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	err = GenerateEndpointDeployScripts(environ)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	return nil
}

func DeployWizard() error {
	config := DeploymentConfig{}
	fmt.Print("Enter new environment name.  This is typically something like" +
		"'Dev' or 'Prod' or 'Testing, but can be anything you want it to be: ")
	//env_name, err := reader.ReadString('\n')
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
	fmt.Scan(&logBucketName)
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
	config.SslPrivateKey = priv_key
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
	fmt.Println("Would you like to proceed with deployment? Y/N")
	d, err := json.Marshal(config)
	fn := fmt.Sprintf("terraform/%s/%s.json", env_name, env_name)
	err = ioutil.WriteFile(fn, d, 0644)
	var ans string
	_, err = fmt.Scan(&ans)
	confirm_strings := []string{"y", "Y", "yes", "YES"}
	var deploy bool
	for _, i := range confirm_strings {
		if strings.Contains(i, ans) {
			deploy = true
			break
		}
	}
	if deploy {
		curdir, err := os.Getwd()
		if err != nil {
			logger.Error(err)
			return err
		}
		err = DeployAll(curdir, config.Environment)
		if err != nil {
			logger.Error(err)
			return err
		}
	}
	return nil
}

func UserAwsCredFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		logger.Error(err)
		return "", err
	}
	credfile := filepath.Join(usr.HomeDir, ".aws", "credentials")
	return credfile, nil
}

type TState struct {
	Version          int        `json:"version"`
	TerraformVersion string     `json:"terraform_version"`
	Serial           int        `json:"serial"`
	Lineage          string     `json:"lineage"`
	Modules          []TFModule `json:"modules"`
}

type TFModule struct {
	Path    []string            `json:"path"`
	Outputs map[string]TFOutput `json:"outputs"`
}

type TFOutput struct {
	Sensitive bool   `json:"sensitive"`
	Type      string `json:"type"`
	Value     string `json:"value"`
}

func DeployDefaultPacks(environ string) error {
	var files []string
	config_file := fmt.Sprintf("terraform/%s/%s.json", environ, environ)
	config, err := ParseDeploymentConfig(config_file)
	if err != nil {
		logger.Error(err)
		return err
	}
	//if environ specific dir exists in packs, deploy those.  Otherwise use defaults
	if _, err := os.Stat(filepath.Join("packs", environ)); os.IsNotExist(err) {
		logger.Info(fmt.Sprintf("No environment specific packs found for: %s", environ))
		logger.Info("using default packs")
		files, err = filepath.Glob("packs/*")
		if err != nil {
			logger.Error(err)
			return err
		}
	} else {
		logger.Info(fmt.Sprintf("Environment specific folder found for: %s \nUsing %s query packs", environ, environ))
		path := fmt.Sprintf("packs/%s/*", environ)
		files, err = filepath.Glob(path)
		if err != nil {
			logger.Error(err)
			return err
		}
	}
	s := spinner.New(spinner.CharSets[43], time.Millisecond*500)
	s.Start()
	for _, fn := range files {
		_, filename := filepath.Split(fn)
		if strings.HasSuffix(filename, "json") {
			pack := osq_types.QueryPack{}
			helper_pack := helpers.OsqueryPack{}
			s, err := helpers.CleanPack(filename)
			if err != nil {
				logger.Error(err)
			}
			file := strings.NewReader(s)
			if err != nil {
				logger.Error(err)
				return err
			}
			decoder := json.NewDecoder(file)
			err = decoder.Decode(&helper_pack)
			if err != nil {
				logger.Error(err)
				return err
			}
			credfile, err := UserAwsCredFile()
			if err != nil {
				logger.Error(err)
				return err
			}
			dyn_svc := auth.CrendentialedDbInstance(credfile, config.AWSProfile)
			mu := sync.Mutex{}
			for k, v := range helper_pack.Queries {
				pq := osq_types.PackQuery{}
				pq.QueryName = k
				pq.Query = v.Query
				pq.Value = v.Value
				pq.Description = v.Description
				pq.Interval = v.Interval
				pq.Version = v.Version
				dyndb.UpsertPackQuery(pq, dyn_svc, mu)
			}
			pack.Queries = helper_pack.ListQueries()
			pack.PackName = strings.Split(filename, ".")[0]
			err = dyndb.UpsertPack(pack, dyn_svc, mu)
			if err != nil {
				logger.Error(err)
				return err
			}
		}
	}
	s.Stop()
	return nil
}

func DeployDefaultConfigs(environ string) error {
	var files []string
	config_file := fmt.Sprintf("terraform/%s/%s.json", environ, environ)
	config, err := ParseDeploymentConfig(config_file)
	if err != nil {
		logger.Error(err)
		return err
	}
	//if environ specific dir exists in packs, deploy those.  Otherwise use defaults
	env_specific_configs := false
	if _, err := os.Stat(filepath.Join("osquery_configs", environ)); os.IsNotExist(err) {
		logger.Info(fmt.Sprintf("No environment specific configs found for: %s", environ))
		logger.Info("using default configs")
		files, err = filepath.Glob("osquery_configs/defaults/*")
		environ = "defaults"
		env_specific_configs = true
		if err != nil {
			logger.Error(err)
			return err
		}
	} else {
		logger.Info(fmt.Sprintf("Environment specific folder found for: %s \nUsing %s configs", environ, environ))
		path := fmt.Sprintf("osquery_configs/%s/*", environ)
		env_specific_configs = true
		files, err = filepath.Glob(path)
		if err != nil {
			logger.Error(err)
			return err
		}
	}
	s := spinner.New(spinner.CharSets[43], time.Millisecond*500)
	s.Start()
	credfile, err := UserAwsCredFile()
	dync_svc := auth.CrendentialedDbInstance(credfile, config.AWSProfile)
	for _, fn := range files {
		_, filename := filepath.Split(fn)
		if strings.HasSuffix(filename, "json") {
		}
		fp := ""
		if env_specific_configs {
			fp = filepath.Join("osquery_configs", environ, filename)
		} else {
			fp = filepath.Join("osquery_configs", filename)
		}
		file, err := os.Open(fp)
		named_config := osq_types.OsqueryNamedConfig{}
		defer file.Close()
		if err != nil {
			logger.Error(err)
			return err

		}
		config := osq_types.OsqueryConfig{}
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&config)
		if err != nil {
			logger.Error(err)
			return err

		}
		//fmt.Printf("%s", config.Packs)
		named_config.Config_name = strings.Split(filename, ".")[0]
		switch {
		case strings.Contains(filename, "mac"):
			named_config.Os_type = "mac"
		case strings.Contains(filename, "windows"):
			named_config.Os_type = "windows"
		case strings.Contains(filename, "Linux"):
			named_config.Os_type = "Linux"
		default:
			named_config.Os_type = "all"
		}
		var pl []string
		err = json.Unmarshal(*config.Packs, &pl)
		if err != nil {
			logger.Error(err)
			return err

		}
		named_config.PackList = pl
		//blank out config packs since the options config doesn't have a packs kv
		config.Packs = nil
		named_config.Osquery_config = config
		mu := sync.Mutex{}
		ans := dyndb.UpsertNamedConfig(dync_svc, &named_config, mu)
		s.Stop()
		if ans {
			logger.Info(fmt.Sprintf("%s: success", named_config.Config_name))
		} else {
			logger.Info(fmt.Sprintf("%s: failed", named_config.Config_name))
		}
	}
	return nil
}

func CreateDirIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, 0755)
		if err != nil {
			logger.Error(err)
			return err
		}
	}
	return nil
}

func FindAndReplace(filename, original, replacement string) error {
	fileinfo, _ := os.Stat(filename)
	perms := fileinfo.Mode()
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		logger.Error(err)
		return err
	}
	output := bytes.Replace(input, []byte(original), []byte(replacement), -1)

	if err = ioutil.WriteFile(filename, output, perms); err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

func GenerateEndpointDeployScripts(environ string) error {
	logger.Info(fmt.Sprintf("Updating endpoint deployments scripts for %s environment...", environ))
	config_file := fmt.Sprintf("terraform/%s/%s.json", environ, environ)
	config, err := ParseDeploymentConfig(config_file)
	if err != nil {
		logger.Error(err)
		return err
	}
	// make sure all dirs are created
	err = CreateDirIfNotExists(filepath.Join("endpoints", "deploy", environ))
	if err != nil {
		logger.Error(err)
		return err
	}
	err = CreateDirIfNotExists(filepath.Join("endpoints", "deploy", environ, "Mac"))
	if err != nil {
		logger.Error(err)
		return err
	}
	err = CreateDirIfNotExists(filepath.Join("endpoints", "deploy", environ, "Windows"))
	if err != nil {
		logger.Error(err)
		return err
	}
	err = CreateDirIfNotExists(filepath.Join("endpoints", "deploy", environ, "Linux"))
	if err != nil {
		logger.Error(err)
		return err
	}
	//copy example config files
	err = CopyFile(filepath.Join("endpoints", "deploy", "example_environment", "Mac", "mac_deploy.sh"),
		filepath.Join("endpoints", "deploy", environ, "Mac", "mac_deploy.sh"))
	if err != nil {
		logger.Error(err)
		return err
	}
	CopyFile(filepath.Join("endpoints", "deploy", "example_environment", "Windows", "windows_deploy.ps1"),
		filepath.Join("endpoints", "deploy", environ, "Windows", "windows_deploy.ps1"))
	if err != nil {
		logger.Error(err)
		return err
	}
	CopyFile(filepath.Join("endpoints", "deploy", "example_environment", "Linux", "linux_deploy.sh"),
		filepath.Join("endpoints", "deploy", environ, "Linux", "linux_deploy.sh"))
	if err != nil {
		logger.Error(err)
		return err
	}
	//now modify deployment files for environment
	err = FindAndReplace(filepath.Join("endpoints", "deploy", environ, "Mac", "mac_deploy.sh"), "example-secret", config.SgtNodeSecret)
	if err != nil {
		logger.Error(err)
		return err
	}
	dom_string := fmt.Sprintf("%s.%s", config.Subdomain, config.Domain)
	err = FindAndReplace(filepath.Join("endpoints", "deploy", environ, "Mac", "mac_deploy.sh"), "example.domain.endpoint.com", dom_string)
	if err != nil {
		logger.Error(err)
		return err
	}
	err = FindAndReplace(filepath.Join("endpoints", "deploy", environ, "Linux", "linux_deploy.sh"), "example-secret", config.SgtNodeSecret)
	if err != nil {
		logger.Error(err)
		return err
	}
	dom_string = fmt.Sprintf("%s.%s", config.Subdomain, config.Domain)
	err = FindAndReplace(filepath.Join("endpoints", "deploy", environ, "Linux", "linux_deploy.sh"), "example.domain.endpoint.com", dom_string)
	if err != nil {
		logger.Error(err)
		return err
	}
	err = FindAndReplace(filepath.Join("endpoints", "deploy", environ, "Windows", "windows_deploy.ps1"), "example-secret", config.SgtNodeSecret)
	if err != nil {
		logger.Error(err)
		return err
	}
	dom_string = fmt.Sprintf("%s.%s", config.Subdomain, config.Domain)
	err = FindAndReplace(filepath.Join("endpoints", "deploy", environ, "Windows", "windows_deploy.ps1"), "example.domain.endpoint.com", dom_string)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Info("DONE!")
	return nil
}
