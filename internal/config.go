package internal

import (
	"encoding/json"
	"os"
)

func LoadConfig(path *string) []string {
	var configContent []string
	if path != nil {
		jsonFile, _ := os.ReadFile(*path)
		json.Unmarshal(jsonFile, &configContent)
	} else {
		jsonFile, _ := os.ReadFile("./taglog.json")
		json.Unmarshal(jsonFile, &configContent)
	}
	defaultPrefixes := []string{"feat", "fix", "perf", "doc", "ref"}
	if len(configContent) > 0 {
		return configContent
	} else {
		return defaultPrefixes
	}
}
