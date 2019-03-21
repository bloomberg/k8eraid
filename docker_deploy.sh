#!/bin/bash

docker login -u "$DOCKER_USER" -p "$DOCKER_PASSWORD" \
  && make pushcontainer
