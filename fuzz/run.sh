#!/bin/bash

cd $(dirname $0)

f=${1:-FuzzSkip}

go-fuzz -func $f -workdir ${f}_wd -bin nikandjson-fuzz.zip
