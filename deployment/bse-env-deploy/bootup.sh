#!/bin/bash
kubectl apply -f manifests/sender.yml
kubectl apply -f manifests/trader.yml
kubectl apply -f manifests/middleware.yml
kubectl apply -f manifests/angular.yml