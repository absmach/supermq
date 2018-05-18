FROM node:boron-alpine
MAINTAINER Mainflux

COPY . .
RUN npm install

EXPOSE 1883
EXPOSE 8880

CMD ["node", "mqtt.js"]
