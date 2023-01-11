#!/bin/bash
# Copyright (c) Mainflux
# SPDX-License-Identifier: Apache-2.0

###
# Runs all Mainflux microservices (builds and installs if not done already)
# 
# Uses Schemathesis to check the openAPI configuration with the actual endpoints
###

chmod 776 end-to-end.sh

cd ..

echo running all docker containers now
sudo make run

echo "now provisioning for MF_TOKEN"

# 

EMAIL=example@eg.com
PASSWORD=12345678
DEVICE=mf-device

#provision user:
printf "Provisoning user with email $EMAIL and password $PASSWORD \n"
curl -s -S --cacert docker/ssl/certs/mainflux-server.crt --insecure -X POST -H "Content-Type: application/json" https://localhost/users -d '{"email":"'"$EMAIL"'", "password":"'"$PASSWORD"'"}'

#get jwt token
JWTTOKEN=$(curl -s -S --cacert docker/ssl/certs/mainflux-server.crt --insecure -X POST -H "Content-Type: application/json" https://localhost/tokens -d '{"email":"'"$EMAIL"'", "password":"'"$PASSWORD"'"}' | grep -Po "token\":\"\K(.*)(?=\")")
printf "JWT TOKEN for user is $JWTTOKEN \n"

#provision thing
printf "Provisioning thing with name $DEVICE \n"
curl -s -S --cacert docker/ssl/certs/mainflux-server.crt --insecure -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $JWTTOKEN" https://localhost/things -d '{"name":"'"$DEVICE"'"}'

#get thing token
DEVICETOKEN=$(curl -s -S --cacert docker/ssl/certs/mainflux-server.crt --insecure -H "Authorization: Bearer $JWTTOKEN" https://localhost/things/1 | grep -Po "key\":\"\K(.*)(?=\")")
printf "Device token is $DEVICETOKEN \n"

echo setting mf base path
export MF_BASE_PATH=$(pwd)

echo setting mf auth bearer token
export MF_TOKEN=$JWTTOKEN
#TODO: Define rest of the constants like {id} or {key} , etc.
printf "Got the MF_TOKEN : $MF_TOKEN \n"

#! TASK -> Automate below step instead of manually typing

# cd ./scripts
# make test
# cd -

# echo deleting the stapi container
# sudo docker rm stapi

# echo stopping the running containers
# sudo docker-compose -f docker/docker-compose.yml down
