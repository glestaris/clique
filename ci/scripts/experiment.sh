#!/bin/bash
root_path=$(cd $(dirname $BASH_SOURCE); pwd)
set -e -x

$root_path/setup.sh

cd clique/
make
ginkgo experiment
