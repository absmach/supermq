ws://localhost:8186/channels/d6a43e46-d204-4c49-a412-3a9f0fe189b3/messages?authorization=37f3241b-ef60-48a1-bfac-1585d8f9992f

ws://localhost:8186/channels/d6a43e46-d204-4c49-a412-3a9f0fe189b3/messages?authorization=6f11618a-f245-487c-bd88-dab2e791483a

curl -sSi -X POST -H "Content-Type: application/senml+json" -H "Authorization: Thing 37f3241b-ef60-48a1-bfac-1585d8f9992f" http://localhost:8008/channels/d6a43e46-d204-4c49-a412-3a9f0fe189b3/messages -d '[{"bn":"demo", "bu":"A","bver":5, "n":"voltage","u":"V","v":120.1}]'
coap-cli post channels/d6a43e46-d204-4c49-a412-3a9f0fe189b3/messages -auth 37f3241b-ef60-48a1-bfac-1585d8f9992f -d '[{"bn":"demo", "bu":"A","bver":5, "n":"voltage","u":"V","v":120.1}]'
mosquitto_pub -I mainflux -u 53b3b749-338a-438a-93f5-6ad8ea95df49 -P 37f3241b-ef60-48a1-bfac-1585d8f9992f -t channels/d6a43e46-d204-4c49-a412-3a9f0fe189b3/messages -h localhost -m '[{"bn":"demo", "bu":"A","bver":5, "n":"voltage","u":"V","v":120.1}]'


mosquitto_sub -I mainflux -u 53b3b749-338a-438a-93f5-6ad8ea95df49 -P 37f3241b-ef60-48a1-bfac-1585d8f9992f -t channels/d6a43e46-d204-4c49-a412-3a9f0fe189b3/messages -h localhost 

coap-cli get channels/d6a43e46-d204-4c49-a412-3a9f0fe189b3/messages -auth 37f3241b-ef60-48a1-bfac-1585d8f9992f -o


max=100000;
for (( i=1; i <= max; ++i )); do
   timestamp=$(date +%s%N | cut -b1-13)
   payload='[{"bn":"present","bt":'$timestamp', "bu":"A","bver":5, "n":"voltage","u":"V","v":'$i'}]'
   mosquitto_pub -I mainflux -u 53b3b749-338a-438a-93f5-6ad8ea95df49 -P 37f3241b-ef60-48a1-bfac-1585d8f9992f -t channels/d6a43e46-d204-4c49-a412-3a9f0fe189b3/messages -h localhost -m "$payload"
  #  curl -sSi -X POST -H "Content-Type: application/senml+json" -H "Authorization: Thing 37f3241b-ef60-48a1-bfac-1585d8f9992f" http://localhost/http/channels/d6a43e46-d204-4c49-a412-3a9f0fe189b3/messages -d "$payload"
done

curl -sSi  -H "Authorization: Thing 37f3241b-ef60-48a1-bfac-1585d8f9992f" http://localhost:9003/channels/d6a43e46-d204-4c49-a412-3a9f0fe189b3/messages?offset=0&limit=5
curl -sSi  -H "Authorization: Thing 37f3241b-ef60-48a1-bfac-1585d8f9992f" http://localhost:9005/channels/d6a43e46-d204-4c49-a412-3a9f0fe189b3/messages?offset=0&limit=5
curl -sSi  -H "Authorization: Thing 37f3241b-ef60-48a1-bfac-1585d8f9992f" http://localhost:9007/channels/d6a43e46-d204-4c49-a412-3a9f0fe189b3/messages?offset=0&limit=5
curl -sSi  -H "Authorization: Thing 37f3241b-ef60-48a1-bfac-1585d8f9992f" http://localhost:9009/channels/d6a43e46-d204-4c49-a412-3a9f0fe189b3/messages?offset=0&limit=5
curl -sSi  -H "Authorization: Thing 37f3241b-ef60-48a1-bfac-1585d8f9992f" http://localhost:9011/channels/d6a43e46-d204-4c49-a412-3a9f0fe189b3/messages?offset=0&limit=5&protocol=ws

max=100;
for (( i=1; i <= max; ++i )); do
   go run tools/e2e/cmd/main.go -n 90 
done


KEYWORD='    container_name: mainflux-mqtt
    depends_on:
      - things
      - nats
    restart: on-failure'
REPLACE='    container_name: mainflux-mqtt
    depends_on:
      - things
      - nats
      - vernemq
    restart: on-failure'
ESCAPED_KEYWORD=$(printf '%s\n' "$KEYWORD" | sed -z 's/[]\/$*.^[]/\\&/g; s/\n/\\n/g');
ESCAPED_REPLACE=$(printf '%s\n' "$REPLACE" | sed -z 's/[\/&]/\\&/g; s/\n/\\n/g')
sed -z "s/$ESCAPED_KEYWORD/$ESCAPED_REPLACE/" docker/docker-compose.yml > docker/docker-compose.yml.new
mv docker/docker-compose.yml.new docker/docker-compose.yml


curl -sSiX POST http://localhost:9000/connect -H "Content-Type: application/json" -H "Authorization: Bearer $T" -d @- << EOF
{"thing_id":"ee4c8403-1aad-4619-9dc7-333d2a8bcaeb","channel_id":"9c612971-cc25-491d-b765-3276eafeeaa8","permission":"publish"}
EOF

