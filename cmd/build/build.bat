@echo off

go build ../receiver/receiver.go
go build ../classifier/classifier.go
go build ../generator/generator.go
go build ../calculator/calculator.go