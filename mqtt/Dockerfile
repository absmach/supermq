FROM node:boron-alpine
MAINTAINER Mainflux

COPY . .
RUN npm install

EXPOSE 1883

CMD ["node", "mqtt.js"]
