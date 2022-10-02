#!/bin/sh

GOARCH=wasm GOOS=js go build -o libgo.wasm go/main.go
go run server.go
