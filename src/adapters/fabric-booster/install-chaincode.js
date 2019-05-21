/*
 * Copyright 2019. Nuri Telecom. All Rights Reserved.
 *
 * - install-chaincode.js
 * - author: Sungyub NA <mailto: syna@nuritelecom.com>
 */

'use strict';

const e2eUtils = require('./e2eUtils.js');
const testUtil = require('./util.js');
const commUtils = require('../../comm/util');
// const commLogger = commUtils.getLogger('install-chaincode.js');

const commLogger = {
  debug: (msg) => {
    process.send({
      type: 'socket.io',
      data: {
        message: msg
      }
    });
    return commUtils.getLogger('install-chaincode.js').debug(msg);
  },
  info: (msg) => {
    process.send({
      type: 'socket.io',
      data: {
        message: msg
      }
    });
    return commUtils.getLogger('install-chaincode.js').info(msg);
  },
  warn: (msg) => {
    process.send({
      type: 'socket.io',
      data: {
        message: msg
      }
    });
    return commUtils.getLogger('install-chaincode.js').warn(msg);
  },
  error: (msg) => {
    process.send({
      type: 'socket.io',
      data: {
        message: msg
      }
    });
    return commUtils.getLogger('install-chaincode.js').error(msg);
  }
};

/**
 * Install the chaincode listed within config
 * @param {*} config_path The path to the Fabric network configuration file.
 * @async
 */
async function run(config_path) {
  const fabricSettings = commUtils.parseYaml(config_path).fabric;
  let chaincodes = fabricSettings.chaincodes;
  if (typeof chaincodes === 'undefined' || chaincodes.length === 0) {
    return;
  }

  testUtil.setupChaincodeDeploy();

  try {
    commLogger.info('installing all chaincodes......');

    for (const chaincode of chaincodes) {
      let channel = testUtil.getChannel(chaincode.channel);
      if (channel === null) {
        throw new Error('could not find channel in config');
      }

      for (let orgIndex in channel.organizations) {
        // NOTE: changed execution to sequential for easier debugging (this is a one-time task, performance doesn't matter)
        commLogger.info(`Installing chaincode ${chaincode.id}...`);
        await e2eUtils.installChaincode(channel.organizations[orgIndex],
            chaincode);
      }

      commLogger.info(
          `Installed chaincode ${chaincode.id} successfully in all peers`);

    }
  } catch (err) {
    commLogger.error(
        `Failed to install chaincodes: ${(err.stack ? err.stack : err)}`);
    throw err;
  }
}

module.exports.run = run;