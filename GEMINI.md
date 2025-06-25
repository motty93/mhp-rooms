# Gemini Project Brief: mhp-rooms

This file provides a brief overview of the project for the Gemini agent.

## Project Overview

`mhp-rooms` appears to be a web application written in Go. It uses HTML templates for the frontend and seems to have features related to user authentication, rooms, and profiles. The database schema is likely defined in `scripts/init.sql`.

## Common Commands

Based on the `Makefile` and `cmd` directory, here are some likely commands:

*   `make server`: To build and run the server.
*   `make migrate`: To run database migrations.
*   `make seed`: To seed the database with initial data.
*   `go run ./cmd/server/main.go`: To run the server directly.
*   `go test ./...`: To run tests.

## Key Files

*   `go.mod`, `go.sum`: Go module dependencies.
*   `compose.yml`: Docker Compose configuration for local development.
*   `Dockerfile`: Docker configuration for building the application container.
*   `cmd/server/main.go`: The main entry point for the web server.
*   `internal/handlers/`: Contains the HTTP handlers for different routes.
*   `internal/repository/`: Contains the database access logic.
*   `internal/models/`: Defines the data structures used in the application.
*   `templates/`: HTML templates for the user interface.
*   `static/`: Static assets like CSS, JavaScript, and images.
*   `scripts/init.sql`: SQL script for initializing the database schema.
