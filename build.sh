#!/bin/bash

go build -o build/mq        mq.go route.go endpoint.go store.go uuid.go
go build -o build/mq-mover  bin/mover.go bin/watch.go
go build -o build/mq-reaper bin/reaper.go bin/watch.go