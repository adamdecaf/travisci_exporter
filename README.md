## travisci_exporter

[![Build Status](https://travis-ci.com/adamdecaf/travisci_exporter.svg?branch=master)](https://travis-ci.com/adamdecaf/travisci_exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/adamdecaf/travisci_exporter)](https://goreportcard.com/report/github.com/adamdecaf/travisci_exporter)
[![Support via Flattr](https://button.flattr.com/flattr-badge-large.png)](https://flattr.com/@adamdecaf)

Prometheus exporter for [TravisCI](https://travis-ci.com/) builds and jobs. Useful for tracking and monitoring your organizations CI/CD timings.

### Metrics

| Metric Name | Type | Description |
|----|-----|-----|
| `travisci_job_duration_seconds` | Gauge | Duration of jobs in seconds. |

### Install / Usage

You can download and run the latest docker image [`adamdecaf/travisci_exporter`](https://hub.docker.com/r/adamdecaf/travisci_exporter/) from the Docker Hub.

Then you can run the Docker image:

```
$ docker run adamdecaf/travisci_exporter:0.2.0 -config.file config.toml
```

### Configuration

travisci_exporter reads a YAML config file like the following, but you'll need to [download an API token](https://travis-ci.com/account/preferences).

```yaml
organizations:
  - name: adamdecaf
    token: "fill-me-in"
    org: true # Required for orgs still on travis-ci.org
  - name: moov-io
    token: "other-token"
```

### Developing / Contributing

If you find a bug, have a question or want more metrics exposed feel free to open either an issue or a Pull Request. I'll try and review it quickly and have it merged.

You can build the sources with `make build`. Currently we require Go 1.11.
