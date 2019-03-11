package main

import (
	"github.com/kennep/timelapse/domain"
	"github.com/kennep/timelapse/endpoints"
	"github.com/kennep/timelapse/mongo_repository"
	log "github.com/sirupsen/logrus"
)

func main() {
	repository, err := mongo_repository.NewMongoRepository()
	if err != nil {
		log.Errorf("Error connecting to repository: %s", err)
		panic(err)
	}
	log.Info("Connected to repository.")

	users := domain.InitUsersCollection(repository)

	// Serve will not return until the listener is shut down.
	err = endpoints.Serve(users)
	if err != nil {
		log.Errorf("Error starting API server: %s", err)
		panic(err)
	}

	log.Info("Timelapse server exiting.")
}
