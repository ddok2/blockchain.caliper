#!/bin/bash
cd network/nuritelecom/exchange-bc-production-v1.0/container_yaml;

docker-compose -f caserver.yaml up -d
docker-compose -f kafka0.yaml up -d
docker-compose -f kafka1.yaml up -d
docker-compose -f kafka2.yaml up -d
docker-compose -f orderer0.yaml up -d
docker-compose -f orderer1.yaml up -d
docker-compose -f peer0.nuriorg.yaml up -d
docker-compose -f peer1.nuriorg.yaml up -d
docker-compose -f peer0.nflexorg.yaml up -d
docker-compose -f peer1.nflexorg.yaml up -d

docker exec -it cli bash
