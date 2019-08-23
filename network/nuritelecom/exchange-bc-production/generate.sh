#!/bin/sh
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#
export PATH=$GOPATH/src/github.com/hyperledger/fabric/build/bin:${PWD}/bin:${PWD}:$PATH
mkdir channel-artifacts
export FABRIC_CFG_PATH=${PWD}
CHANNEL_NAME=exchange-channel

# remove previous crypto material and config transactions
rm -fr config/*
rm -fr crypto-config/*

# generate crypto material
cryptogen generate --config=./crypto-config.yaml
if [ "$?" -ne 0 ]; then
  echo "Failed to generate crypto material..."
  exit 1
fi

# generate genesis block for orderer
configtxgen -profile TwoOrgsExchangeOrdererGenesis -outputBlock ./channel-artifacts/genesis.block
if [ "$?" -ne 0 ]; then
  echo "Failed to generate orderer genesis block..."
  exit 1
fi

# generate channel configuration transaction
configtxgen -profile TwoOrgsExchangeChannel -outputCreateChannelTx ./channel-artifacts/channel.tx -channelID $CHANNEL_NAME
if [ "$?" -ne 0 ]; then
  echo "Failed to generate channel configuration transaction..."
  exit 1
fi

# generate anchor peer transaction
configtxgen -profile TwoOrgsExchangeChannel -outputAnchorPeersUpdate ./channel-artifacts/NuriOrgMSPanchors.tx -channelID $CHANNEL_NAME -asOrg NuriOrgMSP
if [ "$?" -ne 0 ]; then
  echo "Failed to generate anchor peer update for NuriOrgMSP..."
  exit 1
fi

configtxgen -profile TwoOrgsExchangeChannel -outputAnchorPeersUpdate ./channel-artifacts/NFlexOrgMSPanchors.tx -channelID $CHANNEL_NAME -asOrg NFlexOrgMSP
if [ "$?" -ne 0 ]; then
  echo "Failed to generate anchor peer update for NFlexOrgMSP..."
  exit 1
fi

echo "\n\nScripts successfully executed..."