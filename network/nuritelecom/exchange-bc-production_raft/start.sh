#!/bin/bash
source ./scripts/setenv.sh

echo " ███╗   ██╗██╗   ██╗██████╗ ██╗    ██████╗ ██╗      ██████╗  ██████╗██╗  ██╗ ██████╗██╗  ██╗ █████╗ ██╗███╗   ██╗"
echo " ████╗  ██║██║   ██║██╔══██╗██║    ██╔══██╗██║     ██╔═══██╗██╔════╝██║ ██╔╝██╔════╝██║  ██║██╔══██╗██║████╗  ██║"
echo " ██╔██╗ ██║██║   ██║██████╔╝██║    ██████╔╝██║     ██║   ██║██║     █████╔╝ ██║     ███████║███████║██║██╔██╗ ██║"
echo " ██║╚██╗██║██║   ██║██╔══██╗██║    ██╔══██╗██║     ██║   ██║██║     ██╔═██╗ ██║     ██╔══██║██╔══██║██║██║╚██╗██║"
echo " ██║ ╚████║╚██████╔╝██║  ██║██║    ██████╔╝███████╗╚██████╔╝╚██████╗██║  ██╗╚██████╗██║  ██║██║  ██║██║██║ ╚████║"
echo " ╚═╝  ╚═══╝ ╚═════╝ ╚═╝  ╚═╝╚═╝    ╚═════╝ ╚══════╝ ╚═════╝  ╚═════╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝╚═╝  ╚═══╝"


docker-compose -f container_yaml/caserver.yaml down  #--volumes --remove-orphans
docker-compose -f container_yaml/orderer0.yaml down  #--volumes --remove-orphans
docker-compose -f container_yaml/orderer1.yaml down  #--volumes --remove-orphans
docker-compose -f container_yaml/orderer2.yaml down  #--volumes --remove-orphans
docker-compose -f container_yaml/peer0.nuriorg.yaml down  #--volumes --remove-orphans
docker-compose -f container_yaml/peer1.nuriorg.yaml down  #--volumes --remove-orphans
docker-compose -f container_yaml/peer0.nflexorg.yaml down  #--volumes --remove-orphans
docker-compose -f container_yaml/peer1.nflexorg.yaml down #--volumes --remove-orphans

docker-compose -f container_yaml/caserver.yaml up -d
docker-compose -f container_yaml/orderer0.yaml up -d
docker-compose -f container_yaml/orderer1.yaml up -d
docker-compose -f container_yaml/orderer2.yaml up -d
docker-compose -f container_yaml/peer0.nuriorg.yaml up -d
docker-compose -f container_yaml/peer1.nuriorg.yaml up -d
docker-compose -f container_yaml/peer0.nflexorg.yaml up -d
docker-compose -f container_yaml/peer1.nflexorg.yaml up -d

docker exec -it cli bash
