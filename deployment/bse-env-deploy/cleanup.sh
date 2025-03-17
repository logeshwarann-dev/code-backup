#!/bin/bash
kubectl delete -f manifests/sender.yml
kubectl delete -f manifests/trader.yml
kubectl delete -f manifests/middleware.yml
kubectl delete -f manifests/angular.yml