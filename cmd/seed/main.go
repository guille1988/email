package main

import (
	"email/internal/bootstrap"
	"email/internal/infrastructure/logger"
)

func main() {
	seeder, err := bootstrap.NewSeeder()

	if err != nil {
		logger.Fatal(err)
	}

	err = bootstrap.RunSeeder(seeder)

	if err != nil {
		logger.Fatal(err)
	}
}
