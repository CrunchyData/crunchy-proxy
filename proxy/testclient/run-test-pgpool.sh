#!/bin/bash
go run testclient.go \
	-rows=onerow \
	-hostport=localhost:12003 \
	-userid=testuser \
	-password=password \
	-database=userdb 
