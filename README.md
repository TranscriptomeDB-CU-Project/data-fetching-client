# Data Client Fetching

This repository is used to fetching the sdrf from biostudies and save locally before processing in the data cleaning process also it saves the fetching status to mongodb

## Requirement

- go 1.21.3

## How to run

1. Copy `.env.example` and rename it to `.env`
2. Create new folder called `sdrf`
3. Run `go run main.go` or `make build && ./main`

## Configuration

- You can adjust fetching parameter in folder `constants/api.go` and `constants/worker.go`
