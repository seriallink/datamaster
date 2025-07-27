package tests

import (
	"log"
	"os"

	"github.com/seriallink/datamaster/app/core"
)

func MustPersistTestConfig() {

	profile := os.Getenv("AWS_PROFILE")
	region := os.Getenv("AWS_REGION")

	err := core.PersistAWSConfig(profile, "", "", region)
	if err != nil {
		log.Fatalf("Failed to persist AWS config for test: %v", err)
	}

}
