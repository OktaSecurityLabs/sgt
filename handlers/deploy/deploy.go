package deploy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
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

const (
	vpc                   = "vpc"
	datastore             = "datastore"
	elasticsearch         = "elasticsearch"
	firehose              = "firehose"
	elasticsearchFirehose = "elasticsearch_firehose"
	s3                    = "s3"
	secrets               = "secrets"
	autoscaling           = "autoscaling"
	packs                 = "packs"
	configs               = "configs"
	scripts               = "scripts"
)

var (
	spin = spinner.New(spinner.CharSets[43], time.Millisecond*500)

	// DeployOrder dictates what order the deployed components should happen in
	DeployOrder = []string{
		vpc,
		datastore,
		firehose,
		s3,
		secrets,
		autoscaling,
	}

	ElasticDeployOrder = []string{
		vpc,
		datastore,
		elasticsearch,
		elasticsearchFirehose,
		s3,
		secrets,
		autoscaling,
	}

	// OsqueryOpts holds all deploy options for osquery
	OsqueryOpts = []string{
		packs,
		configs,
		scripts,
	}

	osqueryDeployCommands = map[string]func(DeploymentConfig, string) error{
		packs:   osqueryDefaultPacks,
		configs: osqueryDefaultConfigs,
		scripts: generateEndpointDeployScripts,
	}
)

// DeploymentConfig configuration file used by all environment deployments
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
	CreateElasticsearch         int    `json:"create_elasticsearch"`
}

// copyFile copies file from src to dst
func copyFile(src, dst string) error {
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

// ParseDeploymentConfig returns the loaded config given its path
// on disk or exits with status 1 on failure
func ParseDeploymentConfig(environ string) (DeploymentConfig, error) {
	depConf := DeploymentConfig{}
	configFilePath := fmt.Sprintf("terraform/%[1]s/%[1]s.json", environ)
	file, err := os.Open(configFilePath)
	if err != nil {
		return depConf, err
	}

	decoder := json.NewDecoder(file)

	if err = decoder.Decode(&depConf); err != nil {
		return depConf, err
	}

	if err = depConf.checkEnvironMatchConfig(environ); err != nil {
		return depConf, err
	}

	return depConf, nil
}

//checkEnvironMatchconfig checks to make sure the environment config passed matches the environment
//specified in the config
func (d DeploymentConfig) checkEnvironMatchConfig(environ string) error {
	if d.Environment != environ {
		return errors.New("config environment and passed environment variable do not match")
	}
	return nil
}

//CreateDeployDirectory  Creates deployment director based on environment
func CreateDeployDirectory(environ string) error {
	path := fmt.Sprintf("terraform/%s", environ)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logger.Infof("creating new deployment environment: %s\n", environ)
		os.Mkdir(path, 0755)
	} else {
		logger.Info("environment already exists, are you sure you meant to to use deploy to\n")
		logger.Info(environ)
		os.Exit(0)
	}
	dirs := []string{"vpc", "datastore", "firehose", "elasticsearch_firehose", "elasticsearch", "s3", "autoscaling", "secrets"}
	for _, p := range dirs {
		dir := filepath.Join(path, p)
		//logger.Info(dir)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			logger.Infof("Creating %s directory\n", dir)
			os.Mkdir(dir, 0755)
		}
	}
	return nil
}

// UserAwsCredFile returns a users aws credential file from their home dir
func UserAwsCredFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		logger.Error(err)
		return "", err
	}
	credfile := filepath.Join(usr.HomeDir, ".aws", "credentials")
	return credfile, nil
}

// osqueryDefaultPacks deploys default packs for an environment if they exist, otherwise deploys
// from normal defaults
func osqueryDefaultPacks(config DeploymentConfig, environ string) error {
	var files []string

	//if environ specific dir exists in packs, deploy those.  Otherwise use defaults
	if _, err := os.Stat(filepath.Join("packs", environ)); os.IsNotExist(err) {
		logger.Infof("No environment specific packs found for: %s\n", environ)
		logger.Info("using default packs")
		files, err = filepath.Glob("packs/*")
		if err != nil {
			return err
		}
	} else {
		logger.Infof("Environment specific folder found for: %[1]s\nUsing %[1]s query packs\n", environ)
		path := fmt.Sprintf("packs/%s/*", environ)
		files, err = filepath.Glob(path)
		if err != nil {
			return err
		}
	}

	for _, fn := range files {

		_, filename := filepath.Split(fn)
		if strings.HasSuffix(filename, "json") {
			logger.Infof("Deploying %s", filename)
			pack := osq_types.QueryPack{}
			helperPack := helpers.OsqueryPack{}
			s, err := helpers.CleanPack(filename)
			if err != nil {
				return err
			}
			file := strings.NewReader(s)
			if err != nil {
				return err
			}
			decoder := json.NewDecoder(file)
			err = decoder.Decode(&helperPack)
			if err != nil {
				return err
			}
			credfile, err := UserAwsCredFile()
			if err != nil {
				return err
			}
			dynDBInstance := auth.CrendentialedDbInstance(credfile, config.AWSProfile)
			mu := sync.Mutex{}
			//logger.Infof("%+v", pack)
			//logger.Infof("%+v", helperPack)
			for k, v := range helperPack.Queries {
				pq := osq_types.PackQuery{}
				pq.QueryName = k
				pq.Query = v.Query
				pq.Value = v.Value
				pq.Description = v.Description
				pq.Interval = v.Interval
				pq.Version = v.Version
				dyndb.UpsertPackQuery(pq, dynDBInstance, &mu)
			}
			//logger.Info("queries done\n")
			pack.Queries = helperPack.ListQueries()
			pack.PackName = strings.Split(filename, ".")[0]
			err = dyndb.UpsertPack(pack, dynDBInstance)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// osqueryDefaultConfigs deploys default configs for env
func osqueryDefaultConfigs(config DeploymentConfig, environ string) error {
	var files []string

	//if environ specific dir exists in packs, deploy those.  Otherwise use defaults
	var envSpecificConfigs bool
	if _, err := os.Stat(filepath.Join("osquery_configs", environ)); os.IsNotExist(err) {
		logger.Infof("No environment specific configs found for: %s\n", environ)
		logger.Info("using default configs")
		files, err = filepath.Glob("osquery_configs/defaults/*")
		environ = "defaults"
		envSpecificConfigs = true
		if err != nil {
			return err
		}
	} else {
		logger.Infof("Environment specific folder found for: %[1]s\nUsing %[1]s configs\n", environ)
		path := fmt.Sprintf("osquery_configs/%s/*", environ)
		envSpecificConfigs = true
		files, err = filepath.Glob(path)
		if err != nil {
			return err
		}
	}
	//spin.Start()
	//defer spin.Stop()
	credfile, err := UserAwsCredFile()
	if err != nil {
		logger.Fatal(err)
	}
	dynDB := auth.CrendentialedDbInstance(credfile, config.AWSProfile)
	for _, fn := range files {
		_, filename := filepath.Split(fn)
		logger.Infof("Updating %s pack", filename)
		if strings.HasSuffix(filename, "json") {
		}
		var fp string
		if envSpecificConfigs {
			fp = filepath.Join("osquery_configs", environ, filename)
		} else {
			fp = filepath.Join("osquery_configs", filename)
		}
		file, err := os.Open(fp)
		namedConfig := osq_types.OsqueryNamedConfig{}
		defer file.Close()
		if err != nil {
			return err
		}
		config := osq_types.OsqueryConfig{}
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&config)
		if err != nil {
			return err
		}
		//fmt.Printf("%s", config.Packs)
		namedConfig.ConfigName = strings.Split(filename, ".")[0]
		switch {
		case strings.Contains(filename, "mac"):
			namedConfig.OsType = "mac"
		case strings.Contains(filename, "windows"):
			namedConfig.OsType = "windows"
		case strings.Contains(filename, "Linux"):
			namedConfig.OsType = "Linux"
		default:
			namedConfig.OsType = "all"
		}
		var pl []string
		err = json.Unmarshal(*config.Packs, &pl)
		if err != nil {
			return err
		}

		namedConfig.PackList = pl
		//blank out config packs since the options config doesn't have a packs kv
		config.Packs = nil
		namedConfig.OsqueryConfig = config
		mu := sync.Mutex{}
		err = dyndb.UpsertNamedConfig(dynDB, &namedConfig, &mu)
		if err != nil {
			logger.Infof("%s: failed\n", namedConfig.ConfigName)
			return err
		}

		logger.Infof("%s: success\n", namedConfig.ConfigName)
	}

	return nil
}

//CreateDirIfNotExists creates directory if it does not exist
func CreateDirIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

//FindAndReplace finds a string and resplaces it with replacement
func FindAndReplace(filename, original, replacement string) error {
	fileinfo, _ := os.Stat(filename)
	perms := fileinfo.Mode()
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	output := bytes.Replace(input, []byte(original), []byte(replacement), -1)

	return ioutil.WriteFile(filename, output, perms)
}

// generateEndpointDeployScripts generates endpoint scripts for installation
func generateEndpointDeployScripts(config DeploymentConfig, environ string) error {
	logger.Infof("Updating endpoint deployments scripts for %s environment...\n", environ)

	// make sure all dirs are created
	err := CreateDirIfNotExists(filepath.Join("endpoints", "deploy", environ))
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
	err = copyFile(filepath.Join("endpoints", "deploy", "example_environment", "Mac", "mac_deploy.sh"),
		filepath.Join("endpoints", "deploy", environ, "Mac", "mac_deploy.sh"))
	if err != nil {
		logger.Error(err)
		return err
	}
	copyFile(filepath.Join("endpoints", "deploy", "example_environment", "Windows", "windows_deploy.ps1"),
		filepath.Join("endpoints", "deploy", environ, "Windows", "windows_deploy.ps1"))
	if err != nil {
		logger.Error(err)
		return err
	}
	copyFile(filepath.Join("endpoints", "deploy", "example_environment", "Linux", "linux_deploy.sh"),
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
	domainString := fmt.Sprintf("%s.%s", config.Subdomain, config.Domain)
	err = FindAndReplace(filepath.Join("endpoints", "deploy", environ, "Mac", "mac_deploy.sh"), "example.domain.endpoint.com", domainString)
	if err != nil {
		logger.Error(err)
		return err
	}
	err = FindAndReplace(filepath.Join("endpoints", "deploy", environ, "Linux", "linux_deploy.sh"), "example-secret", config.SgtNodeSecret)
	if err != nil {
		logger.Error(err)
		return err
	}
	domainString = fmt.Sprintf("%s.%s", config.Subdomain, config.Domain)
	err = FindAndReplace(filepath.Join("endpoints", "deploy", environ, "Linux", "linux_deploy.sh"), "example.domain.endpoint.com", domainString)
	if err != nil {
		logger.Error(err)
		return err
	}
	err = FindAndReplace(filepath.Join("endpoints", "deploy", environ, "Windows", "windows_deploy.ps1"), "example-secret", config.SgtNodeSecret)
	if err != nil {
		logger.Error(err)
		return err
	}
	domainString = fmt.Sprintf("%s.%s", config.Subdomain, config.Domain)
	err = FindAndReplace(filepath.Join("endpoints", "deploy", environ, "Windows", "windows_deploy.ps1"), "example.domain.endpoint.com", domainString)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Info("DONE!")
	return nil
}
