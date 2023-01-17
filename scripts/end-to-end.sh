#!/bin/bash
# Copyright (c) Mainflux
# SPDX-License-Identifier: Apache-2.0

###
# Runs all Mainflux microservices (builds and installs if not done already)
# 
# Uses Schemathesis to check the openAPI configuration with the actual endpoints
###

echo go to root directory
cd ..

echo build images locallu
make dockers

echo run all the containers
make run ARGS="-d"

EMAIL=example@eg.com
PASSWORD=12345678
DEVICE=mf-device

printf "Provisioning user with email $EMAIL and password $PASSWORD \n"
curl -s -S --insecure -X POST -H "Content-Type: application/json" http://localhost:8180/users -d '{"email":"'"$EMAIL"'", "password":"'"$PASSWORD"'"}'

#get jwt token
JWTTOKEN=$(curl -s -S -i -X POST -H "Content-Type: application/json" http://localhost:8180/tokens -d '{"email":"'"$EMAIL"'", "password":"'"$PASSWORD"'"}' | grep -Po "token\":\"\K(.*)(?=\")")
printf "JWT TOKEN for user is $JWTTOKEN \n"

echo setting mf base path $(pwd)
export MF_BASE_PATH=$(pwd)

echo setting mf auth bearer token $(JWTTOKEN)
export MF_TOKEN=$(JWTTOKEN)