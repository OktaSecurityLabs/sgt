package helpers

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/oktasecuritylabs/sgt/logger"
)

// GetValueFromUser prompts the user to provide a value
func GetValueFromUser(prompt string) (string, error) {

	if prompt == "" {
		prompt = "Please provide a value"
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s:\n", prompt)

	value, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("a non-empty value name must be supplied")
	}

	return value, nil
}

// ConfirmAction scans user input for a yes or no confirmation
func ConfirmAction(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)

	// The input gets lower-cased so only need to iterate on these options
	confirmStrings := []string{"y", "yes"}
	denyStrings := []string{"n", "no"}

	for tries := 0; tries < 3; tries++ {
		fmt.Printf("%s [y/n]: ", prompt)

		response, err := reader.ReadString('\n')
		if err != nil {
			logger.Error(err)
			return false
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if len(response) == 0 {
			continue
		}

		// Check for approvals
		for _, i := range confirmStrings {
			if i == response {
				return true
			}
		}

		// Check for denials
		for _, i := range denyStrings {
			if i == response {
				return false
			}
		}
	}

	return false
}

func CleanPack(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	pack := ""
	full_line := ""
	use_nextline := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasSuffix(line, "\\") {
			use_nextline = true
			no_backslash := strings.Replace(line, "\\", "", -1)
			full_line += " "
			full_line += strings.TrimSpace(no_backslash)
		}
		if use_nextline {
			full_line += " "
			no_backslash := strings.Replace(line, "\\", "", -1)
			full_line += strings.TrimSpace(no_backslash)
			pack += full_line
			full_line = ""
			use_nextline = false
		} else {
			no_backslash := strings.Replace(line, "\\", "", -1)
			pack += strings.TrimSpace(no_backslash)
		}
	}
	return pack, nil
}

type OsqueryPack struct {
	Queries  map[string]PackQuery `json:"queries"`
	Platform string               `json:"platform"`
}

type PackQuery struct {
	Query       string `json:"query"`
	Interval    string `json:"interval"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Value       string `json:"value"`
	Snapshot	string `json:"snapshot"`
}

func (op OsqueryPack) ListQueries() []string {
	queries := []string{}
	for i, _ := range op.Queries {
		queries = append(queries, i)
	}
	return queries
}
