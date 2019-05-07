#!/bin/sh

f=${1:-FuzzSkip}

go-fuzz -func $f -workdir ${f}_wd -bin nikandjson-fuzz.zip
