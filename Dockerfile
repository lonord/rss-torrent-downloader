FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

COPY . /app

WORKDIR /app

RUN apk add git make
RUN make $TARGETOS/$TARGETARCH
RUN cd build && mv rss-torrent-dl_* rss-torrent-dl

FROM alpine:3.19

COPY --from=builder /app/build/rss-torrent-dl /app/rss-torrent-dl

RUN apk add ca-certificates \
    && rm -rf /var/cache/apk/* \
    && update-ca-certificates

WORKDIR /app

ENV TZ=Asia/Shanghai
EXPOSE 6900
VOLUME [ "/subscription" ]

CMD [ "/app/rss-torrent-dl", "-subscription", "/subscription", "-http", ":6900" ]