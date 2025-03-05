#!/bin/bash
kubectl delete -f sender.yml
kubectl delete -f trader.yml
# kubectl delete -f peak-generator.yml
kubectl delete -f middleware.yml
kubectl delete -f angular.yml