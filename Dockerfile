FROM golang:1.17-alpine AS build-env
ENV GO111MODULE=on
WORKDIR /go/src/github.com/fiatjaf/bridgeaddr/
RUN apk add ca-certificates
COPY . /go/src/github.com/fiatjaf/bridgeaddr/
RUN cd /go/src/github.com/fiatjaf/bridgeaddr && \
    go get && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM scratch
COPY --from=build-env /go/src/github.com/fiatjaf/bridgeaddr/main /
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
CMD ["/main"]
