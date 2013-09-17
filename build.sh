#!/bin/bash

set -e

GO=~/go/bin/go
USER=root
HOST=198.23.111.7

$GO build -o build/mq        mq.go route.go endpoint.go store.go uuid.go
$GO build -o build/mq-mover  bin/mover.go bin/watch.go
$GO build -o build/mq-reaper bin/reaper.go bin/watch.go

scp etc/sysctl.conf $USER@$HOST:/etc/sysctl.conf
scp etc/security/limits.conf $USER@$HOST:/etc/security/limits.conf
scp etc/pam.d/common-session $USER@$HOST:/etc/pam.d/common-session
scp etc/supervisor/conf.d/mq.conf $USER@$HOST:/etc/supervisor/conf.d/mq.conf

ssh $USER@$HOST 'supervisorctl reread; supervisorctl update'
ssh $USER@$HOST 'supervisorctl stop all'

for x in mq mq-mover mq-reaper
do
	scp build/$x $USER@$HOST:~/$x
done

ssh $USER@$HOST 'supervisorctl start all'