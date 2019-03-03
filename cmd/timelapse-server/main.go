package main

import (
	"github.com/kennep/timelapse/api"
	"github.com/kennep/timelapse/repository"
	log "github.com/sirupsen/logrus"
)

func main() {
	repository, err := repository.NewRepository()
	if err != nil {
		log.Errorf("Error connecting to repository: %s", err)
		panic(err)
	}
	log.Info("Connected to repository.")

	// Serve will not return until the listener is shut down.
	err = api.Serve(repository)
	if err != nil {
		log.Errorf("Error starting API server: %s", err)
		panic(err)
	}

	log.Info("Timelapse server exiting.")
}
