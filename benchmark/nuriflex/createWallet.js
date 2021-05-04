/*
 * Copyright 2021. NuriFlex Co., Ltd. All Rights Reserved.
 *
 * - createWallet.js
 * - author: Sungyub NA <mailto: syna@nuriflex.co.kr>
 */

const { randomBytes } = require('crypto');
const moment = require('moment');
const info = 'create wallet';

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
      func: 'createWallet',
      wallet_id: userId,
      balance: '0',
      user_id: `user-${userId}`,
      user_name: `user-${userId}`,
      wallet_status: 'active',
      created: moment().toISOString(),
      token_id: 'general_token_id',
      token_name: 'general_token',
      token_type: 'GENERL'
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

  if (bc.bcType === 'booster') {
    return bc.invokeSmartContract(contx, 'nuriflex', '1.0', args, 15000);
  }

  return bc.invokeSmartContract(contx, 'nuriflex', '1.0', args, 50000);
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
