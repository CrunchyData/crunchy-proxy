#!/bin/bash
go run testclient.go \
	-count=100 \
	-rows=onerow \
	-hostport=localhost:5432 \
	-userid=postgres \
	-password=password \
	-database=postgres 
