#!/usr/bin/env bash

go test -race -v ./...
ret=$?; if [[ $ret !=0 && $ret != 1 ]]; then
    echo "Test failed, exit $ret"
    exit $ret
fi
