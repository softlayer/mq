#!/bin/bash

# Echo a list of volume mount points.
# echo `for i in {a..z}; do echo /mnt/vol0$i; done` | tr " " ","

# Mount all volumes.
# for i in {a..z}; do mkdir /mnt/vol0$i; mount -t glusterfs gluster-00:/vol0 /mnt/vol0$i; done

set -e

GO=~/go/bin/go
USER=root
HOST=198.23.111.7

$GO build -o build/mq        mq.go route.go endpoint.go store.go uuid.go
$GO build -o build/mq-mover  bin/mover.go
$GO build -o build/mq-reaper bin/reaper.go

scp etc/supervisor/conf.d/mq.conf $USER@$HOST:/etc/supervisor/conf.d/mq.conf

ssh $USER@$HOST 'supervisorctl reread; supervisorctl update'
ssh $USER@$HOST 'supervisorctl stop all'

sleep 5

for x in mq mq-mover mq-reaper
do
	scp build/$x $USER@$HOST:~/$x
done

ssh $USER@$HOST 'supervisorctl start all'