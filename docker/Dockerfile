# Copyright (c) Abstract Machines
# SPDX-License-Identifier: Apache-2.0

FROM --platform=$BUILDPLATFORM golang:1.25.2-alpine3.22 AS builder
ARG TARGETOS
ARG TARGETARCH
ARG SVC
ARG VERSION
ARG COMMIT
ARG TIME

WORKDIR /go/src/github.com/absmach/supermq
COPY . .
RUN apk update \
    && apk add make upx\
    && GOOS=${TARGETOS} GOARCH=${TARGETARCH} make $SVC\
    && if [ "$TARGETARCH" != "riscv64" ] && [ "$TARGETOS" != "windows" ] && [ "$TARGETOS" != "darwin" ]; then upx build/$SVC; else echo "Skipping UPX for $TARGETOS/$TARGETARCH"; fi\
    && mv build/$SVC /exe

FROM scratch
# Certificates are needed so that mailing util can work.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /exe /
ENTRYPOINT ["/exe"]
