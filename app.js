/*
 * Copyright 2019. Nuri Telecom. All Rights Reserved.
 *
 * - app.js
 * - author: Sungyub NA <mailto: syna@nuritelecom.com>
 */

'use strict';

const { fork } = require('child_process');
const config = require('./src/comm/config-util');
const io = require('socket.io')();

const port = config.getConfigSetting('socket-io:port', 3000);

let isCaliperRunning = false;

const startCaliper = (path, args, callback) => {
  let invoked = false;

  let process = fork(path, args);

  process.on('message', msg => {
    console.log('child send msg');
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
const checkCaliperRunning = (socket, target) => {
  if (!isCaliperRunning) {
    socket.emit('status', { status: 'start' });

    let { path, args } = target;

    return startCaliper(path, args, err => {
      if (err) {
        throw err;
      }
    });
  }
  return socket.emit('status', { status: 'busy' });
};

io.of('/start').on('connection', socket => {
  console.log(`socket.i0 new connection: ${socket.id}`);

  socket.on('test', target => {
    checkCaliperRunning(socket, target);
  });

});

io.listen(port);

// app.use(async ctx => {
//   ctx.body = 'hello world';
//
//   startCaliper(`${__dirname}/scripts/main.js`, [
//     '-c',
//     'network/nuritelecom/exchange-bc-tls/config.yaml',
//     '-n',
//     'network/nuritelecom/exchange-bc-tls/fabric-go-tls.json',
//   ], err => {
//     if (err) {
//       throw err;
//     }
//   });
//
// });

