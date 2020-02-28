
# 1. 블록체인 구성

블록체인 구성은 다음과 같다.

  AWS 인스턴스  |  역할               
  :------:|:------
  EC01    |   CA 노드
  EC02    |   Kafka, Zookeeper
  EC03    |   Kafka, Zookeeper
  EC04    |   Kafka, Zookeeper
  EC05    |   Orderer 노드
  EC06    |   Orderer 노드
  EC07    |   CA 노드, Peer 노드
  EC08    |   Peer 노드
  EC09    |   CA 노드, Peer 노드
  EC10    |   Peer 노드
  EC11    |   TxBooster 서버
  EC12    |   Admin Tool
  EC13    |   Monitoring Tool

# 2. 블록체인 서버 실행
서버 시작 순서: 인스턴스 번호대로 시작하면 된다. (ex. EC1 시작 후 EC2 시작)

### 2.1. CA 노드(ca.exchange.com) 시작 
1) *EC01* SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2) start.sh 실행

```bash
$ ./start.sh # CA 노드 시작
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

- /home/ubuntu/start.sh 없을 시
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f caserver.yaml down
$ docker-compose -f caserver.yaml up -d  # CA 노드 시작
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

### 2.2. Kafka0, Zookeeper0 시작
1) **EC02** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2) start.sh 실행
```bash
$ ./start.sh
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

- /home/ubuntu/start.sh 없을 시
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f kafka0.yaml down
$ docker-compose -f kafka0.yaml up -d  # Kafka,Zookeeper 노드 시작
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

### 2.3. Kafka1, Zookeeper1 시작
1) **EC03** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2) start.sh 실행
```bash
$ ./start.sh
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

- /home/ubuntu/start.sh 없을 시
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f kafka1.yaml down
$ docker-compose -f kafka1.yaml up -d  # Kafka,Zookeeper 노드 시작
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

### 2.4. Kafka2, Zookeeper2 시작
1. **EC04** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. start.sh 실행
```bash
$ ./start.sh
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

- /home/ubuntu/start.sh 없을 시
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f kafka2.yaml down
$ docker-compose -f kafka2.yaml up -d  # Kafka,Zookeeper 노드 시작
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

### 2.5. Orderer0.exchange.com, Orderer1.exchange.com 시작
1. **EC05** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. start.sh 실행
```bash
$ ./start.sh
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

- /home/ubuntu/start.sh 없을 시
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f orderer0.yaml down
$ docker-compose -f orderer0.yaml up -d  # orderer0/orderer1 노드 시작
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

### 2.6. Orderer2.exchange.com, Orderer3.exchange.com 시작
1. **EC06** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. start.sh 실행
```bash
$ ./start.sh
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

- /home/ubuntu/start.sh 없을 시
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f orderer1.yaml down
$ docker-compose -f orderer1.yaml up -d  # orderer2/orderer3 노드 시작
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

### 2.7. CA.nuriorg.exchange.com, Peer0.nuriorg.exchange.com 시작
1. **EC07** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. start.sh 실행
```bash
$ ./start.sh
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

- /home/ubuntu/start.sh 없을 시
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f peer0.nuriorg.yaml down
$ docker-compose -f peer0.nuriorg.yaml up -d  # ca/peer0.nuriorg.exchange.com 노드 시작
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

### 2.8. Peer1.nuriorg.exchange.com 시작
1. **EC08** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. start.sh 실행
```bash
$ ./start.sh
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

- /home/ubuntu/start.sh 없을 시
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f peer1.nuriorg.yaml down
$ docker-compose -f peer1.nuriorg.yaml up -d  # peer1.nuriorg.exchange.com 노드 시작
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

### 2.9. CA.nflexorg.exchange.com, Peer0.nflexorg.exchange.com 시작
1. **EC09** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. start.sh 실행
```bash
$ ./start.sh
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

- /home/ubuntu/start.sh 없을 시
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f peer0.nflexorg.yaml down
$ docker-compose -f peer0.nflexorg.yaml up -d  # ca/peer0.nflexorg.exchange.com 노드 시작
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

### 2.10. Peer1.nflexorg.exchange.com 시작
1. **EC10** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. start.sh 실행
```bash
$ ./start.sh
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

- /home/ubuntu/start.sh 없을 시
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f peer1.nflexorg.yaml down
$ docker-compose -f peer1.nflexorg.yaml up -d  # peer1.nflexorg.exchange.com 노드 시작
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```


# 3. 블록체인 서버 중지
서버 중지 순서: 인스턴스 번호대로 중지하면 된다. (ex. EC1 중지 후 EC2 중지)

### 3.1. CA 노드(ca.exchange.com) 중지 {#ca-노드ca.exchange.com-중지 .21}
1. **EC01** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. docker-compose down
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f caserver.yaml down
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml down  # node-exporter 중지 
```

### 3.2. Kafka0, Zookeeper0 중지
1. **EC02** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. docker-compose down
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f kafka0.yaml down
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml down  # node-exporter 중지 
```

### 3.3. Kafka1, Zookeeper1 중지
1. **EC03** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. docker-compose down
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f kafka1.yaml down
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml down  # node-exporter 중지 
```

### 3.4. Kafka2, Zookeeper2 중지
1. **EC04** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. docker-compose down
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f kafka2.yaml down
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml down # node-exporter 중지
```

### 3.5. Orderer0.exchange.com, Orderer1.exchange.com 중지
1. **EC05** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. docker-compose down
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f orderer0.yaml down
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml down # node-exporter 중지
```

### 3.6. Orderer2.exchange.com, Orderer3.exchange.com 중지
1. **EC06** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. docker-compose down
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f orderer1.yaml down
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml down # node-exporter 중지
```

### 3.7. CA.nuriorg.exchange.com, Peer0.nuriorg.exchange.com 중지
1. **EC07** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. docker-compose down
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f peer0.nuriorg.yaml down
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml down # node-exporter 중지
```

### 3.8. Peer1.nuriorg.exchange.com 중지
1. **EC08** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. docker-compose down
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f peer1.nuriorg.yaml down
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml down # node-exporter 중지
```

### 3.9. CA.nflexorg.exchange.com, Peer0.nflexorg.exchange.com 중지
1. **EC09** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. docker-compose down
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f peer0.nflexorg.yaml down
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml down # node-exporter 중지
```

### 3.10. Peer1.nflexorg.exchange.com 중지
1. **EC10** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. docker-compose down
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f peer1.nflexorg.yaml down
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml down # node-exporter 중지
```


# 4. 블록체인 서버 초기화
서버 초기화 순서: 인스턴스 번호대로 초기화하면 된다.

### 4.1. CA 노드(ca.exchange.com) 초기화
1. **EC01** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. clean.sh 실행
```bash
$ ./clean.sh 
```

- /home/ubuntu/clean.sh 없을 시
```bash
$ docker rm $(docker ps -aq) -f; 
$ docker rmi $(docker images dev* -q) -f; docker volume rm $(docker volume ls -q); 
$ sudo rm -rf /home/ubuntu/hf/exchange-bc-production/volumes;
```

### 4.2. Kafka0, Zookeeper0 초기화
1. **EC02** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. clean.sh 실행
```bash
$ ./clean.sh 
```

- /home/ubuntu/clean.sh 없을 시
```bash
$ docker rm $(docker ps -aq) -f; 
$ docker rmi $(docker images dev* -q) -f; docker volume rm $(docker volume ls -q); 
$ sudo rm -rf /home/ubuntu/hf/exchange-bc-production/volumes;
```

### 4.3. Kafka1, Zookeeper1 초기화
1. **EC03** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. clean.sh 실행
```bash
$ ./clean.sh 
```

- /home/ubuntu/clean.sh 없을 시
```bash
$ docker rm $(docker ps -aq) -f; 
$ docker rmi $(docker images dev* -q) -f; docker volume rm $(docker volume ls -q); 
$ sudo rm -rf /home/ubuntu/hf/exchange-bc-production/volumes;
```

### 4.4. Kafka2, Zookeeper2 초기화
1. **EC04** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. clean.sh 실행
```bash
$ ./clean.sh 
```
- /home/ubuntu/clean.sh 없을 시
```bash
$ docker rm $(docker ps -aq) -f; 
$ docker rmi $(docker images dev* -q) -f; docker volume rm $(docker volume ls -q); 
$ sudo rm -rf /home/ubuntu/hf/exchange-bc-production/volumes;
```

### 4.5. Orderer0.exchange.com, Orderer1.exchange.com 초기화
1. **EC05** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. clean.sh 실행
```bash
$ ./clean.sh 
```

- /home/ubuntu/clean.sh 없을 시
```bash
$ docker rm $(docker ps -aq) -f; 
$ docker rmi $(docker images dev* -q) -f; docker volume rm $(docker volume ls -q); 
$ sudo rm -rf /home/ubuntu/hf/exchange-bc-production/volumes;
```

### 4.6. Orderer2.exchange.com, Orderer3.exchange.com 초기화
1. **EC06** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. clean.sh 실행
```bash
$ ./clean.sh 
```

- /home/ubuntu/clean.sh 없을 시
```bash
$ docker rm $(docker ps -aq) -f; 
$ docker rmi $(docker images dev* -q) -f; docker volume rm $(docker volume ls -q); 
$ sudo rm -rf /home/ubuntu/hf/exchange-bc-production/volumes;
```

### 4.7. CA.nuriorg.exchange.com, Peer0.nuriorg.exchange.com 초기화
1. **EC07** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. clean.sh 실행
```bash
$ ./clean.sh 
```
- /home/ubuntu/clean.sh 없을 시
```bash
$ docker rm $(docker ps -aq) -f; 
$ docker rmi $(docker images dev* -q) -f; docker volume rm $(docker volume ls -q); 
$ sudo rm -rf /home/ubuntu/hf/exchange-bc-production/volumes;
```

### 4.8. Peer1.nuriorg.exchange.com 초기화
1. **EC08** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. clean.sh 실행
```bash
$ ./clean.sh 
```
- /home/ubuntu/clean.sh 없을 시
```bash
$ docker rm $(docker ps -aq) -f; 
$ docker rmi $(docker images dev* -q) -f; docker volume rm $(docker volume ls -q); 
$ sudo rm -rf /home/ubuntu/hf/exchange-bc-production/volumes;
```

### 4.9. CA.nflexorg.exchange.com, Peer0.nflexorg.exchange.com 초기화
1. **EC09** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. clean.sh 실행
```bash
$ ./clean.sh 
```
- /home/ubuntu/clean.sh 없을 시
```bash
$ docker rm $(docker ps -aq) -f; 
$ docker rmi $(docker images dev* -q) -f; docker volume rm $(docker volume ls -q); 
$ sudo rm -rf /home/ubuntu/hf/exchange-bc-production/volumes;
```

### 4.10. Peer1.nflexorg.exchange.com 초기화
1. **EC10** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. clean.sh 실행
```bash
$ ./clean.sh 
```
- /home/ubuntu/clean.sh 없을 시
```bash
$ docker rm $(docker ps -aq) -f; 
$ docker rmi $(docker images dev* -q) -f; docker volume rm $(docker volume ls -q); 
$ sudo rm -rf /home/ubuntu/hf/exchange-bc-production/volumes;
```

### 4.11. CA 노드(ca.exchange.com) 재시작
1. **EC01** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. start.sh 실행
```bash
$ ./start.sh
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```
- /home/ubuntu/start.sh 없을 시
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f caserver.yaml up -d  # CA 노드 시작
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

### 4.12. Kafka0, Zookeeper0 재시작
1. **EC02** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. start.sh 실행
```bash
$ ./start.sh
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```
- /home/ubuntu/start.sh 없을 시
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f kafka0.yaml up -d  # Kafka,Zookeeper 노드 시작
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

### 4.13. Kafka1, Zookeeper1 재시작 {#kafka1-zookeeper1-재시작 .21}
1. **EC03** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. start.sh 실행
```bash
$ ./start.sh
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```
- /home/ubuntu/start.sh 없을 시
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f kafka1.yaml up -d  # Kafka,Zookeeper 노드 시작
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

### 4.14. Kafka2, Zookeeper2 재시작
1. **EC04** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. start.sh 실행
```bash
$ ./start.sh
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```
- /home/ubuntu/start.sh 없을 시
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f kafka2.yaml up -d  # Kafka,Zookeeper 노드 시작
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

### 4.15. Orderer0.exchange.com, Orderer1.exchange.com 재시작
1. **EC05** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. start.sh 실행
```bash
$ ./start.sh
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```
- /home/ubuntu/start.sh 없을 시
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f orderer0.yaml up -d  # orderer0/orderer1 노드 시작
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

### 4.16. Orderer2.exchange.com, Orderer3.exchange.com 재시작
1. **EC06** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. start.sh 실행
```bash
$ ./start.sh
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

- /home/ubuntu/start.sh 없을 시
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f orderer1.yaml up -d  # orderer2/orderer3 노드 시작
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

### 4.17. CA.nuriorg.exchange.com, Peer0.nuriorg.exchange.com 재시작
1. **EC07** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. start.sh 실행
```bash
$ ./start.sh
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```
- /home/ubuntu/start.sh 없을 시
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f peer0.nuriorg.yaml up -d  # ca/peer0.nuriorg.exchange.com 노드 시작
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

### 4.18. Peer1.nuriorg.exchange.com 재시작
1. **EC08** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. start.sh 실행
```bash
$ ./start.sh
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

- /home/ubuntu/start.sh 없을 시
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f peer1.nuriorg.yaml up -d  # peer1.nuriorg.exchange.com 노드 시작
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

### 4.19. CA.nflexorg.exchange.com, Peer0.nflexorg.exchange.com 재시작
1. **EC09** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. start.sh 실행
```bash
$ ./start.sh
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

- /home/ubuntu/start.sh 없을 시
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f peer0.nflexorg.yaml up -d  # ca/peer0.nflexorg.exchange.com 노드 시작
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

### 4.20. Peer1.nflexorg.exchange.com 재시작
1. **EC10** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. start.sh 실행
```bash
$ ./start.sh
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

- /home/ubuntu/start.sh 없을 시
```bash
$ cd ~/hf/exchange-bc-production/container_yaml/  # docker-compose 파일 폴더로 이동
$ docker-compose -f peer1.nflexorg.yaml up -d  # peer1.nflexorg.exchange.com 노드 시작
$ cd ~/node-exporter/
$ docker-compose -f docker-compose-nodeexporter.yml up -d;  # node-exporter 시작
```

### 4.21. 체인코드 설치
1. **EC10** SSH접속 및 홈 디렉토리로 이동(/home/ubuntu)
2. docker 명령어 실행
```bash
$ docker attach cli  # cli 컨테이너 진입 (엔터 두세번 누름)
# cli 컨테이너 진입 후
$ ./scripts/start.sh exchange-channel 10 60 # Chaincode 설치 시작
```
