#!/bin/bash
kubectl delete statefulset client fileparser secondary-redis
kubectl delete deployment data-processor metrics-generator