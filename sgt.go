package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/oktasecuritylabs/sgt/dyndb"
	"github.com/oktasecuritylabs/sgt/handlers/auth"
	"github.com/oktasecuritylabs/sgt/handlers/deploy"
	"github.com/oktasecuritylabs/sgt/handlers/helpers"
	"github.com/oktasecuritylabs/sgt/logger"
	"github.com/oktasecuritylabs/sgt/server"
)

const (
	runServer    = "server"
	runWizard    = "wizard"
	runNewDeploy = "new-deployment"
	createUser   = "create-user"
	runDeploy    = "deploy"
	runDestroy   = "destroy"
)

var commands = map[string]string{
	runServer:    "Run in server mode, launching the TLS server",
	runWizard:    "Run deployment configuration wizard",
	runNewDeploy: "Created new deployment",
	createUser:   "Create a new user",
	runDeploy:    "Deploy new sgt environment",
	runDestroy:   "Destroy existing infrastructure",
}

func printHelp(err interface{}) {
	// Print any optional errors passed
	if err != nil {
		logger.Error(err)
	}

	fmt.Print("usage: sgt <command> [<args>]\n\n")
	fmt.Printf("Commands:\n%s\n", formatSection(commands))
	os.Exit(0)
}

func formatSection(commands map[string]string) string {
	keys := make([]string, 0, len(commands))
	maxLen := 0
	for key := range commands {
		if len(key) > maxLen {
			maxLen = len(key)
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	for _, key := range keys {
		commandHelp, ok := commands[key]
		if !ok {
			logger.Fatal(fmt.Sprintf("command not found: %s", key))
		}

		key = fmt.Sprintf("%s%s", key, strings.Repeat(" ", maxLen-len(key)))
		buf.WriteString(fmt.Sprintf("    %s    %s\n", key, commandHelp))
	}

	return buf.String()
}

// New type for a list of deploy commands
type componentChoices []string

// Implement String() from the flag.Value interface
func (s *componentChoices) String() string {
	return fmt.Sprintf("%v", *s)
}

// Implement Set(string) from the flag.Value interface
func (s *componentChoices) Set(value string) error {
	*s = strings.Split(value, ",")
	return nil
}

// validateChoices takes one list of acceptable options and a second list
// of provided options and returns an error if an invalid options are selected.
// If all options are valid, this returns a de-duplicated map of them
func validateChoices(validOpts, chosenOpts []string) ([]string, error) {

	if chosenOpts == nil {
		return nil, errors.New("at least one component to deploy must be supplied")
	}

	chosenOptsMap := make(map[string]bool)
	for _, k := range chosenOpts {
		chosenOptsMap[k] = true
	}

	validOptsMap := make(map[string]bool)
	for _, k := range validOpts {
		validOptsMap[k] = true
	}

	var chosenOptions []string
	for k := range chosenOptsMap {
		if _, ok := validOptsMap[k]; !ok {
			return nil, fmt.Errorf("invalid deploy component specified: %s", k)
		}
		chosenOptions = append(chosenOptions, k)
	}

	return chosenOptions, nil
}

func runSGT() error {

	if len(os.Args) == 1 {
		printHelp(nil)
	}

	switch os.Args[1] {
	case runWizard:
		err := deploy.Wizard()
		if err != nil {
			logger.Fatal(err)
		}

	case runNewDeploy:
		// Create a FlagSet for the new-deployment command
		newDeployCommand := flag.NewFlagSet(runNewDeploy, flag.ExitOnError)
		envFlag := newDeployCommand.String("env", "", "Deployment environment name")

		newDeployCommand.Parse(os.Args[2:])

		envName := strings.TrimSpace(*envFlag)
		if envName == "" {
			var err error
			if envName, err = helpers.GetValueFromUser("Enter new environment name"); err != nil {
				newDeployCommand.Usage()
				return err
			}
		}

		if err := deploy.CreateDeployDirectory(envName); err != nil {
			return err
		}

		logger.Warn(fmt.Sprintf("New deployment created. Remember to go change the defaults in your terraform/%[1]s/%[1]s.json files!", envName))

	case createUser:
		// Create a FlagSet for the create-user command
		createUserCommand := flag.NewFlagSet("create-user", flag.ExitOnError)
		usernameFlag := createUserCommand.String("username", "", "username to create")
		credFileFlag := createUserCommand.String("credentials-file", "~/.aws/credentials",
			"path to aws credentials file (default: ~/.aws/credentials)")
		profileFlag := createUserCommand.String("profile", "", "aws profile name")
		roleFlag := createUserCommand.String("role", "user", "role to be used (default: user)")

		createUserCommand.Parse(os.Args[2:])

		// invalid is an anonymous helper function to validate the length of the input
		invalid := func(v *string) bool {
			return len(strings.TrimSpace(*v)) < 4
		}

		if invalid(usernameFlag) {
			createUserCommand.Usage()
			return errors.New("username required, please pass username via -username flag")
		}
		if invalid(credFileFlag) {
			createUserCommand.Usage()
			return errors.New("aws credentials file required, please pass via -credentialsFile flag")
		}
		if invalid(profileFlag) {
			createUserCommand.Usage()
			return errors.New("aws profile name required, please pass via -profile flag")
		}
		dyn := dyndb.DynDB{}
		err := auth.NewUser(*credFileFlag, *profileFlag, *usernameFlag, *roleFlag, dyn)
		if err != nil {
			return err
		}

	case runDeploy:

		validComponents := map[string]bool{}
		components := []string{}
		components = append(components, deploy.DeployOrder...)
		components = append(components, deploy.OsqueryOpts...)
		components = append(components, deploy.ElasticDeployOrder...)
		for _, component := range components {
			validComponents[component] = true
		}
		validComponentOptions := []string{}
		for k, _ := range validComponents {
			validComponentOptions = append(validComponentOptions, k)
		}
		sort.Strings(validComponentOptions)
		componentList := strings.Join(validComponentOptions, ", ")

		// Create a FlagSet for the deploy command
		deployCommand := flag.NewFlagSet(runDeploy, flag.ExitOnError)
		envFlag := deployCommand.String("env", "", "Deployment environment name")
		allFlag := deployCommand.Bool("all", false, fmt.Sprintf("Deploy all components [%s]", componentList))

		// Create a list of components to deploy that can be chosen from
		var chosenDeployOptions componentChoices
		deployCommand.Var(&chosenDeployOptions, "components",
			fmt.Sprintf("A comma-seperated (without spaces) list of components to deploy. Choices are: %s",
				componentList))

		deployCommand.Parse(os.Args[2:])
		envName := *envFlag
		if envName == "" {
			var err error
			if envName, err = helpers.GetValueFromUser("Enter new environment name"); err != nil {
				deployCommand.Usage()
				return err
			}
		}

		config, err := deploy.ParseDeploymentConfig(envName)
		if err != nil {
			return err
		}

		if *allFlag {
			prompt := fmt.Sprint("All components will be deployed.\nDo you want to continue?")
			if !helpers.ConfirmAction(prompt) {
				return nil // User canceled action
			}
			return deploy.AllComponents(config, envName)
		}

		chosenOptions, err := validateChoices(validComponentOptions, chosenDeployOptions)
		if err != nil {
			deployCommand.Usage()
			return err
		}

		// Prompt for deployment confirmation
		prompt := fmt.Sprintf("The following components will be deployed: %s\nDo you want to continue?", strings.Join(chosenOptions, ", "))
		if helpers.ConfirmAction(prompt) {
			log.Printf("Beginning deployment to %[1]s using configuration specified in %[1]s.json", envName)
			for _, componentName := range chosenOptions {
				if err = deploy.Component(config, componentName, envName); err != nil {
					return err
				}
			}
		}

	case runDestroy:

		componentList := strings.Join(deploy.DeployOrder, ", ")

		// Create a FlagSet for the destory command
		destroyCommand := flag.NewFlagSet("destroy", flag.ExitOnError)
		envFlag := destroyCommand.String("env", "", "Deployment environment name")
		allFlag := destroyCommand.Bool("all", false, fmt.Sprintf("Deploy all components [%s]", componentList))

		// Create a list of components to destroy that can be chosen from
		var chosenDestroyOptions componentChoices
		destroyCommand.Var(&chosenDestroyOptions, "components",
			fmt.Sprintf("A comma-seperated (without spaces) list of components to destroy. Choices are: %s",
				componentList))

		destroyCommand.Parse(os.Args[2:])
		envName := *envFlag
		config, err := deploy.ParseDeploymentConfig(envName)

		if envName == "" {
			err := errors.New("No environment specified.  Please rerun with an existing environment")
			logger.Fatal(err)
		}

		if *allFlag {
			prompt := fmt.Sprint("All components will be destroyed.\nDo you want to continue?")
			if !helpers.ConfirmAction(prompt) {
				return nil // User canceled action
			}
			return deploy.DestroyAllComponents(config, envName)
		}

		chosenOptions, err := validateChoices(deploy.DeployOrder, chosenDestroyOptions)
		if err != nil {
			destroyCommand.Usage()
			return err
		}

		// Prompt for destroy confirmation
		prompt := fmt.Sprintf("The following components will be destroyed: %s\nDo you want to continue?", strings.Join(chosenOptions, ", "))
		if helpers.ConfirmAction(prompt) {
			for _, componentName := range chosenOptions {
				deploy.DestroyComponent(componentName, envName)
			}
		}

	case runServer:
		return server.Serve()
	default:
		printHelp(nil)
	}

	return nil
}

func main() {
	err := runSGT()
	if err != nil {
		logger.Fatal(err)
	}

	os.Exit(0)
}
