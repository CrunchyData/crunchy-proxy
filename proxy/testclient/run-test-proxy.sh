#!/bin/bash
go run testclient.go \
	-rows=onerow \
	-hostport=localhost:5432 \
	-userid=postgres \
	-password=password \
	-database=postgres 
