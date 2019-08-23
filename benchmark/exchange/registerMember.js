/*******************************************************************************
 * Copyright 2019. Nuri Telecom. All Rights Reserved.
 *
 * Sungyub NA <mailto: syna@nuritelecom.com>
 ******************************************************************************/

'use strict';
const { randomBytes } = require('crypto');
const moment = require('moment');

const { v4: uuid } = require('uuid');

const info = 'register users';

let account_array = [],
    txnPerBatch,
    bc,
    contx;

const init = function(blockchain, context, args) {
  if (!args.hasOwnProperty('txnPerBatch')) {
    args.txnPerBatch = 1;
  }
  txnPerBatch = args.txnPerBatch;
  bc = blockchain;
  contx = context;

  return Promise.resolve();
};

/**
 * Generate Register Member (exchange)
 * @returns {Array} workload
 */
function generateWorkload() {
  let workload = [];
  for (let i = 0; i < txnPerBatch; i++) {
    const userId = randomBytes(20).toString('hex');
    account_array.push(userId);

    workload.push({
      func: 'registerMember',
      txId: uuid(),
      memberId: userId,
      vsCode: 'v1',
      countryCode: 'ghana',
      currencyCode: 'cedi',
      memberRole: 'test',
      walletAddress: userId,
      createdDate: moment().format('YY-MM-DD HH:mm:ss'),
    });
  }
  return workload;
}

/**
 * Start Test (exchange)
 * @returns {Promise<Object>}
 * invokeSmartContract(contx, chaincodeid, version, args, timeout)
 */
const run = function() {
  let args = generateWorkload();
  return bc.invokeSmartContract(contx, 'exchange', '1.0', args, 50000);
};

const end = function() {
  return Promise.resolve();
};

module.exports = {
  info,
  init,
  run,
  end,
};
