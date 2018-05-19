FROM node:boron-alpine
MAINTAINER Mainflux

COPY *.proto ./
COPY mqtt/ mqtt/

WORKDIR mqtt/

RUN npm rebuild && npm install

EXPOSE 1883
EXPOSE 8880

CMD ["node", "mqtt.js"]
