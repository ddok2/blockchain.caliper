/*
 * Copyright 2019. Nuri Telecom. All Rights Reserved.
 *
 * - main.js
 * - author: Sungyub NA <mailto: syna@nuritelecom.com>
 */

'use strict';

const path = require('path');
const fs = require('fs-extra');
const logger = require('../src/comm/util').getLogger('scripts/main.js');

const framework = require('../src/comm/bench-flow.js');
const program = require('commander');

const config = require('../src/comm/config-util');

async function main() {

  // let {
  //   host = 'http://localhost',
  //   port = 4000,
  // } = config.getConfigSetting('socket-io');
  // const io = require('socket.io-client')(`${host}:${port}`);
  // const getSocekt = async () => {
  //   return new Promise((resolve, reject) => {
  //     io.on('connection', socket => {
  //       return resolve(socket);
  //     });
  //   });
  // };

  program.allowUnknownOption().
      option('-c, --config <file>', 'config file of the benchmark').
      option('-n, --network <file>',
          'config file of the blockchain system under test').
      parse(process.argv);

  let absConfigFile;
  if (typeof program.config === 'undefined') {
    logger.error('config file is required');
    return;
  } else {
    absConfigFile = path.isAbsolute(program.config) ?
        program.config :
        path.join(__dirname, '/../', program.config);
  }
  if (!fs.existsSync(absConfigFile)) {
    logger.error('file ' + absConfigFile + ' does not exist');
    return;
  }

  let absNetworkFile;
  if (typeof program.network === 'undefined') {
    logger.error('network file is required');
    return;
  } else {
    absNetworkFile = path.isAbsolute(program.network) ?
        program.network :
        path.join(__dirname, '/../', program.network);
  }
  if (!fs.existsSync(absNetworkFile)) {
    logger.error('file ' + absNetworkFile + ' does not exist');
    return;
  }

  // absConfigFile = path.join(__dirname,
  //     '../network/nuritelecom/exchange-bc-tls/config.yaml');
  // absNetworkFile = path.join(__dirname,
  //     '../network/nuritelecom/exchange-bc-tls/fabric-go-tls.json');

  try {
    // const socket = await getSocekt();
    // socket.emit('status', {status: 'test will start'});

    await framework.run(absConfigFile, absNetworkFile);
    logger.info('Benchmark run successfully');
    process.exit(0);
  } catch (err) {
    logger.error(`Error while executing the benchmark: ${err.stack ?
        err.stack :
        err}`);
    process.exit(1);
  }
}

main();
