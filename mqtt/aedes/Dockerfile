FROM node:10.15.1-alpine

COPY *.proto mqtt/* ./

RUN apk --no-cache add --virtual builds-deps build-base python
RUN npm rebuild && npm install
RUN apk del builds-deps

EXPOSE 1883 8880

CMD ["node", "mqtt.js"]
