/*
 * Copyright 2019. Nuri Telecom. All Rights Reserved.
 *
 * - main.js
 * - author: Sungyub NA <mailto: syna@nuritelecom.com>
 */

'use strict';

const path = require('path');
const fs = require('fs-extra');
const logger = require('../src/comm/util').getLogger('scripts/standalone.js');

const framework = require('../src/comm/bench-flow.js');
const program = require('commander');
const { fork } = require('child_process');
const config = require('../src/comm/config-util');

const mkResultsDirSync = () => {
  let dirname = `${__dirname}/results`;

  if (!fs.existsSync(dirname)) {
    fs.mkdirSync(dirname);
  }
};

const startCaliper = (path, args, callback) => {
  let invoked = false;
  let process = fork(path, args);

  process.on('message', msg => {
    // logger.info(msg ? msg.data.message : '');
  }).on('error', err => {
    if (invoked) {
      return;
    }
    invoked = true;
    callback(err);

  }).on('exit', code => {
    if (invoked) {
      return;
    }
    invoked = true;
    let err = code === 0 ? null : new Error('exit code ' + code);
    callback(err);
  });
};

async function start() {
  mkResultsDirSync();

  program.allowUnknownOption().option('-m, --mode <mode>',
      'test mode: [fabric, booster, online, localhost, raft]',
  ).on('--help', () => {
    console.log('');
    console.log('Examples:');
    console.log('   $ ./start-standalone -m booster');
    console.log('   $ ./start-standalone --mode online');
  }).parse(process.argv);

  let absConfigFile, absNetworkFile;

  if (typeof program.mode === 'undefined') {
    logger.error(`
    test mode is required.
    Usage: ./start-standalone -m [options]
    help: ./start-standalone --help
    `);
    return;
  } else {
    let { mode } = program;

    switch (mode) {
      case 'booster':
        absConfigFile =
            path.join(__dirname,
                '/../',
                'network/nuritelecom/exchange-bc-production_raft' +
                '/caliper-config/host-config.yaml');
        absNetworkFile =
            path.join(__dirname,
                '/../',
                'network/nuritelecom/exchange-bc-production_raft' +
                '/caliper-config/booster-go-tls.json');
        break;

      case 'localhost':
        absConfigFile =
            path.join(__dirname,
                '/../',
                'network/nuritelecom/exchange-bc-production-v1.0' +
                '/caliper-config/host-config.yaml');
        absNetworkFile =
            path.join(__dirname,
                '/../',
                'network/nuritelecom/exchange-bc-production-v1.0' +
                '/caliper-config/local-go-tls.json');
        break;

      case 'online':
        absConfigFile =
            path.join(__dirname,
                '/../',
                'network/nuritelecom/exchange-bc-production-v1.0' +
                '/caliper-config/host-config.yaml');
        absNetworkFile =
            path.join(__dirname,
                '/../',
                'network/nuritelecom/exchange-bc-production-v1.0' +
                '/caliper-config/host-go-tls.json');
        break;

      case 'raft':
        absConfigFile =
            path.join(__dirname,
                '/../',
                'network/nuritelecom/exchange-bc-production_raft' +
                '/caliper-config/host-config.yaml');
        absNetworkFile =
            path.join(__dirname,
                '/../',
                'network/nuritelecom/exchange-bc-production_raft' +
                '/caliper-config/fabric-go-tls.json');
        break;

      case 'fabric':
      default:
        absConfigFile =
            path.join(__dirname,
                '/../',
                'network/nuritelecom/exchange-bc-production' +
                '/caliper-config/config.yaml');
        absNetworkFile =
            path.join(__dirname,
                '/../',
                'network/nuritelecom/exchange-bc-production' +
                '/caliper-config/fabric-go-tls.json');
        break;
    }
  }

  if (!fs.existsSync(absConfigFile)) {
    logger.error('file ' + absConfigFile + ' does not exist');
    return;
  }
  if (!fs.existsSync(absNetworkFile)) {
    logger.error('file ' + absNetworkFile + ' does not exist');
    return;
  }

  try {
    // await framework.run(absConfigFile, absNetworkFile);
    require('figlet')(
        `
      NURI
      Blockchain
      Caliper
      `, {
          font: 'ANSI Shadow',
          kerning: 'fitted',
        }, (err, data) => {
          console.log(data);
          console.log(
              `NURI Blockchain Caliper Test Start`);
          console.log('\n');
          console.log(`pid is ${process.pid}`);
          console.log('\n');

          return startCaliper(
              `${__dirname}/main.js`,
              ['-c', absConfigFile, '-n', absNetworkFile],
              err => {
                if (err) {
                  throw err;
                }
              });
        });
  } catch (err) {
    logger.error(`Error while executing the benchmark: ${err.stack ?
        err.stack :
        err}`);
    process.exit(1);
  }
}

start().catch(() => {process.exit(1);});
