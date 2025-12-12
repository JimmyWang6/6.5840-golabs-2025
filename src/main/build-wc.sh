#!/bin/bash
go build -buildmode=plugin -gcflags=all=-N -l -o ../mrapps/wc.so ../mrapps/wc.go
rm -rf mr-*