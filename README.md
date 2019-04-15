# alertmanager-webhook-rocketchat
[![Build Status](https://travis-ci.org/FXinnovation/alertmanager-webhook-rocketchat.svg?branch=master)](https://travis-ci.org/FXinnovation/alertmanager-webhook-rocketchat)

Prometheus AlertManager webhook receiver to [Rocket.Chat](https://rocket.chat/), written in Go.

The goal of this project is to provide a standard component to send alerts from the AlertManager to the 
[Rocket.Chat](https://rocket.chat/) team communication tool.

## Getting Started

### Prerequisites

To run this project, you will need a [working Go environment](https://golang.org/doc/install).

### Installing

```bash
go get -u github.com/FXinnovation/alertmanager-webhook-rocketchat
```

## Running the tests

```bash
make test
```

## Usage

```bash
./alertmanager-webhook-rocketchat -h
```

## Deployment

The project takes 2 optional parameters to be configured :
- config.file to specify RocketChat configuration. Cf config/rocketchat_example.yml (default : config/rocketchat.yml)
- listen.address to specify the listening port (default : 9876)

Configuration is done at three levels: alertmanager-webhook-rocketchat, AlertManager, and Prometheus server.

### alertmanager-webhook-rocketchat config
alertmanager-webhook-rocketchat Rocket.Chat endpoint and credentials are configured in a yml file, i.e.:

```
endpoint:
  scheme: "https"
  host: "<host.url>"
credentials:
  name: "<user>"
  email: "<user@local.local>"
  password: "<password>"
```

### AlertManager config
In the AlertManger config (e.g., alertmanager.yml), a `webhook_configs` target the alertmanager-webhook-rocketchat URL, e.g.:

```
global:
  resolve_timeout: 5m

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 10m
  receiver: 'rocketchat'

receivers:
- name: 'rocketchat'
  webhook_configs:
  - url: "http://localhost:9876/webhook"
    send_resolved: true
```

### Prometheus rules config
In the Prometheus server rules files, alerts defines `channel_name` and `severity`, e.g.:

```
groups:
- name: example
  rules:
  # Alert for any instance that is unreachable for >1 minutes.
  - alert: InstanceDown
    expr: up == 0
    for: 1m
    labels:
      severity: critical
      channel_name: "prometheus-test-room"
    annotations:
      summary: "Instance {{ $labels.instance }} down"
      description: "{{ $labels.instance }} of job {{ $labels.job }} has been down for more than 1 minutes."
```

## Contributing

Refer to [CONTRIBUTING.md](https://github.com/FXinnovation/alertmanager-webhook-rocketchat/blob/master/CONTRIBUTING.md).

## License

Apache License 2.0, see [LICENSE](https://github.com/FXinnovation/alertmanager-webhook-rocketchat/blob/master/LICENSE).
