#!/bin/bash

kubectl delete deployment app1
kubectl delete crd backingservices.app1.example.org
kubectl delete secrets secret1

kubectl delete deployment app2
kubectl delete crd backingservices.app2.example.org
kubectl delete secrets secret2

kubectl delete crd servicebindings.binding.x-k8s.io
sleep 5
kubectl get pod
