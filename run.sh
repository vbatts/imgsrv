#!/bin/sh

binary=${APP_BINARY:-/usr/local/bin/imgsrv}
MONGODB_SERVICE_HOST=${MONGODB_SERVICE_HOST:-localhost}
MONGODB_SERVICE_PORT=${MONGODB_SERVICE_PORT:-27017}

args=""
if [ ! -z "${DATABASE_USER}" ] ; then
	args=" ${args} -mongo-username=$DATABASE_USER "
fi

if [ ! -z "${DATABASE_PASSWORD}" ] ; then
	args=" ${args} -mongo-password=$DATABASE_PASSWORD "
fi

if ! echo "${1:-$@}" | grep -q '\-mongo-host' ; then 
	args=" ${args} -mongo-host=$MONGODB_SERVICE_HOST:$MONGODB_SERVICE_PORT "
fi

${binary} \
	-server \
	-mongo-db=$MONGODB_DATABASE \
	${1:-$@} \
	${args}
