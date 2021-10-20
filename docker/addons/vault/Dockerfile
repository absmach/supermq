# base image
FROM alpine:3.7

# set vault version
ENV VAULT_VERSION 1.6.2

# create a new directory
RUN mkdir /vault

# download dependencies
RUN apk --no-cache add \
      bash \
      at \
      ca-certificates \
      wget \
      jq

# add wait program
ADD https://github.com/ufoscout/docker-compose-wait/releases/download/2.9.0/wait /wait
RUN chmod +x /wait

# download and set up vault
RUN wget --quiet --output-document=/tmp/vault.zip https://releases.hashicorp.com/vault/${VAULT_VERSION}/vault_${VAULT_VERSION}_linux_amd64.zip && \
    unzip /tmp/vault.zip -d /vault && \
    rm -f /tmp/vault.zip && \
    chmod +x /vault


# update PATH
ENV PATH="PATH=$PATH:$PWD/vault"

# expose port 8100
EXPOSE 8100

# run vault
ENTRYPOINT ["bash"]
