package main

import (
	"log"
	"os"
	"path"
)

func main() {

	logfile := path.Join(".", "jiraterm.log")
	f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0700)
	if err != nil {
		log.Fatal("Failed to open logfile", err)
	}
	defer f.Close()
	log.SetOutput(f)
	RunUI()
}
