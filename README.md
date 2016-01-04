# Operator

![Build status](https://travis-ci.org/wakeful-deployment/operator.svg?branch=master)

Operator is a daemon which listens for changes to consul key/value storage and reacts to those changes. The daemon reacts in the following ways:

* start and stop docker containers
* register and deregister consul services

## More Specifics

The daemon listens to a specific consul key/value namespace "_wakeful/nodes/$NODENAME"where $NODENAME is the name of the node on which Operator runs. This is specified at boot of the daemon via a command line argument.

When a key is added to the "_wakeful/nodes/$NODENAME/apps" namespace, the daemon reacts by starting a docker container named after the last section of the key and using the key's value to determine which container image to used. See below for the struct the value must have. Additionally, a consul ["service"](https://consul.io/docs/agent/services.html) with the same name as the container will be registered.

For example, if "_wakeful/nodes/$NODENAME/myapp" is added, then a docker container will be started named "myapp" and that key's value will be used to determine which image to use. Additionally, a consul service with the name "myapp" will be registered in consul.

## Necessary structure of the consul key's value

The key's value must be equal to image name that will be used to run the docker container. For example, if you want to run the "redis:latest" container image, then "redis:latest" should be the content of the value. Consul automatically base64 encodes all keys' values, and Operator will decode this automatically.

In the future, the structure of the value's contents may change to allow for specification of enviroment, port mappings, etc.

## Bootstrapping

On boot Operator relies on an operator.json file to specify configuration of the node as well as the "global" containers that should always be running on the node. Any cli flag can also be specified in this json file and will be merged into the already passed cli values.

Example:

    {
      "metadata": {
        "location": "eastus",
        "size": "Basic_A1"
      },
      "services": {
        "statsite": {
          "image": "wakeful/wake-statsite:latest",
          "ports": [{
            "incoming": 8125,
            "outgoing": 8125,
            "udp": true
          }],
          "env": {},
          "restart": "always",
          "tags": ["statsd", "udp"]
          "checks": []
        },
        "consul": {
          "image": "wakeful/wake-consul-server:latest",
          "ports": [{
            "incoming": 8300,
            "outgoing": 8300
          }, {
            "incoming": 8301,
            "outgoing": 8301
          }, {
            "incoming": 8301,
            "outgoing": 8301,
            "udp": true
          }, {
            "incoming": 8302,
            "outgoing": 8302,
            "udp": true
          }, {
            "incoming": 8400,
            "outgoing": 8400
          }, {
            "incoming": 8500,
            "outgoing": 8500
          }, {
            "incoming": 8600,
            "outgoing": 8600
          }, {
            "incoming": 8600,
            "outgoing": 8600,
            "udp": true
          }],
          "env": {
            "BOOTSTRAP_EXPECT":"1"
          }
        }
      }
    }
