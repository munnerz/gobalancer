#!/bin/bash

if [ $TRAVIS_PULL_REQUEST != 'false' ] || [ "$TRAVIS_BRANCH" != "master" ] || [ "$TRAVIS_GO_VERSION" != "1.5" ];

  then echo "Skipping pushing docker image to registry as this is a test branch"

else

  docker login -e "$DOCKER_EMAIL" -u "$DOCKER_USERNAME" -p "$DOCKER_PASSWORD"
  docker build -t "munnerz/gobalancer:${TRAVIS_COMMIT:0:7}" .
  docker tag -f "munnerz/gobalancer:${TRAVIS_COMMIT:0:7}" "munnerz/gobalancer:latest"
  docker push "munnerz/gobalancer:${TRAVIS_COMMIT:0:7}"
  docker push "munnerz/gobalancer:latest"

fi
