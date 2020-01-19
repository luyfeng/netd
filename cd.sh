#!/bin/bash

set -o errexit -o nounset
TRAVIS_BRANCH=master
if [ "$TRAVIS_BRANCH" != "master" ]
then 
	    echo "This commit was made against the $TRAVIS_BRANCH and not the master! No deploy!" 
		    exit 0
fi

rev=$(git rev-parse --short HEAD)
docker login -u docker-image-builder -p $HUB hub.sky-cloud.net
docker build -t hub.sky-cloud.net/nap2/netd:${TRAVIS_BRANCH}_build-${rev} .
docker push hub.sky-cloud.net/nap2/netd:${TRAVIS_BRANCH}_build-${rev}


