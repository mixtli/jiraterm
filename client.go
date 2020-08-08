package main

import (
	"fmt"
	"github.com/andygrunwald/go-jira"
	"log"
)

func GetClient() *jira.Client {
	cfg := GetConfig()

	tp := jira.BasicAuthTransport{
		Username: cfg.Email,
		Password: cfg.ApiKey,
	}

	jiraClient, err := jira.NewClient(tp.Client(), cfg.Endpoint)
	if err != nil {
		fmt.Print("Error Creating Client")
		log.Fatal(err)
	}
	return jiraClient
}
