#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#
# Exit on first error, print all commands.
set -e

help()
{
    echo "Usage: $0 [command]"
}

if [ $# -ne 1 ]
then
    help
    exit 0
fi

CLEAN=$1

if [ ${CLEAN} == "clean" ]; then
    echo "Clean teardown"
    cd ../volumes
    rm -rf peer*.*.exchange.com
    rm -rf zookeeper*_*
    rm -rf kafka*_*
    rm -rf couchdb*
    rm -rf orderer*.exchange.com    
fi

docker stop $(docker ps -aq)
docker rm $(docker ps -aq)
docker rmi $(docker images dev-* -q)


