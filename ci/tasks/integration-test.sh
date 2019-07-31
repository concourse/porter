#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

check(){
 PODNAME=$1
 attempt_counter=0
 max_attempts=5
 until [ $(kubectl get pod $PODNAME -o custom-columns=NAME:.status.phase --no-headers=true) == "Succeeded" ]; do
     if [ ${attempt_counter} -eq ${max_attempts} ];then
       echo "Max attempts reached"
       exit 1
     fi

     echo "..."
     attempt_counter=$(($attempt_counter+1))
     sleep 1
 done
}

export CLUSTER_NAME=porter-test-cluster
export CLUSTER_ZONE=us-central1

echo $GCP_SERVICE_KEY > /tmp/key

gcloud auth activate-service-account --key-file /tmp/key
gcloud container clusters create $CLUSTER_NAME \
  --cluster-version=latest --zone=$CLUSTER_ZONE \
  --enable-autoscaling --min-nodes=1 --max-nodes=3


kubectl create clusterrolebinding cluster-admin-binding \
     --clusterrole=cluster-admin \
     --user=default
kubectl create secret generic gcp-secrets --from-file=gcp-service-key=/tmp/key

kubectl apply -f porter-repo/ci/pod-push.yaml
check gcp-push-pod
kubectl apply -f porter-repo/ci/pod-pull.yaml
check gcp-pull-pod

#gcloud container clusters delete $CLUSTER_NAME --zone $CLUSTER_ZONE
