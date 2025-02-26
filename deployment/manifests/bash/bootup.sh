#!/bin/bash
cd ..
kubectl apply -f redis
kubectl apply -f config
kubectl apply -f metric-generator
kubectl apply -f parser
kubectl apply -f data-processor
kubectl apply -f client