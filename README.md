# [![Build Status](https://travis-ci.org/mudler/ekcp.svg?branch=master)](https://travis-ci.org/mudler/ekcp) Ekcp (Ephemeral Kubernetes Clusters Provider)

EKCP aims to build a simple API to provide Kubernetes environment for development and :rocket: lab environments.

# Requires

- Docker and docker-compose on the host

## Deploy with docker-compose

    $> git clone https://github.com/mudler/ekcp
    $> cd ekcp
    $> vim docker-compose.yml # Edit DOMAIN (pick one, reccomend to xip.io or nip.io) and KUBEHOST (your external IP)
    $> docker-compose up -d

## Simple API to create ephemeral clusters

### Create a new cluster

    curl -d "name=test" -X POST http://127.0.0.1:8030/new

### Delete a cluster

    curl -X DELETE http://127.0.0.1:8030/test

### Get a cluster kubeconfig file

    curl  http://127.0.0.1:8030/kubeconfig/test

## Architecture

EKCP currently uses ```kind``` as backend to create new Kubernetes cluster. A proxy is setted up for each cluster to allow remote connection leveraging ```kubectl proxy```. Gorouter is setted up with docker-compose and the routes are registered to a NATS server if ```ROUTE_REGISTER=true``` is set, allowing to use the gorouter as http proxy to resolve internal domains.
