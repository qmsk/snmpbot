#!/bin/bash -x

set -ue

VERSION=${TRAVIS_TAG#v}

echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin

docker tag qmsk/snmpbot qmsk/snmpbot:$VERSION

docker push qmsk/snmpbot:latest
docker push qmsk/snmpbot:$VERSION
