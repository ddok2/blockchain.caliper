#!/usr/bin/env bash

CHANNEL_NAME=$1
DELAY=$2
TIMEOUT=$3
ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/exchange.com/orderers/orderer0.exchange.com/msp/tlscacerts/tlsca.exchange.com-cert.pem
LANGUAGE="golang"

CC_SRC_PATH=github.com/chaincode/exchange-ex

verifyResult () {
    if [[ $1 -ne 0 ]] ; then
        echo "!!!!!!!!!!!!!!! "$2" !!!!!!!!!!!!!!!!"
        echo "========= ERROR !!! FAILED to execute End-2-End Scenario ==========="
        echo
        exit 1
    fi
}

setGlobals () {
        PEER=$1
        ORG=$2
        if [[ ${ORG} -eq 1 ]] ; then
                CORE_PEER_LOCALMSPID="NuriOrgMSP"
                CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/nuriorg.exchange.com/peers/peer0.nuriorg.exchange.com/tls/ca.crt
                CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/nuriorg.exchange.com/users/Admin@nuriorg.exchange.com/msp
                if [[ ${PEER} -eq 0 ]]; then
                        CORE_PEER_ADDRESS=peer0.nuriorg.exchange.com:7051
                else
                        CORE_PEER_ADDRESS=peer1.nuriorg.exchange.com:8051
                fi
                echo "### Target Peer Address: ${CORE_PEER_ADDRESS} ###"
        elif [[ ${ORG} -eq 2 ]] ; then
                CORE_PEER_LOCALMSPID="NFlexOrgMSP"
                CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/nflexorg.exchange.com/peers/peer0.nflexorg.exchange.com/tls/ca.crt
                CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/nflexorg.exchange.com/users/Admin@nflexorg.exchange.com/msp
                if [[ ${PEER} -eq 0 ]]; then
                        CORE_PEER_ADDRESS=peer0.nflexorg.exchange.com:9051
                else
                        CORE_PEER_ADDRESS=peer1.nflexorg.exchange.com:10051
                fi
                echo "### Target Peer Address: ${CORE_PEER_ADDRESS} ###"
        elif [[ ${ORG} -eq 3 ]] ; then
                CORE_PEER_LOCALMSPID="Org3MSP"
                CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org3.exchange.com/peers/peer0.org3.exchange.com/tls/ca.crt
                CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org3.exchange.com/users/Admin@org3.exchange.com/msp
                if [[ ${PEER} -eq 0 ]]; then
                        CORE_PEER_ADDRESS=peer0.org3.exchange.com:7051
                else
                        CORE_PEER_ADDRESS=peer1.org3.exchange.com:7051
                fi
                echo "### Target Peer Address: ${CORE_PEER_ADDRESS} ###"
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

  if [[ ${res} -ne 0 && ${COUNTER} -lt ${MAX_RETRY} ]]; then
    COUNTER=$(expr ${COUNTER} + 1)
    echo "peer${PEER}.org${ORG} failed to join the channel, Retry after $DELAY seconds"
    sleep ${DELAY}
    joinChannelWithRetry ${PEER} ${ORG}
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

    if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
        set -x
        peer chaincode instantiate -o orderer0.exchange.com:7050 -C $CHANNEL_NAME -n exchange -l ${LANGUAGE} -v ${VERSION} -c '{"Args":[]}' -P "OR ('NuriOrgMSP.peer','NFlexOrgMSP.peer')" &> log.txt
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

echo "Creating channel..."
createChannel
echo "Having all peers join the channel..."
joinChannel
echo "Updating anchor peers for org1..."
updateAnchorPeers 0 1
echo "Updating anchor peers for org2..."
updateAnchorPeers 0 2

echo "Installing chaincode on peer0.org1..."
installChaincode 0 1
echo "Install chaincode on peer0.org2..."
installChaincode 0 2
echo "Install chaincode on peer1.org1..."
installChaincode 1 1
echo "Install chaincode on peer1.org2..."
installChaincode 1 2
echo "Instantiating chaincode on peer0.org1..."
instantiateChaincode 0 1

echo
echo "========= All GOOD, NURI Blockchain execution completed =========== "
echo

echo
echo " _____   _   _   ____   "
echo "| ____| | \ | | |  _ \  "
echo "|  _|   |  \| | | | | | "
echo "| |___  | |\  | | |_| | "
echo "|_____| |_| \_| |____/  "
echo

exit 0
