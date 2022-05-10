#!/bin/sh

log_exit(){
	echo $2
	exit $1
}

[ -z "$ED_USER_NAME" ] && log_exit 1 "Credentials are not set. Remove aborted"

rm "$ED_APP_ROOT/bin/mesh
rm -rf "$ED_APP_ROOT/www"

