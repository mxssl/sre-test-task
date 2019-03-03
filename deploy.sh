#!/bin/bash

set -e

# Create namespace
kubectl create namespace sre-test-task || true

# Deploy the app
kubectl apply -f kube/
