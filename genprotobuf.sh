#!/bin/bash

protoc --go_out=paths=source_relative:. protobuf/Message.proto
