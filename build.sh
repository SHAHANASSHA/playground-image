#!/bin/bash

go build -o ./firestarter/firestarter firestarter.go

go build -o ./firestarter/ws-server ws-server.go

docker build -t repo.synnefo.solutions/devaraj/2vmpg .