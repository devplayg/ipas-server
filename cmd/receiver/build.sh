#!/bin/sh

go build receiver.go
go build ../classifier/classifier.go
go build ../generator/generator.go
