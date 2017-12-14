# spring_exporter for prometheus

This is an exporter for prometheus, written as a [cobra](https://github.com/spf13/cobra) applicationto export the data collected by
the [spring metrics actuator](https://docs.spring.io/spring-boot/docs/current/reference/html/production-ready-metrics.html).

# docker image

There is an automatically build docker image out there: [kalypsocloud/spring_exporter](https://hub.docker.com/r/kalypsocloud/spring_exporter/)

# usage

The exporter is configured using command line flags and arguments. Usage is as follows:

```
$ ./spring_exporter export --help
Exports spring actuator metrics from given endpoint

Usage:
  spring_exporter export <spring-endpoint> [flags]

Flags:
      --basic-auth-password string   HTTP Basic auth password for authentication on the spring endpoint
      --basic-auth-user string       HTTP Basic auth user for authentication on the spring endpoint
  -e, --endpoint string              Path the exporter should listen listen on (default "/metrics")
  -h, --help                         help for export
  -i, --insecure                     Whether to use insecure https mode, i.e. skip ssl cert validation (only useful with https endpoint)
  -l, --listen string                Host/Port the exporter should listen listen on (default ":9321")

Global Flags:
      --config string   config file (default is $HOME/.spring_exporter.yaml)
```

Example usage in a docker-compose file:

```yaml
version: "3"
services:
  fixtures:
    build:
      context: fixtures
    image: kalypsocloud/spring_exporter_test_server
    ports:
      - "3000:3000"
  exporter:
    build:
      context: .
    image: kalypsocloud/spring_exporter
    command: [
      "export", "http://fixtures:3000/manage/metrics",
      "--basic-auth-user", "admin",
      "--basic-auth-password", "secret",
      "--insecure"
    ]
    links:
      - fixtures
    ports:
      - "9321:9321"
```

# license

MIT License

Copyright (c) 2017 KalypsoCloud GmbH
