#!/bin/bash

kubectl delete nikhawsec2 nikhawsec2-sample -n nikhil-vmstate-operator-system;
kubectl delete nikhawsmanager nikhawsmanager-sample -n nikhil-vmstate-operator-system;
make undeploy;