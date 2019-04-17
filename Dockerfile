FROM golang:1.12 as builder
WORKDIR /go/src/github.com/FXinnovation/alertmanager-webhook-rocketchat
COPY . .
RUN make build

FROM quay.io/prometheus/busybox:glibc AS app
LABEL maintainer="FXinnovation CloudToolDevelopment <CloudToolDevelopment@fxinnovation.com>"
COPY --from=builder /go/src/github.com/FXinnovation/alertmanager-webhook-rocketchat/alertmanager-webhook-rocketchat /bin/alertmanager-webhook-rocketchat

EXPOSE      9876

WORKDIR /
ENTRYPOINT  [ "/bin/alertmanager-webhook-rocketchat" ]
