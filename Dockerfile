FROM golang:alpine as builder

WORKDIR /go/src/github.com/mxssl/sre-test-task
COPY . .

# install dep package manager
RUN apk add --no-cache ca-certificates curl git
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
RUN dep ensure
RUN CGO_ENABLED=0 GOOS=`go env GOHOSTOS` GOARCH=`go env GOHOSTARCH` go build -o app

# Copy compiled binary to clear Alpine Linux image
FROM alpine:latest
WORKDIR /
RUN apk add --no-cache ca-certificates
COPY --from=builder /go/src/github.com/mxssl/sre-test-task .
RUN chmod +x app
CMD ["./app"]
