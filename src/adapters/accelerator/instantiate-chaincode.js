/*
 * Copyright 2019. Nuri Telecom. All Rights Reserved.
 *
 * - instantiate-chaincode.js
 * - author: Sungyub NA <mailto: syna@nuritelecom.com>
 */

'use strict';

const e2eUtils = require('./e2eUtils.js');
const commUtils = require('../../comm/util');
// const commLogger = commUtils.getLogger('instantiate-chaincode.js');

const commLogger = {
    debug: (msg) => {
        process.send({
            type: 'socket.io',
            data: {
                message: msg
            }
        });
        return commUtils.getLogger('instantiate-chaincode.js').debug(msg);
    },
    info: (msg) => {
        process.send({
            type: 'socket.io',
            data: {
                message: msg
            }
        });
        return commUtils.getLogger('instantiate-chaincode.js').info(msg);
    },
    warn: (msg) => {
        process.send({
            type: 'socket.io',
            data: {
                message: msg
            }
        });
        return commUtils.getLogger('instantiate-chaincode.js').warn(msg);
    },
    error: (msg) => {
        process.send({
            type: 'socket.io',
            data: {
                message: msg
            }
        });
        return commUtils.getLogger('instantiate-chaincode.js').error(msg);
    }
};

/**
 * Install the chaincode listed within config
 * @param {*} config_path The path to the Fabric network configuration file.
 * @async
 */
async function run(config_path) {
    const config = commUtils.parseYaml(config_path);
    const fabricSettings = config.fabric;
    const policy = fabricSettings['endorsement-policy'];  // TODO: support multiple policies
    let chaincodes = fabricSettings.chaincodes;
    if(typeof chaincodes === 'undefined' || chaincodes.length === 0) {
        return;
    }

    try {
        commLogger.info('Instantiating chaincodes...');
        for (let chaincode of chaincodes) {
            await e2eUtils.instantiateChaincode(chaincode, policy, false);
            commLogger.info(`Instantiated chaincode ${chaincode.id} successfully`);
        }

        commLogger.info('Sleeping 5s...');
        await commUtils.sleep(5000);
    } catch (err) {
        commLogger.error(`Failed to instantiate chaincodes: ${(err.stack ? err.stack : err)}`);
        throw err;
    }
}

module.exports.run = run;