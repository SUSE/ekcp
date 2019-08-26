#!/bin/bash

docker build -t gorouter ./
docker run -d -p 4222:4222 -p 6222:6222 -p 8222:8222  --name nats --restart=always -ti nats:latest
docker run --link nats:nats --name gorouter -p 8081:8081 -p 8082:8082 -t -d gorouter:latest
