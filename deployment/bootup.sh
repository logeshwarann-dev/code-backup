#!/bin/bash
kubectl apply -f sender.yml
kubectl apply -f trader.yml
# kubectl apply -f peak-generator.yml
kubectl apply -f middleware.yml
kubectl apply -f angular.yml