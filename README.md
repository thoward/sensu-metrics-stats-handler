[![Sensu Bonsai Asset](https://img.shields.io/badge/Bonsai-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/thoward/sensu-metrics-stats-handler)
![goreleaser](https://github.com/thoward/sensu-metrics-stats-handler/workflows/goreleaser/badge.svg)
[![Go Test](https://github.com/thoward/sensu-metrics-stats-handler/workflows/Go%20Test/badge.svg)](https://github.com/thoward/sensu-metrics-stats-handler/actions?query=workflow%3A%22Go+Test%22)
[![goreleaser](https://github.com/thoward/sensu-metrics-stats-handler/workflows/goreleaser/badge.svg)](https://github.com/thoward/sensu-metrics-stats-handler/actions?query=workflow%3Agoreleaser)

# sensu-metrics-stats-handler

## Table of Contents
- [Overview](#overview)
- [Requirements](#requirements)
- [Usage examples](#usage-examples)
- [Configuration](#configuration)
  - [Asset registration](#asset-registration)
  - [Handler definition](#handler-definition)
- [Installation from source](#installation-from-source)
- [Contributing](#contributing)

## Overview

The **sensu-metrics-stats-handler** is a [Sensu Handler][sensu-handler] that calculates stats about metric points, storing those in a Redis database. It also re-publishes the metric events via Redis pub/sub so that they can be consumed by other tools (like [metrics-top]).

## Requirements

- A **[Redis] instance**, reachable from the Sensu Backend that will be running this handler. This will be used to track statistics over time, and as the pub/sub engine for event publishing.

## Usage examples

```
sensu-metrics-stats-handler --host redis --port 6379 --password "sensu"
```

## Configuration

The handler accepts the following command-line switches:

- `--host <address>`: the Redis server address (default: `localhost`)
- `--port <number>`: the Redis server port (default: `6379`)
- `--password <string>`: the Redis server password (default: _none_) 

### Asset registration

[Sensu Assets][assets-docs] are the best way to make use of this plugin. If you're not using an asset, please
consider doing so! If you're using sensuctl 5.13 with Sensu Backend 5.13 or later, you can use the
following command to add the asset:

```
sensuctl asset add thoward/sensu-metrics-stats-handler
```

If you're using an earlier version of sensuctl, you can find the asset on the [Bonsai Asset Index][sensu-metric-stats-handler-bonsai].

### Handler definition

```yml
---
type: Handler
api_version: core/v2
metadata:
  created_by: sensu
  labels:
    sensu.io/managed_by: sensuctl
  name: metrics-stats
  namespace: default
spec:
  command: sensu-metrics-stats-handler --host redis --port 6379 --password "sensu"
  filters:
  - has_metrics
  runtime_assets:
  - thoward/sensu-metrics-stats-handler:0.0.1
  timeout: 10
  type: pipe
```

## Installation from source

The preferred way of installing and deploying this plugin is to use it as an Asset. If you would
like to compile and install the plugin from source or contribute to it, download the latest version
or create an executable script from this source.

From the local path of the `sensu-metrics-stats-handler` repository:

```
go build
```

## Contributing

For more information about contributing to this plugin, see [Contributing].

## Releases with Github Actions

To release a version of this project, tag the target sha with a semver release without a `v`
prefix (ex. `1.0.0`). This will trigger the [GitHub action][github-action] workflow to [build and release][workflow-release]
the plugin with goreleaser. Register the asset with [Bonsai] to share it with the community!

[contributing]: https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md
[workflow-release]: https://github.com/sensu/handler-plugin-template/blob/master/.github/workflows/release.yml
[github-action]: https://github.com/sensu/handler-plugin-template/actions
[sensu-handler]: https://docs.sensu.io/sensu-go/latest/reference/handlers/
[bonsai]: https://bonsai.sensu.io/
[sensu-metrics-stats-handler-bonsai]: https://bonsai.sensu.io/assets/thoward/sensu-metrics-stats-handler
[assets-docs]: https://docs.sensu.io/sensu-go/latest/reference/assets/
[metrics-top]: https://github.com/thoward/metrics-top
[redis]: https://redis.io/
