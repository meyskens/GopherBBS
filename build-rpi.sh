#!/bin/bash
export GOARCH=arm
export GOARM=6

go build ./cmd/gopherbbs

scp gopherbbs pi@192.168.0.14:/home/pi
