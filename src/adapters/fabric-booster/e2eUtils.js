/*
 * Copyright 2019. Nuri Telecom. All Rights Reserved.
 *
 * - e2eUtils.js
 * - author: Sungyub NA <mailto: syna@nuritelecom.com>
 */

'use strict';

const commUtils = require('../../comm/util');
// const commLogger = commUtils.getLogger('e2eUtils.js');
const TxStatus = require('../../comm/transaction');

const FabricCAServices = require('fabric-ca-client');
const Client = require('fabric-client');
const fs = require('fs');
const util = require('util');
const testUtil = require('./util.js');

const client = require('http');
const qs = require('qs');

const commLogger = {
  debug: (msg) => {
    process.send({
      type: 'socket.io',
      data: {
        message: msg
      }
    });
    return commUtils.getLogger('e2eUtils.js').debug(msg);
  },
  info: (msg) => {
    process.send({
      type: 'socket.io',
      data: {
        message: msg
      }
    });
    return commUtils.getLogger('e2eUtils.js').info(msg);
  },
  warn: (msg) => {
    process.send({
      type: 'socket.io',
      data: {
        message: msg
      }
    });
    return commUtils.getLogger('e2eUtils.js').warn(msg);
  },
  error: (msg) => {
    process.send({
      type: 'socket.io',
      data: {
        message: msg
      }
    });
    return commUtils.getLogger('e2eUtils.js').error(msg);
  }
};

// const signedOffline = require('./signTransactionOffline.js');

let Gateway, InMemoryWallet, X509WalletMixin;
let ORGS;
let isLegacy;
let tx_id = null;
let the_user = null;
let boosterConfig = null;

let signedTransactionArray = [];
let signedCommitProposal = [];
let txFile;
let invokeCount = 0;
let clientIndex = 0;

/**
 * Initialize the Fabric client configuration.
 * @param {string} config_path The path of the Fabric network configuration file.
 */
function init(config_path) {
  const config = commUtils.parseYaml(config_path);
  ORGS = config.fabric.network;
  boosterConfig = config.booster || {};
}

/**
 * Enrol and get the cert
 * @param {*} fabricCAEndpoint url of org endpoint
 * @param {*} caName name of caName
 * @return {Object} something useful in a promise
 */
async function tlsEnroll(fabricCAEndpoint, caName) {
  const tlsOptions = {
    trustedRoots: [],
    verify: false,
  };
  const caService = new FabricCAServices(fabricCAEndpoint, tlsOptions,
      caName);
  const req = {
    enrollmentID: 'admin',
    enrollmentSecret: 'adminpw',
    profile: 'tls',
  };

  const enrollment = await caService.enroll(req);
  enrollment.key = enrollment.key.toBytes();
  return enrollment;
}

/**
 * Read signed proposal from file.
 * @param {string} name The prefix name of the file.
 * @async
 */
async function readFromFile(name) {
  try {
    signedTransactionArray = [];
    signedCommitProposal = [];
    invokeCount = 0;
    let fileName = name + '.signed.metadata.' + clientIndex;
    let binFileName = name + '.signed.binary.' + clientIndex;

    let data = fs.readFileSync(fileName);
    signedTransactionArray = JSON.parse(data);
    commLogger.debug('read buffer file ok');
    let signedBuffer = fs.readFileSync(binFileName);
    let start = 0;
    for (let i = 0; i < signedTransactionArray.length; i++) {
      let length = signedTransactionArray[i].signatureLength;
      let signature = signedBuffer.slice(start, start + length);
      start += length;
      length = signedTransactionArray[i].payloadLength;
      let payload = signedBuffer.slice(start, start + length);
      signedCommitProposal.push(
          { signature: signature, payload: payload });
      start += length;
    }
  } catch (err) {
    commLogger.error('read err: ' + err);
  }
}

module.exports.readFromFile = readFromFile;

/**
 * Deploy the given chaincode to the given organization's peers.
 * @param {string} org The name of the organization.
 * @param {object} chaincode The chaincode object from the configuration file.
 * @async
 */
async function installChaincode(org, chaincode) {
  Client.setConfigSetting('request-timeout', 60000);
  const channel_name = chaincode.channel;

  const client = new Client();
  const channel = client.newChannel(channel_name);

  // Conditional action on TLS enablement
  if (ORGS.orderer.url.toString().startsWith('grpcs')) {
    const fabricCAEndpoint = ORGS[org].ca.url;
    const caName = ORGS[org].ca.name;
    const tlsInfo = await tlsEnroll(fabricCAEndpoint, caName);
    client.setTlsClientCertAndKey(tlsInfo.certificate, tlsInfo.key);
  }

  const orgName = ORGS[org].name;
  const cryptoSuite = Client.newCryptoSuite();
  cryptoSuite.setCryptoKeyStore(
      Client.newCryptoKeyStore({ path: testUtil.storePathForOrg(orgName) }));
  client.setCryptoSuite(cryptoSuite);

  const caRootsPath = ORGS.orderer.tls_cacerts;
  let data = fs.readFileSync(commUtils.resolvePath(caRootsPath));
  let caroots = Buffer.from(data).toString();

  channel.addOrderer(
      client.newOrderer(
          ORGS.orderer.url,
          {
            'pem': caroots,
            'ssl-target-name-override': ORGS.orderer['server-hostname'],
          },
      ),
  );

  const targets = [];
  for (let key in ORGS[org]) {
    if (ORGS[org].hasOwnProperty(key)) {
      if (key.indexOf('peer') === 0) {
        let data = fs.readFileSync(
            commUtils.resolvePath(ORGS[org][key].tls_cacerts));
        let peer = client.newPeer(
            ORGS[org][key].requests,
            {
              pem: Buffer.from(data).toString(),
              'ssl-target-name-override': ORGS[org][key]['server-hostname'],
            },
        );

        targets.push(peer);
        channel.addPeer(peer);
      }
    }
  }

  const store = await Client.newDefaultKeyValueStore(
      { path: testUtil.storePathForOrg(orgName) });
  client.setStateStore(store);

  // get the peer org's admin required to send install chaincode requests
  the_user = await testUtil.getSubmitter(client,
      true /* get peer org admin */, org);

  // Don't re-install existing chaincode
  let peers = channel.getPeers();
  let res = await client.queryInstalledChaincodes(
      peers[0].constructor.name.localeCompare('Peer') === 0
          ? peers[0]
          : peers[0]._peer);
  let found = false;
  for (let i = 0; i < res.chaincodes.length; i++) {
    if (res.chaincodes[i].name === chaincode.id &&
        res.chaincodes[i].version === chaincode.version &&
        res.chaincodes[i].path === chaincode.path) {
      found = true;
      commLogger.debug(
          'installedChaincode: ' + JSON.stringify(res.chaincodes[i]));
      break;
    }
  }
  if (found) {
    return;
  }

  let resolvedPath = chaincode.path;
  let metadataPath = chaincode.metadataPath
      ? commUtils.resolvePath(chaincode.metadataPath)
      : chaincode.metadataPath;
  if (chaincode.language === 'node') {
    resolvedPath = commUtils.resolvePath(chaincode.path);
  }

  // send proposal to endorser
  const request = {
    targets: targets,
    chaincodePath: resolvedPath,
    metadataPath: metadataPath,
    chaincodeId: chaincode.id,
    chaincodeType: chaincode.language,
    chaincodeVersion: chaincode.version,
  };

  const results = await client.installChaincode(request);

  const proposalResponses = results[0];

  let all_good = true;
  const errors = [];
  for (let i in proposalResponses) {
    let one_good = false;
    if (proposalResponses && proposalResponses[i].response &&
        proposalResponses[i].response.status === 200) {
      one_good = true;
    } else {
      commLogger.error('install proposal was bad');
      errors.push(proposalResponses[i]);
    }
    all_good = all_good && one_good;
  }
  if (!all_good) {
    throw new Error(util.format(
        'Failed to send install Proposal or receive valid response: %s',
        errors));
  }
}

/**
 * Assemble a chaincode proposal request.
 * @param {Client} client The Fabric client object.
 * @param {object} chaincode The chaincode object from the configuration file.
 * @param {boolean} upgrade Indicates whether the request is an upgrade or not.
 * @param {object} transientMap The transient map the request.
 * @param {object} endorsement_policy The endorsement policy object from the configuration file.
 * @return {object} The assembled chaincode proposal request.
 */
function buildChaincodeProposal(
    client, chaincode, upgrade, transientMap, endorsement_policy) {
  const tx_id = client.newTransactionID();

  // send proposal to endorser
  const request = {
    chaincodePath: chaincode.path,
    chaincodeId: chaincode.id,
    chaincodeType: chaincode.language,
    chaincodeVersion: chaincode.version,
    fcn: 'init',
    args: chaincode.init || [],
    txId: tx_id,
    'endorsement-policy': endorsement_policy,
  };

  if (upgrade) {
    // use this call to test the transient map support during chaincode instantiation
    request.transientMap = transientMap;
  }

  return request;
}

/**
 * Instantiate or upgrade the given chaincode with the given endorsement policy.
 * @param {object} chaincode The chaincode object from the configuration file.
 * @param {object} endorsement_policy The endorsement policy object from the configuration file.
 * @param {boolean} upgrade Indicates whether the call is an upgrade or a new instantiation.
 * @async
 */
async function instantiate(chaincode, endorsement_policy, upgrade) {
  Client.setConfigSetting('request-timeout', 600000);

  let channel = testUtil.getChannel(chaincode.channel);
  if (channel === null) {
    throw new Error('Could not find channel in config');
  }
  const channel_name = channel.name;
  const userOrg = channel.organizations[0];

  const targets = [];
  const eventhubs = [];
  let type = 'instantiate';
  if (upgrade) {
    type = 'upgrade';
  }
  const client = new Client();
  channel = client.newChannel(channel_name);

  const orgName = ORGS[userOrg].name;
  const cryptoSuite = Client.newCryptoSuite();
  cryptoSuite.setCryptoKeyStore(
      Client.newCryptoKeyStore({ path: testUtil.storePathForOrg(orgName) }));
  client.setCryptoSuite(cryptoSuite);

  const caRootsPath = ORGS.orderer.tls_cacerts;
  let data = fs.readFileSync(commUtils.resolvePath(caRootsPath));
  let caroots = Buffer.from(data).toString();

  // Conditional action on TLS enablement
  if (ORGS.orderer.url.toString().startsWith('grpcs')) {
    const fabricCAEndpoint = ORGS[userOrg].ca.url;
    const caName = ORGS[userOrg].ca.name;
    const tlsInfo = await tlsEnroll(fabricCAEndpoint, caName);
    client.setTlsClientCertAndKey(tlsInfo.certificate, tlsInfo.key);
  }

  channel.addOrderer(
      client.newOrderer(
          ORGS.orderer.url,
          {
            'pem': caroots,
            'ssl-target-name-override': ORGS.orderer['server-hostname'],
          },
      ),
  );

  const transientMap = { 'test': 'transientValue' };
  let request = null;

  const store = await Client.newDefaultKeyValueStore(
      { path: testUtil.storePathForOrg(orgName) });
  client.setStateStore(store);
  the_user = await testUtil.getSubmitter(client, true /* use peer org admin*/,
      userOrg);

  for (let org in ORGS) {
    if (ORGS.hasOwnProperty(org) && org.indexOf('org') === 0) {
      for (let key in ORGS[org]) {
        if (ORGS[org].hasOwnProperty(key) && key.indexOf('peer') ===
            0) {
          let data = fs.readFileSync(
              commUtils.resolvePath(ORGS[org][key].tls_cacerts));
          let peer = client.newPeer(
              ORGS[org][key].requests,
              {
                pem: Buffer.from(data).toString(),
                'ssl-target-name-override': ORGS[org][key]['server-hostname'],
              });
          targets.push(peer);
          channel.addPeer(peer);

          const eh = channel.newChannelEventHub(peer);
          eventhubs.push(eh);
        }
      }
    }
  }

  await channel.initialize();

  let res = await channel.queryInstantiatedChaincodes();
  let found = false;
  for (let i = 0; i < res.chaincodes.length; i++) {
    if (res.chaincodes[i].name === chaincode.id &&
        res.chaincodes[i].version === chaincode.version &&
        res.chaincodes[i].path === chaincode.path) {
      found = true;
      commLogger.debug(
          'instantiatedChaincode: ' + JSON.stringify(res.chaincodes[i]));
      break;
    }
  }
  if (found) {
    return;
  }

  let results;
  // the v1 chaincode has Init() method that expects a transient map
  if (upgrade) {
    let request = buildChaincodeProposal(client, chaincode, upgrade,
        transientMap, endorsement_policy);
    tx_id = request.txId;
    results = await channel.sendUpgradeProposal(request);
  } else {
    let request = buildChaincodeProposal(client, chaincode, upgrade,
        transientMap, endorsement_policy);
    tx_id = request.txId;
    results = await channel.sendInstantiateProposal(request);
  }

  const proposalResponses = results[0];

  const proposal = results[1];
  let all_good = true;
  for (const i in proposalResponses) {
    if (proposalResponses && proposalResponses[i].response &&
        proposalResponses[i].response.status === 200) {
      commLogger.info(type + ' proposal was good');
    } else {
      commLogger.warn(
          type + ' proposal was bad: ' + proposalResponses[i]);
      all_good = false;
    }
  }
  if (all_good) {
    commLogger.info(
        'Successfully sent Proposal and received ProposalResponse');
    request = {
      proposalResponses: proposalResponses,
      proposal: proposal,
    };
  } else {
    commLogger.warn(JSON.stringify(proposalResponses));
    throw new Error('All proposals were not good');
  }
  const deployId = tx_id.getTransactionID();

  const eventPromises = [];
  eventPromises.push(channel.sendTransaction(request));
  eventhubs.forEach((eh) => {
    let txPromise = new Promise((resolve, reject) => {
      let handle = setTimeout(reject, 300000);

      eh.registerTxEvent(deployId.toString(), (tx, code) => {
        commLogger.info('The chaincode ' + type +
            ' transaction has been committed on peer ' +
            eh.getPeerAddr());
        clearTimeout(handle);
        if (code !== 'VALID') {
          commLogger.warn('The chaincode ' + type +
              ' transaction was invalid, code = ' + code);
          reject();
        } else {
          commLogger.info(
              'The chaincode ' + type + ' transaction was valid.');
          resolve();
        }
      }, (err) => {
        commLogger.warn(
            'There was a problem with the instantiate event ' + err);
        clearTimeout(handle);
        reject();
      }, {
        disconnect: true,
      });
      eh.connect();
    });
    eventPromises.push(txPromise);
  });

  results = await Promise.all(eventPromises);
  if (results && !(results[0] instanceof Error) && results[0].status ===
      'SUCCESS') {
    commLogger.info(
        'Successfully sent ' + type + 'transaction to the orderer.');
  } else {
    commLogger.warn(
        'Failed to order the ' + type + 'transaction. Error code: ' +
        results[0].status);
    throw new Error(
        'Failed to order the ' + type + 'transaction. Error code: ' +
        results[0].status);
  }
}

/**
 * Instantiate or upgrade the given chaincode with the given endorsement policy.
 * @param {object} chaincode The chaincode object from the configuration file.
 * @param {object} endorsement_policy The endorsement policy object from the configuration file.
 * @param {boolean} upgrade Indicates whether the call is an upgrade or a new instantiation.
 * @async
 */
async function instantiateChaincode(chaincode, endorsement_policy, upgrade) {

  // if (isLegacy) {
  //     await instantiateLegacy(chaincode, endorsement_policy, upgrade);
  // } else {
  await instantiate(chaincode, endorsement_policy, upgrade);
  // }
}

/**
 * Get the peers of a given organization.
 * @param {string} orgName The name of the organization.
 * @return {string[]} The collection of peer names.
 */
function getOrgPeers(orgName) {
  const peers = [];
  const org = ORGS[orgName];
  for (let key in org) {
    if (org.hasOwnProperty(key)) {
      if (key.indexOf('peer') === 0) {
        peers.push(org[key]);
      }
    }
  }

  return peers;
}

/**
 * Create a Fabric context based on the channel configuration.
 * @param {object} channelConfig The channel object from the configuration file.
 * @param {Integer} clientIdx the client index
 * @param {object} txModeFile The file information for reading or writing.
 * @return {Promise<object>} The created Fabric context.
 */
async function getcontext(channelConfig, clientIdx, txModeFile) {
  clientIndex = clientIdx;
  txFile = txModeFile;
  Client.setConfigSetting('request-timeout', 120000);
  const channel_name = channelConfig.name;
  // var userOrg = channelConfig.organizations[0];
  // choose a random org to use, for load balancing
  const idx = Math.floor(Math.random() * channelConfig.organizations.length);
  const userOrg = channelConfig.organizations[idx];

  const client = new Client();
  const channel = client.newChannel(channel_name);
  let orgName = ORGS[userOrg].name;
  const cryptoSuite = Client.newCryptoSuite();
  const eventhubs = [];

  // Conditional action on TLS enablement
  if (ORGS[userOrg].ca.url.toString().startsWith('https')) {
    const fabricCAEndpoint = ORGS[userOrg].ca.url;
    const caName = ORGS[userOrg].ca.name;
    const tlsInfo = await tlsEnroll(fabricCAEndpoint, caName);
    client.setTlsClientCertAndKey(tlsInfo.certificate, tlsInfo.key);
  }

  cryptoSuite.setCryptoKeyStore(
      Client.newCryptoKeyStore({ path: testUtil.storePathForOrg(orgName) }));
  client.setCryptoSuite(cryptoSuite);

  const caRootsPath = ORGS.orderer.tls_cacerts;
  let data = fs.readFileSync(commUtils.resolvePath(caRootsPath));
  let caroots = Buffer.from(data).toString();

  channel.addOrderer(
      client.newOrderer(
          ORGS.orderer.url,
          {
            'pem': caroots,
            'ssl-target-name-override': ORGS.orderer['server-hostname'],
          },
      ),
  );

  orgName = ORGS[userOrg].name;

  const store = await Client.newDefaultKeyValueStore(
      { path: testUtil.storePathForOrg(orgName) });
  client.setStateStore(store);
  the_user = await testUtil.getSubmitter(client, true, userOrg);

  // set up the channel to use each org's random peer for
  // both requests and events
  for (let i in channelConfig.organizations) {
    let org = channelConfig.organizations[i];
    let peers = getOrgPeers(org);

    if (peers.length === 0) {
      throw new Error('could not find peer of ' + org);
    }

    // Cycle through available peers based on clientIdx
    let peerInfo = peers[clientIdx % peers.length];
    let data = fs.readFileSync(commUtils.resolvePath(peerInfo.tls_cacerts));
    let peer = client.newPeer(
        peerInfo.requests,
        {
          pem: Buffer.from(data).toString(),
          'ssl-target-name-override': peerInfo['server-hostname'],
        },
    );
    channel.addPeer(peer);

    // an event listener can only register with the peer in its own org
    if (isLegacy) {
      let eh = client.newEventHub();
      eh.setPeerAddr(
          peerInfo.events,
          {
            pem: Buffer.from(data).toString(),
            'ssl-target-name-override': peerInfo['server-hostname'],
            //'request-timeout': 120000
            'grpc.keepalive_timeout_ms': 3000, // time to respond to the ping, 3 seconds
            'grpc.keepalive_time_ms': 360000,   // time to wait for ping response, 6 minutes
            // 'grpc.http2.keepalive_time' : 15
          },
      );
      eventhubs.push(eh);
    } else {
      if (org === userOrg) {
        let eh = channel.newChannelEventHub(peer);
        eventhubs.push(eh);
      }
    }
  }

  // register event listener
  eventhubs.forEach((eh) => {
    eh.connect();
  });

  await channel.initialize();
  return {
    org: userOrg,
    client: client,
    channel: channel,
    submitter: the_user,
    eventhubs: eventhubs,
  };
}

/**
 * Disconnect the event hubs.
 * @param {object} context The Fabric context.
 * @async
 */
async function releasecontext(context) {
  if (context.hasOwnProperty('eventhubs')) {
    for (let key in context.eventhubs) {
      const eventhub = context.eventhubs[key];
      if (eventhub && eventhub.isconnected()) {
        eventhub.disconnect();
      }
    }
    context.eventhubs = [];
  }
}

/**
 * Write signed proposal to file.
 * @param {string} name The prefix name of the file.
 * @async
 */
async function writeToFile(name) {
  let fileName = name + '.signed.metadata.' + clientIndex;
  let binFileName = name + '.signed.binary.' + clientIndex;

  try {
    let reArray = [];
    let bufferArray = [];
    for (let i = 0; i < signedTransactionArray.length; i++) {
      let signedTransaction = signedTransactionArray[i];
      let signedProposal = signedTransactionArray[i].signedTransaction;
      let signature = signedProposal.signature;
      bufferArray.push(signature);
      let payload = signedProposal.payload;
      bufferArray.push(payload);
      reArray.push({
        txId: signedTransaction.txId,
        transactionRequest: signedTransaction.transactionRequest,
        signatureLength: signature.length,
        payloadLength: payload.length,
      });
    }
    let buffer = Buffer.concat(bufferArray);

    fs.writeFileSync(binFileName, buffer);
    let signedString = JSON.stringify(reArray);
    fs.writeFileSync(fileName, signedString);
    signedTransactionArray = [];
    signedCommitProposal = [];
    commLogger.debug('write file ok');

  } catch (err) {
    commLogger.error('write err: ' + err);
  }

}

module.exports.writeToFile = writeToFile;

const TxErrorEnum = require('./constant.js').TxErrorEnum;
const TxErrorIndex = require('./constant.js').TxErrorIndex;

// /**
//  * Submit a transaction to the orderer.
//  * @param {object} context The Fabric context.
//  * @param {object} signedTransaction The transaction information.
//  * @param {object} invokeStatus The result and stats of the transaction.
//  * @param {number} startTime The start time.
//  * @param {number} timeout The timeout for the transaction invocation.
//  * @return {Promise<TxStatus>} The result and stats of the transaction invocation.
//  */
// // eslint-disable-next-line no-unused-vars,require-jsdoc
// async function sendTransaction(
//     context, signedTransaction, invokeStatus, startTime, timeout) {
//
//     const channel = context.channel;
//     const eventHubs = context.eventhubs;
//     const txId = signedTransaction.txId;
//     let errFlag = TxErrorEnum.NoError;
//     try {
//         let newTimeout = timeout * 1000 - (Date.now() - startTime);
//         if (newTimeout < 10000) {
//             commLogger.warn(
//                 'WARNING: timeout is too small, default value is used instead');
//             newTimeout = 10000;
//         }
//
//         // todo change to rest api
//         const eventPromises = [];
//         eventHubs.forEach((eh) => {
//             eventPromises.push(new Promise((resolve, reject) => {
//                 //let handle = setTimeout(() => reject(new Error('Timeout')), newTimeout);
//                 let handle = setTimeout(() => reject(new Error('Timeout')),
//                     100000);
//                 eh.registerTxEvent(txId,
//                     (tx, code) => {
//                         clearTimeout(handle);
//                         eh.unregisterTxEvent(txId);
//
//                         // either explicit invalid event or valid event, verified in both cases by at least one peer
//                         invokeStatus.SetVerification(true);
//                         if (code !== 'VALID') {
//                             let err = new Error('Invalid transaction: ' + code);
//                             errFlag |= TxErrorEnum.BadEventNotificationError;
//                             invokeStatus.SetFlag(errFlag);
//                             invokeStatus.SetErrMsg(
//                                 TxErrorIndex.BadEventNotificationError,
//                                 err.toString());
//                             reject(err); // handle error in final catch
//                         } else {
//                             resolve();
//                         }
//                     },
//                     (err) => {
//                         clearTimeout(handle);
//                         // we don't know what happened, but give the other eventhub connections a chance
//                         // to verify the Tx status, so resolve this call
//                         errFlag |= TxErrorEnum.EventNotificationError;
//                         invokeStatus.SetFlag(errFlag);
//                         invokeStatus.SetErrMsg(
//                             TxErrorIndex.EventNotificationError,
//                             err.toString());
//                         resolve();
//                     },
//                 );
//
//             }));
//         });
//
//         let broadcastResponse;
//         try {
//             let signedProposal = signedTransaction.signedTransaction;
//             let broadcastResponsePromise;
//             let transactionRequest = signedTransaction.transactionRequest;
//             if (signedProposal === null) {
//                 // if(txFile && txFile.readWrite === 'write') {
//                 //     const beforeInvokeTime = Date.now();
//                 //     let signedTransaction = signedOffline.generateSignedTransaction(transactionRequest, channel);
//                 //     invokeStatus.Set('invokeLatency', (Date.now() - beforeInvokeTime));
//                 //     signedTransactionArray.push({
//                 //         txId: txId,
//                 //         signedTransaction: signedTransaction,
//                 //         transactionRequest: {orderer: transactionRequest.orderer}
//                 //     });
//                 //     return invokeStatus;
//                 // }
//                 const beforeTransactionTime = Date.now();
//                 broadcastResponsePromise = channel.sendTransaction(
//                     transactionRequest);
//                 invokeStatus.Set('sT', (Date.now() - beforeTransactionTime));
//             } else {
//                 // const beforeTransactionTime = Date.now();
//                 // //let signature = Buffer.from(signedProposal.signature.data);
//                 // //let payload = Buffer.from(signedProposal.payload.data);
//                 // let signature = signedProposal.signature;
//                 // let payload =  signedProposal.payload;
//                 // broadcastResponsePromise = channel.sendSignedTransaction({
//                 //     signedProposal: {signature: signature, payload: payload},
//                 //     request: signedTransaction.transactionRequest,
//                 // });
//                 // invokeStatus.Set('sT', (Date.now() - beforeTransactionTime));
//                 // invokeStatus.Set('invokeLatency', (Date.now() - startTime));
//             }
//
//             // TODO SEND POST API
//             broadcastResponse = await broadcastResponsePromise;
//         } catch (err) {
//             commLogger.error('Failed to send transaction error: ' + err);
//             // missing the ACK does not mean anything, the Tx could be already under ordering
//             // so let the events decide the final status, but log this error
//             errFlag |= TxErrorEnum.OrdererResponseError;
//             invokeStatus.SetFlag(errFlag);
//             invokeStatus.SetErrMsg(TxErrorIndex.OrdererResponseError,
//                 err.toString());
//         }
//
//         invokeStatus.Set('time_order', Date.now());
//
//         // if status code === 200 ~ 299
//         if (broadcastResponse && broadcastResponse.status === 'SUCCESS') {
//             invokeStatus.Set('status', 'submitted');
//
//             // else status code !== 200
//         } else if (broadcastResponse && broadcastResponse.status !==
//             'SUCCESS') {
//             let err = new Error('Received rejection from orderer service: ' +
//                 broadcastResponse.status);
//             errFlag |= TxErrorEnum.BadOrdererResponseError;
//             invokeStatus.SetFlag(errFlag);
//             invokeStatus.SetErrMsg(TxErrorIndex.BadOrdererResponseError,
//                 err.toString());
//             // the submission was explicitly rejected, so the Tx will definitely not be ordered
//             invokeStatus.SetVerification(true);
//             throw err;
//         }
//
//         await Promise.all(eventPromises);
//         // if the Tx is not verified at this point, then every eventhub connection failed (with resolve)
//         // so mark it failed but leave it not verified
//         if (!invokeStatus.IsVerified()) {
//             invokeStatus.SetStatusFail();
//             commLogger.error(
//                 'Failed to complete transaction [' + txId.substring(0, 5) +
//                 '...]: every eventhub connection closed');
//         } else {
//             invokeStatus.SetStatusSuccess();
//         }
//     } catch (err) {
//         // at this point the Tx should be verified
//         invokeStatus.SetStatusFail();
//         commLogger.error(
//             'Failed to complete transaction [' + txId.substring(0, 5) +
//             '...]:' + (err instanceof Error ? err.stack : err));
//     }
//     return invokeStatus;
// }

/**
 * Submit a transaction to the given chaincode with the specified options.
 * @param {object} context The Fabric context.
 * @param {string} id The name of the chaincode.
 * @param {string} version The version of the chaincode.
 * @param {string[]} args The arguments to pass to the chaincode.
 * @param {number} timeout The timeout for the transaction invocation.
 * @return {Promise<TxStatus>} The result and stats of the transaction invocation.
 */
async function invokebycontext(context, id, version, args, timeout) {

  // timestamps are recorded for every phase regardless of success/failure
  let invokeStatus;

  if (context.engine) {
    context.engine.submitCallback(1);
  }

  const txIdObject = context.client.newTransactionID();
  const txId = txIdObject.getTransactionID().toString();
  invokeStatus = new TxStatus(txId);
  // invokeStatus = new TxStatus('');

  let errFlag = TxErrorEnum.NoError;
  invokeStatus.SetFlag(errFlag);

  const fcn = args[0];
  args.shift();

  const proposalRequest = {
    chaincodeId: id,
    fcn,
    args,
  };

  const {
    hostname = 'localhost',
    port = '8080',
  } = boosterConfig;

  let url = '';
  let form = null;

  switch (proposalRequest.fcn) {
    case 'registerMember':
      url = '/transaction/registeruser';
      form = {
        txID: args[0],
        memberId: args[1],
        vsCode: args[2],
        countryCode: args[3],
        currencyCode: args[4],
        memberRole: args[5],
        walletAddress: args[6],
        txTime: args[7],
      };
      break;

    case 'transferCoin':
      url = '/transfercoin';
      form = {
        txID: args[0],
        senderWalletAddress: args[0],
        receiverWalletAddress: args[1],
        amount: args[2],
        fee: args[3],
        txFlag: args[4],
        txTime: args[5],
      };
      break;

    default:
      break;
  }

  invokeStatus.Set('sTP', 0);
  invokeStatus.Set('time_endorse', Date.now());
  const beforeTransactionTime = Date.now();

  return new Promise(resolve => {

    const options = {
      port,
      hostname,
      method: 'POST',
      path: url,
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
    };

    const postData = qs.stringify(form);

    const req = client.request(options);
    req.setTimeout(timeout);
    req.write(postData);
    req.end();

    req.on('request', res => {
      invokeStatus.Set('sT', (Date.now() - beforeTransactionTime));
      invokeStatus.Set('status', 'submitted');

    }).on('response', res => {
      const { statusCode } = res;

      commLogger.debug(
          `#### REQUEST: ${hostname}${url}, ${postData}  RESPONSE: ${statusCode} ####`);

      switch (statusCode) {
        case 200:
        case 201:
          invokeStatus.SetStatusSuccess();
          return resolve(invokeStatus);

        default: // NOT 200, 201 HTTP STATUS CODE
          errFlag |= TxErrorEnum.BadEventNotificationError;
          invokeStatus.SetFlag(errFlag);
          invokeStatus.SetErrMsg(
              TxErrorIndex.BadEventNotificationError,
              'Invalid Transaction');
          invokeStatus.SetStatusFail();

          return resolve(invokeStatus);
      }
    }).on('error', err => {
      errFlag |= TxErrorEnum.BadProposalResponseError;
      invokeStatus.SetFlag(errFlag);
      invokeStatus.SetStatusFail();

      commLogger.error('Failed to Complete Transaction :' +
          (err instanceof Error ? err.stack : err));

      return resolve(invokeStatus);

    }).on('timeout', () => {
      invokeStatus.SetStatusFail();
      req.end();

      commLogger.error('Failed to Complete Transaction : Timeout');

      return resolve(invokeStatus);
    });

  });
}

/**
 * Submit a query to the given chaincode with the specified options.
 * @param {object} context The Fabric context.
 * @param {string} id The name of the chaincode.
 * @param {string} version The version of the chaincode.
 * @param {string} name The single argument to pass to the chaincode.
 * @param {string} fcn The chaincode query function name.
 * @return {Promise<object>} The result and stats of the transaction invocation.
 */
async function querybycontext(context, id, version, name, fcn) {
  const client = context.client;
  const channel = context.channel;
  const tx_id = client.newTransactionID();
  const txStatus = new TxStatus(tx_id.getTransactionID());

  // send query
  const request = {
    chaincodeId: id,
    chaincodeVersion: version,
    txId: tx_id,
    fcn: fcn,
    args: [name],
  };

  if (context.engine) {
    context.engine.submitCallback(1);
  }

  const responses = await channel.queryByChaincode(request);
  if (responses.length > 0) {
    const value = responses[0];
    if (value instanceof Error) {
      throw value;
    }

    for (let i = 1; i < responses.length; i++) {
      if (responses[i].length !== value.length ||
          !responses[i].every(function(v, idx) {
            return v === value[idx];
          })) {
        throw new Error('conflicting query responses');
      }
    }

    txStatus.SetStatusSuccess();
    txStatus.SetResult(responses[0]);
    return Promise.resolve(txStatus);
  } else {
    throw new Error('no query responses');
  }
}

/**
 * Utility method to recursively resolve the tlsCACerts paths listed within the passed json object
 * @param {Object} jsonObj a json object defining a common connection profile
 */
function resolveTlsCACerts(jsonObj) {
  if (typeof jsonObj === 'object') {
    Object.entries(jsonObj).forEach(([key, value]) => {
      // key is either an array index or object key
      if (key.toString() === 'tlsCACerts') {
        value.path = commUtils.resolvePath(value.path);
        return;
      } else {
        resolveTlsCACerts(value);
      }
    });
  } else {
    return;
  }
}

/**
 * Create and return an InMemoryWallet for a user in the org
 * @param {String} org the org
 * @returns {InMemoryWallet} an InMemoryWallet
 */
async function createInMemoryWallet(org) {
  const orgConfig = ORGS[org];
  const cert = fs.readFileSync(commUtils.resolvePath(orgConfig.user.cert)).
      toString();
  const key = fs.readFileSync(commUtils.resolvePath(orgConfig.user.key)).
      toString();
  const inMemoryWallet = new InMemoryWallet();

  await inMemoryWallet.import(orgConfig.user.name,
      X509WalletMixin.createIdentity(orgConfig.mspid, cert, key));

  if (ORGS.orderer.url.startsWith('grpcs')) {
    const fabricCAEndpoint = orgConfig.ca.url;
    const caName = orgConfig.ca.name;
    const tlsInfo = await tlsEnroll(fabricCAEndpoint, caName);
    await inMemoryWallet.import('tlsId',
        X509WalletMixin.createIdentity(org, tlsInfo.certificate,
            tlsInfo.key));
  }

  return inMemoryWallet;
}

/**
 * Retrieve the Gateway object for use in subsequent network invocation commands
 * @param {String} ccpPath the path to the common connection profile for the network
 * @param {String} opts the name of the organisation to use
 * @returns {Network} the Fabric Network object
 */
async function retrieveGateway(ccpPath, opts) {
  const gateway = new Gateway();

  ccpPath = commUtils.resolvePath(ccpPath);
  let ccp = JSON.parse(fs.readFileSync(ccpPath).toString());

  // need to resolve tlsCACerts paths for current system
  resolveTlsCACerts(ccp);

  await gateway.connect(ccp, opts);
  return gateway;
}

/**
 * Submit a transaction to the given chaincode with the specified options.
 * @param {object} context The Fabric context.
 * @param {string[]} args The arguments to pass to the chaincode.
 * @return {Promise<TxStatus>} The result and stats of the transaction invocation.
 */
async function submitTransaction(context, args) {
  const TxErrorEnum = require('./constant.js').TxErrorEnum;
  const txIdObject = context.gateway.client.newTransactionID();
  const txId = txIdObject.getTransactionID().toString();

  // timestamps are recorded for every phase regardless of success/failure
  let invokeStatus = new TxStatus(txId);
  let errFlag = TxErrorEnum.NoError;
  invokeStatus.SetFlag(errFlag);

  if (context.engine) {
    context.engine.submitCallback(1);
  }

  try {
    const result = await context.contract.submitTransaction(...args);
    invokeStatus.result = result;
    invokeStatus.verified = true;
    invokeStatus.SetStatusSuccess();
    return invokeStatus;
  } catch (err) {
    commLogger.error(
        'failed to submit transaction using args [' + JSON.stringify(args) +
        '], with error: ' + (err instanceof Error ? err.stack : err));
    invokeStatus.SetStatusFail();
    invokeStatus.result = [];
    return Promise.resolve(invokeStatus);
  }
}

/**
 * Evaluates the given chaincode function with the specified options; this will not append to the ledger
 * @param {object} context The Fabric context.
 * @param {string[]} args The arguments to pass to the chaincode.
 * @return {Promise<TxStatus>} The result and stats of the transaction invocation.
 */
async function evaluateTransaction(context, args) {
  const TxErrorEnum = require('./constant.js').TxErrorEnum;
  const txIdObject = context.gateway.client.newTransactionID();
  const txId = txIdObject.getTransactionID().toString();

  // timestamps are recorded for every phase regardless of success/failure
  let invokeStatus = new TxStatus(txId);
  let errFlag = TxErrorEnum.NoError;
  invokeStatus.SetFlag(errFlag);

  if (context.engine) {
    context.engine.submitCallback(1);
  }

  try {
    const result = await context.contract.evaluateTransaction(...args);
    invokeStatus.result = result;
    invokeStatus.SetStatusSuccess();
    return invokeStatus;
  } catch (err) {
    commLogger.error('failed to evaluate transaction using args [' +
        JSON.stringify(args) + '], with error: ' +
        (err instanceof Error ? err.stack : err));
    invokeStatus.SetStatusFail();
    invokeStatus.result = [];
    return Promise.resolve(invokeStatus);
  }
}

module.exports.init = init;
module.exports.installChaincode = installChaincode;
module.exports.instantiateChaincode = instantiateChaincode;
module.exports.getcontext = getcontext;
module.exports.releasecontext = releasecontext;
module.exports.invokebycontext = invokebycontext;
module.exports.querybycontext = querybycontext;
module.exports.tlsEnroll = tlsEnroll;
module.exports.createInMemoryWallet = createInMemoryWallet;
module.exports.retrieveGateway = retrieveGateway;
module.exports.submitTransaction = submitTransaction;
module.exports.evaluateTransaction = evaluateTransaction;
