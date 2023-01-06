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

#! Either check for the command first, if it exists, OR shift to using docker container for schemathesis
echo install schemathesis
pip install schemathesis

echo running all docker containers now
sudo make run
echo have the containers started running

echo "now provisioning for mf token"

# 

EMAIL=example@eg.com
PASSWORD=12345678
DEVICE=aryan
CHANNEL=ch1

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

#provision channel
printf "Provisioning channel with name $CHANNEL \n"
curl -s -S --cacert docker/ssl/certs/mainflux-server.crt --insecure -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $JWTTOKEN" https://localhost/channels -d '{"name":"'"$CHANNEL"'"}'

#connect thing to channel
printf "Connecting thing to channel \n"
curl -s -S --cacert docker/ssl/certs/mainflux-server.crt --insecure -X PUT -H "Authorization: Bearer $JWTTOKEN" https://localhost/channels/1/things/1

# 

echo setting mf auth bearer token
MF_TOKEN=$JWTTOKEN
#TODO: Define rest of the constants like {id} or {key} , etc.
printf "Got the MF_TOKEN : $MF_TOKEN \n"

#! TASK -> Automate below step instead of manually typing

# SERVICES = auth bootstrap certs consumers-notifiers http provision readers things twins users websockets

cd ./scripts
make test
cd -

# Run the schemathesis container
# echo "running tests for auth service"
# st run --base-url=http://localhost -H "Authorization: $MF_TOKEN" --validate-schema=true ./api/openapi/auth.yml

# echo "running tests for bootstrap service"
# st run --base-url=http://localhost -H "Authorization: $MF_TOKEN" --validate-schema=true ./api/openapi/bootstrap.yml

# echo "running tests for certs service"
# st run --base-url=http://localhost -H "Authorization: $MF_TOKEN" --validate-schema=true ./api/openapi/certs.yml

# echo "running tests for consumers-notifiers service"
# st run --base-url=http://localhost -H "Authorization: $MF_TOKEN" --validate-schema=true ./api/openapi/consumers-notifiers.yml

# echo "running tests for http service"
# st run --base-url=http://localhost -H "Authorization: $MF_TOKEN" --validate-schema=true ./api/openapi/http.yml

# echo "running tests for provision service"
# st run --base-url=http://localhost -H "Authorization: $MF_TOKEN" --validate-schema=true ./api/openapi/provision.yml

# echo "running tests for readers service"
# st run --base-url=http://localhost -H "Authorization: $MF_TOKEN" --validate-schema=true ./api/openapi/readers.yml

# echo "running tests for things service"
# st run --base-url=http://localhost -H "Authorization: $MF_TOKEN" --validate-schema=true ./api/openapi/things.yml

# echo "running tests for twins service"
# st run --base-url=http://localhost -H "Authorization: $MF_TOKEN" --validate-schema=true ./api/openapi/twins.yml

# echo "running tests for users service"
# st run --base-url=http://localhost -H "Authorization: $MF_TOKEN" --validate-schema=true ./api/openapi/users.yml

# echo "running tests for websocket service"
# st run --base-url=http://localhost -H "Authorization: $MF_TOKEN" --validate-schema=true ./api/openapi/websocket.yml


# sudo docker run --network="host" schemathesis/schemathesis:stable \
#     run --base-url=http://localhost ./api/openapi/auth.yml

echo stopping the running containers
sudo docker-compose -f docker/docker-compose.yml down
