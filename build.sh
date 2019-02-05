#!/usr/bin/env bash
GOOS=linux GOARCH=amd64 go build -v .
mv hinko builds/linuxamd64
