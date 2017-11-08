# ropecount

[![CircleCI](https://circleci.com/gh/koding/rope-infra/tree/master.svg?style=svg&circle-token=1e8d2ffb37bddb5ad2085d46700fe7263e9a419c)](https://circleci.com/gh/koding/rope-infra/tree/master)

## Running in Docker Compose

docker-compose build && docker-compose up

## Running in Kubernetes

### Install Helm

https://raw.githubusercontent.com/kubernetes/helm/master/scripts/get | bash
helm init
helm repo update

### Install Mongo Chart

helm install kubernetes-charts/mongodb --name=mongo
kubectl run mongo-mongodb-client --rm --tty -i --image bitnami/mongodb --command -- mongo --host mongo-mongodb

### Install Redis Chart

helm install kubernetes-charts/redis --name=redis --set=usePassword=false
redis-cli -h redis-redis

### Install Applications/Services

TODO