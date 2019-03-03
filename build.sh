#!/bin/bash

set -e

# Build container
docker build -t mxssl/sre-test-task .

# Push container to the registry
docker push mxssl/sre-test-task
