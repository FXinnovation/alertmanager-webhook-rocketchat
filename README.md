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
alertmanager-webhook-rocketchat Rocket.Chat endpoint, credentials, severity_colors, and channel are configured in a yml file, as shown in the example below. Note that fields ``severity_colors`` and ``channel`` are **optional**. More severity color mappings can be added.

```
endpoint:
  scheme: "https"
  host: "<host.url>"
credentials:
  name: "<user>"
  email: "<user@local.local>"
  password: "<password>"
severity_colors:
  warning: "<warning_color_hexcode>"
  critical: "<critical_color_hexcode>"
channel:
  default_channel_name: "<default_channel_name>"
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
## Building
Build the sources with 
```bash
make build
```
**Note**: As this is a go build you can use _GOOS_ and _GOARCH_ environment variables to build for another platform.
### Crossbuilding
The _Makefile_ contains a _crossbuild_ target which builds all the platforms defined in _.promu.yml_ file and puts the files in _.build_ folder. Alternatively you can specify one platform to build with the OSARCH environment variable;
```bash
OSARCH=linux/amd64 make crossbuild
```
## Docker image

To run alertmanager-webhook-rocketchat on Docker, you can use the [fxinnovation/alertmanager-webhook-rocketchat](https://hub.docker.com/r/fxinnovation/alertmanager-webhook-rocketchat) image. 

You can also build a docker image from sources using:
```bash
make docker
```
The resulting image is named `fxinnovation/alertmanager-webhook-rocketchat:{git-branch}`.
It exposes port 9876 and expects the config in /config/rocketchat.yml. To configure it, you can bind-mount a config from your host: 
```
$ docker run -p 9876 -v /path/on/host/config/rocketchat.yml:/config/rocketchat.yml fxinnovation/alertmanager-webhook-rocketchat:master
```
## Releasing
The _release_tag_ and _release_docker_ targets are respectively creating and pushing a git tag and creating and pushing a docker image using the VERSION.txt file content as tag name.

## Contributing

Refer to [CONTRIBUTING.md](https://github.com/FXinnovation/alertmanager-webhook-rocketchat/blob/master/CONTRIBUTING.md).

## License

Apache License 2.0, see [LICENSE](https://github.com/FXinnovation/alertmanager-webhook-rocketchat/blob/master/LICENSE).
