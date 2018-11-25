#!/bin/bash

set -e

# Build container
docker build -t mxssl/revolut-sre-test-task .

# Push container to the registry
docker push mxssl/revolut-sre-test-task
