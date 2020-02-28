#!/bin/bash
source ./scripts/setenv.sh

echo " ███╗   ██╗██╗   ██╗██████╗ ██╗    ██████╗ ██╗      ██████╗  ██████╗██╗  ██╗ ██████╗██╗  ██╗ █████╗ ██╗███╗   ██╗"
echo " ████╗  ██║██║   ██║██╔══██╗██║    ██╔══██╗██║     ██╔═══██╗██╔════╝██║ ██╔╝██╔════╝██║  ██║██╔══██╗██║████╗  ██║"
echo " ██╔██╗ ██║██║   ██║██████╔╝██║    ██████╔╝██║     ██║   ██║██║     █████╔╝ ██║     ███████║███████║██║██╔██╗ ██║"
echo " ██║╚██╗██║██║   ██║██╔══██╗██║    ██╔══██╗██║     ██║   ██║██║     ██╔═██╗ ██║     ██╔══██║██╔══██║██║██║╚██╗██║"
echo " ██║ ╚████║╚██████╔╝██║  ██║██║    ██████╔╝███████╗╚██████╔╝╚██████╗██║  ██╗╚██████╗██║  ██║██║  ██║██║██║ ╚████║"
echo " ╚═╝  ╚═══╝ ╚═════╝ ╚═╝  ╚═╝╚═╝    ╚═════╝ ╚══════╝ ╚═════╝  ╚═════╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝╚═╝  ╚═══╝"


docker-compose -f container_yaml/caserver.yaml down  #--volumes --remove-orphans
docker-compose -f container_yaml/kafka0.yaml down  #--volumes --remove-orphans
docker-compose -f container_yaml/kafka1.yaml down  #--volumes --remove-orphans
docker-compose -f container_yaml/kafka2.yaml down  #--volumes --remove-orphans
docker-compose -f container_yaml/orderer0.yaml down  #--volumes --remove-orphans
docker-compose -f container_yaml/orderer1.yaml down  #--volumes --remove-orphans
docker-compose -f container_yaml/peer0.nuriorg.yaml down  #--volumes --remove-orphans
docker-compose -f container_yaml/peer1.nuriorg.yaml down  #--volumes --remove-orphans
docker-compose -f container_yaml/peer0.nflexorg.yaml down  #--volumes --remove-orphans
docker-compose -f container_yaml/peer1.nflexorg.yaml down #--volumes --remove-orphans

docker-compose -f container_yaml/caserver.yaml up -d
# docker-compose -f container_yaml/zookeeper0.yaml down  #--volumes --remove-orphans
# docker-compose -f container_yaml/zookeeper0.yaml up -d
# docker-compose -f container_yaml/zookeeper1.yaml down  #--volumes --remove-orphans
# docker-compose -f container_yaml/zookeeper1.yaml up -d
# docker-compose -f container_yaml/zookeeper2.yaml down  #--volumes --remove-orphans
# docker-compose -f container_yaml/zookeeper2.yaml up -d
docker-compose -f container_yaml/kafka0.yaml up -d
docker-compose -f container_yaml/kafka1.yaml up -d
docker-compose -f container_yaml/kafka2.yaml up -d
docker-compose -f container_yaml/orderer0.yaml up -d
docker-compose -f container_yaml/orderer1.yaml up -d
docker-compose -f container_yaml/peer0.nuriorg.yaml up -d
docker-compose -f container_yaml/peer1.nuriorg.yaml up -d
docker-compose -f container_yaml/peer0.nflexorg.yaml up -d
docker-compose -f container_yaml/peer1.nflexorg.yaml up -d

docker exec -it cli bash
