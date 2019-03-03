[![Go Report Card](https://goreportcard.com/badge/github.com/mxssl/sre-test-task)](https://goreportcard.com/report/github.com/mxssl/sre-test-task)

# SRE test task

## Diagram

[GCP diagram](https://github.com/mxssl/sre-test-task/blob/master/gcp_diagram.pdf)

## Build&Deploy Scripts

- Docker container build [build.sh](https://github.com/mxssl/test-task/blob/master/build.sh)
- Deploy to Kubernetes [deploy.sh](https://github.com/mxssl/test-task/blob/master/deploy.sh)

## Local setup with docker-compose

Use this command:

```
docker-compose up -d
```

You can get api via http://localhost:8080

## Local dev

1. Install [go](https://golang.org/dl)
2. Install [godep](https://golang.github.io/dep)
3. Install [golangci-lint](https://github.com/golangci/golangci-lint)
4. Install dependencies

```
make dep
```

5. Run linter

```
make lint
```

6. Run db for local dev

```
docker \
  run \
  -d \
  -e POSTGRES_DB=app \
  -e POSTGRES_USER=user \
  -e POSTGRES_PASSWORD=password \
  -p "5432:5432" \
  postgres
```

7. Declare env variables

```
export DB_HOST="your_db_ip"
export DB_PORT="your_db_port"
export DB_USER="your_db_user"
export DB_NAME="your_db_name"
export DB_PASSWORD="your_db_password"
```

8. Run tests

```
make test
```

9. Build the app

```
make build
```

10. Run the app

```
./app
```

11. Stop the app

```
ctrl + c
```

12. Remove the binary

```
make clean
```
