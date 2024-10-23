#!/bin/bash
set -e
cd $(dirname $0)
cd example
go build -trimpath -ldflags "-w -s"
./example
