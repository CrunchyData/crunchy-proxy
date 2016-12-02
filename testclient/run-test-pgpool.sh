#!/bin/bash
go run testclient.go \
	-rows=onerow \
	-count=100 \
	-hostport=localhost:12003 \
	-userid=testuser \
	-password=password \
	-database=userdb 
