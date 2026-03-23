package main

import (
	"email/internal/bootstrap"
	"email/internal/infrastructure/logger"
)

func main() {
	worker, err := bootstrap.NewWorker()

	if err != nil {
		logger.Fatal(err)
	}

	err = bootstrap.RunWorker(worker)

	if err != nil {
		logger.Fatal(err)
	}
}
