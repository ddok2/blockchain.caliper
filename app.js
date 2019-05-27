/*
 * Copyright 2019. Nuri Telecom. All Rights Reserved.
 *
 * - app.js
 * - author: Sungyub NA <mailto: syna@nuritelecom.com>
 */

'use strict';

const express = require('express');
const app = express();
const { Server } = require('http');

const server = Server(app);
const io = require('socket.io')(server);

const { fork } = require('child_process');
const config = require('./src/comm/config-util');
const path = require('path');
const logger = require('./src/comm/util').getLogger('app.js');

const port = config.getConfigSetting('socket-io:port', 3000);

let isCaliperRunning = false;

const startCaliper = (socket, path, args, callback) => {
  let invoked = false;

  let process = fork(path, args);

  process.on('message', msg => {
    io.emit('status', msg);

  }).on('error', function(err) {
    if (invoked) {
      return;
    }
    invoked = true;
    callback(err);

  }).on('exit', function(code) {
    if (invoked) {
      return;
    }
    invoked = true;
    let err = code === 0 ? null : new Error('exit code ' + code);
    callback(err);
  });
};
const validate = {
  testMode: mode => {
    switch (mode) {
      case 'booster':
        return [
          '-c',
          'network/nuritelecom/exchange-bc-tls/config.yaml',
          '-n',
          'network/nuritelecom/exchange-bc-tls/booster-go-tls.json',
        ];

      case 'fabric':
      default:
        return [
          '-c',
          'network/nuritelecom/exchange-bc-tls/config.yaml',
          '-n',
          'network/nuritelecom/exchange-bc-tls/fabric-go-tls.json',
        ];
    }
  },
  running: (socket, target) => {
    if (!isCaliperRunning) {
      isCaliperRunning = true;
      socket.emit('status', { status: 'start' });

      let { path, mode } = target;

      if (!path) {
        path = `${__dirname}/scripts/main.js`;
      }

      return startCaliper(
          socket,
          path,
          validate.testMode(mode),
          err => {
            isCaliperRunning = false;
            if (err) {
              throw err;
            }
          });
    }
    return socket.emit('status', { status: 'busy' });
  },
};

app.use('/results', express.static(path.join(__dirname, 'results')));

io.on('connection', socket => {
  logger.info(`socket.io new connection: ${socket.id}`);

  socket.on('start-test', target => {
    validate.running(socket, target);
  });

});

server.listen(port, () => {
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
        // console.log('\n');
        console.log(
            `NURI Blockchain Caliper listening on port: ${port}`);
        console.log('\n');
        console.log(`pid is ${process.pid}`);
        console.log('\n');
      });
});
