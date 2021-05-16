#!/bin/bash
CURDIR=$(cd $(dirname $0); pwd)

BinaryName="gateway-admin"

exec $CURDIR/bin/${BinaryName} -conf=${CONF_FILE_PATH}