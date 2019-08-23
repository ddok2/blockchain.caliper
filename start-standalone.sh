#!/bin/bash

if [[ $# -lt 1 ]]; then
    node ./scripts/standalone.js

elif [[ $1 == "--help" ]]; then
    node ./scripts/standalone.js $1

else
    echo "### Please checkout 'output.log' file ###"
#    node ./scripts/standalone.js $1 $2 > output.log
    node ./scripts/standalone.js $1 $2 | tee output.log
fi
