#!/bin/bash

set -e

GO=~/go/bin/go
USER=root
HOST=108.168.233.11

$GO build -o build/mq        mq.go route.go endpoint.go store.go uuid.go
$GO build -o build/mq-mover  bin/mover.go
$GO build -o build/mq-reaper bin/reaper.go

ssh $USER@$HOST supervisorctl stop all

for x in mq mq-mover mq-reaper
do
	scp build/$x $USER@$HOST:~/$x
done

ssh $USER@$HOST supervisorctl start all