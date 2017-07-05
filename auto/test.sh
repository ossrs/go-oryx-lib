#!/usr/bin/env bash

go test -race -v ./...
ret=$?; if [[ $ret -ne 0 && $ret -ne 1 ]]; then
    echo "Test failed, exit $ret"
    exit $ret
fi
