package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	Email    string  `yaml:"email"`
	Endpoint string  `yaml:"endpoint"`
	ApiKey   string  `yaml:"api_key"`
	Queries  []Query `yaml:"queries"`
}

type Query struct {
	Name  string `yaml:"name"`
	Query string `yaml:"query"`
}

func GetConfig() Config {
	cfg := Config{}
	home, _ := os.UserHomeDir()
	path := home + "/.config/jiraterm/config.yml"
	dat, _ := ioutil.ReadFile(path)
	err := yaml.Unmarshal([]byte(dat), &cfg)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return cfg
}

func GetQueries() []string {
	cfg := GetConfig()
	searchNames := make([]string, len(cfg.Queries))
	for i := 0; i < len(cfg.Queries); i++ {
		searchNames[i] = cfg.Queries[i].Name
	}
	return searchNames
}

func GetQuery(name string) string {
	cfg := GetConfig()
	for i := 0; i < len(cfg.Queries); i++ {
		if cfg.Queries[i].Name == name {
			return cfg.Queries[i].Query
		}
	}
	return ""
}
