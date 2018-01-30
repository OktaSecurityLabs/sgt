package deploy

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/oktasecuritylabs/sgt/logger"
)

// DestroyAllComponents destroys all components, in reverse order
func DestroyAllComponents(envName string) error {
	for i := len(DeployOrder) - 1; i >= 0; i-- {
		if err := destroyAWSComponent(DeployOrder[i], envName); err != nil {
			return err
		}
	}
	return nil
}

// DestroyComponent is a wrapper to match the `deploy.Component` interface style
func DestroyComponent(component, envName string) error {
	return destroyAWSComponent(component, envName)
}

// destroyAWSComponent destroys aws components
// This includes: VPC, Datastore, Firehose, Autoscaling, Secrets, Elasticsearch, S3
func destroyAWSComponent(component, envName string) error {

	// Change back to the top level directory after each component deploy
	cachedCurDir, _ := os.Getwd()
	defer os.Chdir(cachedCurDir)

	spin.Start()
	defer spin.Stop()

	logger.Infof("Destroying %s...", component)

	err := os.Chdir(fmt.Sprintf("terraform/%s/%s", envName, component))
	if err != nil {
		return err
	}

	args := fmt.Sprintf("terraform destroy -force -var-file=../%s.json", envName)
	logger.Info(args)

	cmd := exec.Command("bash", "-c", args)
	stdoutStderr, err := cmd.CombinedOutput()

	logger.Info(string(stdoutStderr))

	return err
}
