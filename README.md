## travisci_exporter

### Reading

can't improve unless we measure

https://blog.petegoo.com/2018/11/09/optimizing-ci-cd-pipelines/

### Notes

https://docs.travis-ci.com/user/developer/
https://github.com/kevinburke/travis/blob/master/lib/travis.go
https://godoc.org/github.com/kevinburke/travis/lib#Job
https://developer.travis-ci.org/resource/builds#Builds


### Configuration

```toml
organizations:
  - name: adamdecaf
    token: "fill-me-in"
  - name: moov-io
    token: "possible-other-token"
```

### Docker

```
docker run adamdecaf/travisci_exporter:latest -config config.toml
```


### Metrics

| Metric Name | Type | Description |
|----|-----|-----|
| `travisci_job_duration_seconds` | Histogram | Histogram of job durations. Buckets: 5s, 10s, 20s, 30s, 60s, 5min, 10min |
