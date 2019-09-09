#!/usr/bin/env bash
cd network/nuritelecom/exchange-bc-production-v1.0/container_yaml;
docker-compose -f caserver.yaml down --volumes --remove-orphans
# docker-compose -f container_yaml/zookeeper0.yaml down  #--volumes --remove-orphans
# docker-compose -f container_yaml/zookeeper0.yaml up -d
# docker-compose -f container_yaml/zookeeper1.yaml down  #--volumes --remove-orphans
# docker-compose -f container_yaml/zookeeper1.yaml up -d
# docker-compose -f container_yaml/zookeeper2.yaml down  #--volumes --remove-orphans
# docker-compose -f container_yaml/zookeeper2.yaml up -d
docker-compose -f kafka0.yaml down --volumes --remove-orphans
docker-compose -f kafka1.yaml down --volumes --remove-orphans
docker-compose -f kafka2.yaml down --volumes --remove-orphans
docker-compose -f orderer0.yaml down --volumes --remove-orphans
docker-compose -f orderer1.yaml down --volumes --remove-orphans
docker-compose -f peer0.nuriorg.yaml down  --volumes --remove-orphans
docker-compose -f peer1.nuriorg.yaml down --volumes --remove-orphans
docker-compose -f peer0.nflexorg.yaml down --volumes --remove-orphans
docker-compose -f peer1.nflexorg.yaml down --volumes --remove-orphans
