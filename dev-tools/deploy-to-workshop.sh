#!/bin/bash

cd ~/github/thoward/sensu-metrics-stats-handler
goreleaser release --snapshot --rm-dist

asset_hash="`shasum -a 512 dist/sensu-metrics-stats-handler_0.0.0-SNAPSHOT-none_linux_amd64.tar.gz | cut -d' ' -f1`"

cp -f "dist/sensu-metrics-stats-handler_0.0.0-SNAPSHOT-none_linux_amd64.tar.gz" ~/github/sensu/sensu-go-workshop/assets/

pushd .

cd ~/github/sensu/sensu-go-workshop
source .envrc

cat << EOF | sensuctl create
---
type: Asset
api_version: core/v2
metadata:
  name: sensu-metrics-stats-handler
  labels: 
  annotations:
spec:
  builds: 
  - url: http://sensu-assets/assets/sensu-metrics-stats-handler_0.0.0-SNAPSHOT-none_linux_amd64.tar.gz
    sha512: ${asset_hash}
    filters:
    - entity.system.os == 'linux'
    - entity.system.arch == 'amd64'
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
  env_vars: null
  filters:
  - has_metrics
  handlers: null
  runtime_assets:
  - sensu-metrics-stats-handler
  secrets: null
  timeout: 10
  type: pipe
---
type: Pipeline
api_version: core/v2
metadata:
  name: metrics
spec:
  workflows:
  - name: metrics_workflow
    filters:
    - name: has_metrics
      type: EventFilter
      api_version: core/v2
    handler:
      name: metrics-stats
      type: Handler
      api_version: core/v2
EOF

popd
