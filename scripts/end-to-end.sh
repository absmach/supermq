#!/bin/bash
# Copyright (c) Mainflux
# SPDX-License-Identifier: Apache-2.0

###
# Runs all Mainflux microservices (builds and installs if not done already)
# 
# Uses Schemathesis to check the openAPI configuration with the actual endpoints
###

cd ..

make dockers
make rundetached

EMAIL=example@eg.com
PASSWORD=12345678
DEVICE=mf-device

#provision user:
printf "Provisioning user with email $EMAIL and password $PASSWORD \n"
curl -s -S --insecure -X POST -H "Content-Type: application/json" http://localhost/users -d '{"email":"'"$EMAIL"'", "password":"'"$PASSWORD"'"}'

#get jwt token
JWTTOKEN=$(curl -s -S -X POST -H "Content-Type: application/json" http://localhost/tokens -d '{"email":"'"$EMAIL"'", "password":"'"$PASSWORD"'"}' | grep -Po "token\":\"\K(.*)(?=\")")
printf "JWT TOKEN for user is $JWTTOKEN \n"

echo setting mf base path $(pwd)
export MF_BASE_PATH=$(pwd)

# MF_TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NzM4ODg4NjIsImlhdCI6MTY3Mzg1Mjg2MiwiaXNzIjoibWFpbmZsdXguYXV0aCIsInN1YiI6ImV4YW1wbGVAZWcuY29tIiwiaXNzdWVyX2lkIjoiNzE0NTk5MmYtMzZkZi00NjE5LWE1YzQtOGJkMzg2YjI3YmE5IiwidHlwZSI6MH0.B1CSAPQawWH2UWt3qiD0KfufWuqgNjTaunr0fq4jAVA

echo setting mf auth bearer token $JWTTOKEN
export MF_TOKEN=$JWTTOKEN