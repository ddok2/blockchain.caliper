rm -rf kafka0
rm -rf kafka1
rm -rf kafka2
rm -rf kafka3
rm -rf couchdb0
rm -rf couchdb1
rm -rf couchdb2
rm -rf couchdb3
rm -rf orderer0.exchange.com
rm -rf orderer1.exchange.com
rm -rf orderer2.exchange.com
rm -rf peer0.nuriorg.exchange.com
rm -rf peer1.nuriorg.exchange.com
rm -rf peer0.nflexorg.exchange.com
rm -rf peer1.nflexorg.exchange.com

docker-compose -f bc1.yaml down  #--volumes --remove-orphans
docker-compose -f bc1.yaml up -d
docker-compose -f bc2.yaml down  #--volumes --remove-orphans
docker-compose -f bc2.yaml up -d
docker-compose -f zookeeper0.yaml down  #--volumes --remove-orphans
docker-compose -f zookeeper0.yaml up -d
docker-compose -f zookeeper2.yaml down  #--volumes --remove-orphans
docker-compose -f zookeeper2.yaml up -d
docker-compose -f kafka0.yaml down  #--volumes --remove-orphans
docker-compose -f kafka0.yaml up -d
docker-compose -f kafka2.yaml down  #--volumes --remove-orphans
docker-compose -f kafka2.yaml up -d
docker-compose -f orderer0.yaml down  #--volumes --remove-orphans
docker-compose -f orderer0.yaml up -d
docker-compose -f orderer1.yaml down  #--volumes --remove-orphans
docker-compose -f orderer1.yaml up -d
docker-compose -f orderer2.yaml down  #--volumes --remove-orphans
docker-compose -f orderer2.yaml up -d
docker-compose -f peer0.nuriorg.yaml down  #--volumes --remove-orphans
docker-compose -f peer0.nuriorg.yaml up -d
docker-compose -f peer1.nuriorg.yaml down  #--volumes --remove-orphans
docker-compose -f peer1.nuriorg.yaml up -d
docker-compose -f peer0.nflexorg.yaml down  #--volumes --remove-orphans
docker-compose -f peer0.nflexorg.yaml up -d
docker-compose -f peer1.nflexorg.yaml down #--volumes --remove-orphans
docker-compose -f peer1.nflexorg.yaml up -d

docker exec -it cli bash
