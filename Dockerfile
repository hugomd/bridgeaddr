FROM golang:1.17-alpine AS build
ENV GO111MODULE=on
WORKDIR /go/src/github.com/fiatjaf/bridgeaddr/
COPY . /go/src/github.com/fiatjaf/bridgeaddr/
RUN cd /go/src/github.com/fiatjaf/bridgeaddr && \
    go get && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
COPY entrypoint.sh /entrypoint.sh
RUN apk add ca-certificates tor
ENTRYPOINT ["/entrypoint.sh"]
