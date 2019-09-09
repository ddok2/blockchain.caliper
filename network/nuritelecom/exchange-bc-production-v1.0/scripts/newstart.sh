#set -ev

CHANNEL_NAME=$1
DELAY=$2
TIMEOUT=$3
ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/exchange.com/orderers/orderer0.exchange.com/msp/tlscacerts/tlsca.exchange.com-cert.pem
LANGUAGE="golang"

CC_SRC_PATH=github.com/chaincode/exchange-ex

verifyResult () {
        if [ $1 -ne 0 ] ; then
                echo "!!!!!!!!!!!!!!! "$2" !!!!!!!!!!!!!!!!"
    echo "========= ERROR !!! FAILED to execute End-2-End Scenario ==========="
                echo
                exit 1
        fi
}

# Set OrdererOrg.Admin globals
setOrdererGlobals() {
        CORE_PEER_LOCALMSPID="OrdererMSP"
        CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/exchange.com/orderers/orderer.exchange.com/msp/tlscacerts/tlsca.exchange.com-cert.pem
        CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/exchange.com/users/Admin@exchange.com/msp
}

setGlobals () {
        PEER=$1
        ORG=$2
        if [ $ORG -eq 1 ] ; then
                CORE_PEER_LOCALMSPID="NuriOrgMSP"
                CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/nuriorg.exchange.com/peers/peer0.nuriorg.exchange.com/tls/ca.crt
                CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/nuriorg.exchange.com/users/Admin@nuriorg.exchange.com/msp
                if [ $PEER -eq 0 ]; then
                        CORE_PEER_ADDRESS=peer0.nuriorg.exchange.com:7051
                else
                        CORE_PEER_ADDRESS=peer1.nuriorg.exchange.com:7051
                fi
        elif [ $ORG -eq 2 ] ; then
                CORE_PEER_LOCALMSPID="NFlexOrgMSP"
                CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/nflexorg.exchange.com/peers/peer0.nflexorg.exchange.com/tls/ca.crt
                CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/nflexorg.exchange.com/users/Admin@nflexorg.exchange.com/msp
                if [ $PEER -eq 0 ]; then
                        CORE_PEER_ADDRESS=peer0.nflexorg.exchange.com:7051
                else
                        CORE_PEER_ADDRESS=peer1.nflexorg.exchange.com:7051
                fi

        elif [ $ORG -eq 3 ] ; then
                CORE_PEER_LOCALMSPID="Org3MSP"
                CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org3.exchange.com/peers/peer0.org3.exchange.com/tls/ca.crt
                CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org3.exchange.com/users/Admin@org3.exchange.com/msp
                if [ $PEER -eq 0 ]; then
                        CORE_PEER_ADDRESS=peer0.org3.exchange.com:7051
                else
                        CORE_PEER_ADDRESS=peer1.org3.exchange.com:7051
                fi
        else
                echo "================== ERROR !!! ORG Unknown =================="
        fi

        env |grep CORE
}

joinChannelWithRetry () {
  PEER=$1
  ORG=$2
  setGlobals $PEER $ORG

        set -x
  peer channel join -b $CHANNEL_NAME.block  &> log.txt
  res=$?
        set +x
  cat log.txt
  echo "Debug String Begin"
  echo $res
  echo $COUNTER
  echo $MAX_RETRY
  echo "Debug String End"

  if [ $res -ne 0 -a $COUNTER -lt $MAX_RETRY ]; then
    COUNTER=` expr $COUNTER + 1`
    echo "peer${PEER}.org${ORG} failed to join the channel, Retry after $DELAY seconds"
    sleep $DELAY
    joinChannelWithRetry $PEER $ORG
  else
    COUNTER=1
  fi
  verifyResult $res "After $MAX_RETRY attempts, peer${PEER}.org${ORG} has failed to Join the Channel"
}


createChannel() {
        setGlobals 0 1

        if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
                set -x
                peer channel create -o orderer0.exchange.com:7050 -c $CHANNEL_NAME -f ./channel-artifacts/channel.tx &> log.txt
                res=$?
                set +x
        else
                echo "dedicated-tls"
                echo $CHANNEL_NAME
                echo $ORDERER_CA
                set -x
                peer channel create -o orderer0.exchange.com:7050 -c $CHANNEL_NAME -f ./channel-artifacts/channel.tx --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA &> log.txt
                res=$?
                set +x
        fi
        cat log.txt
        verifyResult $res "Channel creation failed"
        echo "===================== Channel \"$CHANNEL_NAME\" is created successfully ===================== "
}

joinChannel () {
	sleep $DELAY
        for org in 1 2; do
            for peer in 0 1; do
                joinChannelWithRetry $peer $org
                echo "===================== peer${peer}.org${org} joined on the channel \"$CHANNEL_NAME\" ===================== "
                sleep $DELAY
                echo
            done
        done
}

updateAnchorPeers() {
  PEER=$1
  ORG=$2
  setGlobals $PEER $ORG

  if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
                set -x
                peer channel update -o orderer0.exchange.com:7050 -c $CHANNEL_NAME -f ./channel-artifacts/${CORE_PEER_LOCALMSPID}anchors.tx &> log.txt
                res=$?
                set +x
  else
                set -x
                peer channel update -o orderer0.exchange.com:7050 -c $CHANNEL_NAME -f ./channel-artifacts/${CORE_PEER_LOCALMSPID}anchors.tx --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA &>log.txt
                res=$?
                set +x
  fi
        cat log.txt
        verifyResult $res "Anchor peer update failed"
        echo "===================== Anchor peers for org \"$CORE_PEER_LOCALMSPID\" on \"$CHANNEL_NAME\" is updated successfully ===================== "
        sleep $DELAY
        echo
}

installChaincode () {
        PEER=$1
        ORG=$2
        setGlobals $PEER $ORG
        VERSION=${3:-1.0}
        set -x
        peer chaincode install -n exchange -v ${VERSION} -l ${LANGUAGE} -p ${CC_SRC_PATH} &> log.txt
        res=$?
        set +x
        cat log.txt
        verifyResult $res "Chaincode installation on peer${PEER}.org${ORG} has Failed"
        echo "===================== Chaincode is installed on peer${PEER}.org${ORG} ===================== "
	sleep $DELAY
        echo
}

instantiateChaincode () {
        PEER=$1
        ORG=$2
        setGlobals $PEER $ORG
        VERSION=${3:-1.0}

        # while 'peer chaincode' command can get the orderer endpoint from the peer (if join was successful),
        # lets supply it directly as we know it using the "-o" option
        if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
                set -x
                peer chaincode instantiate -o orderer0.exchange.com:7050 -C $CHANNEL_NAME -n exchange -l ${LANGUAGE} -v ${VERSION} -c '{"Args":[]}' -P "OR  ('NuriOrgMSP.peer','NFlexOrgMSP.peer')" &> log.txt
                res=$?
                set +x
        else
                set -x
                peer chaincode instantiate -o orderer0.exchange.com:7050 --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA -C $CHANNEL_NAME -n exchange -l ${LANGUAGE} -v 1.0 -c '{"Args":[]}' -P "OR ('NuriOrgMSP.peer','NFlexOrgMSP.peer')" &> log.txt
                res=$?
                set +x
        fi
        cat log.txt
        verifyResult $res "Chaincode instantiation on peer${PEER}.org${ORG} on channel '$CHANNEL_NAME' failed"
        echo "===================== Chaincode Instantiation on peer${PEER}.org${ORG} on channel '$CHANNEL_NAME' is successful ===================== "
	sleep $DELAY
        echo
}

createChannel

joinChannel

updateAnchorPeers 0 1

updateAnchorPeers 0 2

installChaincode 0 1

installChaincode 1 1

installChaincode 0 2

installChaincode 1 2

instantiateChaincode 0 1

