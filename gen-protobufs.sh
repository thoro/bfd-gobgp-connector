#!/usr/bin/env bash

set -e
set -x

protoc \
    --go_out=plugins=grpc:. \
    bfd-api/bfd-api.proto

protoc \
    -I./vendor/github.com/golang/protobuf/ptypes \
    -I./gobgp-api \
    --go_out=plugins=grpc:gobgp-api \
    gobgp-api/*.proto
