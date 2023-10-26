#!/bin/bash
set -ex

chmod +x $1
$1 >>console.log 2>&1 &


