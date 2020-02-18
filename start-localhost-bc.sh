#!/bin/bash

./network/nuritelecom/exchange-bc-production-v1.0/start.sh;
docker exec cli bash -c '/opt/gopath/src/github.com/hyperledger/fabric/peer/scripts/newstart.sh exchange-channel 1 10';
