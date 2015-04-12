## Design decisions

Goal: Match Heroku's simplicity - this shouldn't be as complex to use
as fleet or Kubernetes. Instead, they are both possible backends for this.

Add a UI later on that makes this nicer to manage, but start with a command-line
tool that accesses a JSON API.

### Milestone 1: Replace discovery sidekick

* [ ] Make sure etcd setup works
* [ ] Solve global state issue (who owns the docker client connection?)
* [ ] Modify nginx-lb to be cerebro-lb (read settings correctly, use release version)
* [ ] Adjust fleet control files

Deploy this to production (with correct naming) and run cerebro-lb at port 8080 for testing.

### Milestone 2: Deployment flow

...

### etcd layout

/cerebro/APPNAME/config/INSTANCE_TYPE/count
/cerebro/APPNAME/config/INSTANCE_TYPE/environment/KEY
/cerebro/APPNAME/config/INSTANCE_TYPE/lb/hostnames
/cerebro/APPNAME/config/INSTANCE_TYPE/lb/ssl

/cerebro/APPNAME/releases/VERSION/state (starting/running/shutdown)
/cerebro/APPNAME/releases/VERSION/instances/INSTANCE_TYPE/INSTANCE_NUMBER
/cerebro/APPNAME/releases/VERSION/docker_image

fleet name: APPNAME@PROC_TYPE-VERSION-INSTANCE_NUMBER (e.g. pga-staging@web-v214-1)
docker name: APPNAME-PROC_TYPE-VERSION-INSTANCE_NUMBER (e.g. pga-staging-web-v214-1)

// Keep a log in etcd of the previous 4 releases and their docker IDs, so we can rollback if necessary


### Thoughs

// It is probably sensible to have one cerebro instance running on every CoreOS host (data kept in etcd)
// cerebro is also responsible for syncing docker data into etcd (LB data)
// cerebro manages the number of instances, i.e. using etcd makes sure the right number of instances is running across all hosts
// optionally, cerebro API makes itself available using the load balancer - or you can conncet to any CoreOS host to access it
