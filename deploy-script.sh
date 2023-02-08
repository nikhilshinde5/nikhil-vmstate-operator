#!/bin/bash

make generate;
make manifests;
make docker-build;
make docker-push;
make deploy;
kubectl create secret generic aws-secret --from-literal=aws-default-region=us-east-1 --from-literal=aws-secret-access-key=46RBao2j+UAHsIA6Jv9asu8LhVMUR1pBiDb+FhKU --from-literal=aws-access-key-id=AKIAWT4IN5FUAI54GCMK -n nikhil-vmstate-operator-system;
# kubectl apply -f my-config-map.yaml -n nikhil-vmstate-operator-system;
kubectl create configmap my-config-map --from-file=configmap/config.json -n nikhil-vmstate-operator-system
kubectl apply -f config/samples/aws_v1_nikhawsmanager.yaml -n nikhil-vmstate-operator-system;
kubectl apply -f config/samples/aws_v1_nikhawsec2.yaml -n nikhil-vmstate-operator-system;
kubectl get all -n nikhil-vmstate-operator-system;