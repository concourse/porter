#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

export CLUSTER_NAME=porter-test-cluster
export CLUSTER_ZONE=us-central1

echo $GCP_SERVICE_KEY > /tmp/key

gcloud auth activate-service-account --key-file /tmp/key
gcloud container clusters delete $CLUSTER_NAME --zone $CLUSTER_ZONE --quiet

# clean up bucket
# in case the task failed and we are still cleaning up, we don't this to fail
gsutil rm gs://porter-dev-bucket/out.tar || true
