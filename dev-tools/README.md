# DEV TOOLS

This is a directory of misc bits that make it easier to work on this code as a developer.

These are not really meant to be used by anyone else.

That said: 

`deploy-to-workshop.sh` is a simple script that makes a build of the handler, then copies
it to the [Sensu Workshop](https://github.com/sensu/sensu-go-workshop) environment, and
generates the appropriate YAML config to install it as an Asset, and sets up a Pipeline/Handler
using that asset. This basically builds/deploys it directly into your local Sensu Workshop with 
a single command.

It assumes you have a directory layout like this:

- `~/github/sensu/sensu-go-workshop` - the workshop repo
- `~/github/thoward/sensu-metrics-stats/handler` - this repo

It also assumes you have `sensuctl` installed on the path, and that there is a Redis instance 
running in your workshop environment (use the `docker-compose-redis.yaml` config in the workshop
repo).

Finally, it assumes you want the linux-amd64 build for the workshop environment. If that's not true you'll need to edit the build targets and YAML embedded in the script.
