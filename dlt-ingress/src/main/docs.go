package main

import (
	// Imports to force swag to read the packages
	_ "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/pagination"
	_ "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/api/model"
)

// @title DLT Ingress API
// @version 1.0.0
// @description API for DLT Ingress - Build, sign and send transactions on any DLT or Blockchain.
// @host localhost:8080
// @BasePath /api/v1/
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func OpenApi() {}
