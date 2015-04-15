## Design decisions

Goal: Match Heroku's simplicity - this shouldn't be as complex to use
as fleet or Kubernetes. Instead, they are both possible backends for this.

Add a UI later on that makes this nicer to manage, but start with a command-line
tool that accesses a JSON API.

### Milestone 1: Replace discovery sidekick

* [ ] Modify nginx-lb to be cerebro-lb (read settings correctly, use release version)

### Milestone 2: Deployment flow

* [ ] Figure out where to read out app names / current release
* [ ] Health checks
* [ ] State management
* [ ] Switch release version when target version is state running
* [ ] Fleet management (starting/stopping instances)

### Milestone 3: Useful things

* [ ] Command line tool to show current state
* [ ] Easy changing of target instance count (using CLI)
* ...

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


### Thoughts

// It is probably sensible to have one cerebro instance running on every CoreOS host (data kept in etcd)
// cerebro is also responsible for syncing docker data into etcd (LB data)
// cerebro manages the number of instances, i.e. using etcd makes sure the right number of instances is running across all hosts
// optionally, cerebro API makes itself available using the load balancer - or you can conncet to any CoreOS host to access it
