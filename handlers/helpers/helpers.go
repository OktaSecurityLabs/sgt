package helpers

import (
	"os"
	"path/filepath"
	"github.com/oktasecuritylabs/sgt/logger"
	"bufio"
	"strings"
)

func CleanPack(filename string) (string, error) {
	file, err := os.Open(filepath.Join("packs", filename))
	if err != nil {
		logger.Error(err)
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
	Queries map[string]PackQuery `json:"queries"`
	Platform string `json:"platform"`
}

type PackQuery struct {
	Query string `json:"query"`
	Interval string `json:"interval"`
	Version string `json:"version"`
	Description string `json:"description"`
	Value string `json:"value"`
}

func (op OsqueryPack) ListQueries() ([]string) {
	queries := []string{}
	for i, _ := range op.Queries {
		queries = append(queries, i)
	}
	return queries
}
