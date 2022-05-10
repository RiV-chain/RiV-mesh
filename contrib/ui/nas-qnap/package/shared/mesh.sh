#!/bin/sh
CONF="/etc/config/mesh.conf"
QPKG_NAME="mesh"
QPKG_DIR=$(/sbin/getcfg $QPKG_NAME Install_Path -f $CONF)
CONFIG_DIR="/etc/config"
export smb_conf_file=/etc/smb.conf

start_service ()
{

    # Launch the mesh in the background.
    ${QPKG_DIR}/bin/mesh -useconffile "$CONF" \
    -httpaddress "http://127.0.0.1:19019" \
    -wwwroot "$QPKG_DIR/www" \
    -logto "$QPKG_DIR/var/log/mesh.log" &
    if [ $? -ne 0 ]; then
      echo "Starting $QPKG_NAME failed"
      exit 1
    fi
}

stop_service ()
{
    # Kill mesh
    pid=`pidof -s mesh`
    if [ -z "$pid" ]; then
      echo "mesh was not running"
      exit 0
    fi
    kill "$pid"
}

case "$1" in
  start)
    ENABLED=$(/sbin/getcfg $QPKG_NAME Enable -u -d FALSE -f $CONF)
    if [ "$ENABLED" != "TRUE" ]; then
        echo "$QPKG_NAME is disabled."
        exit 1
    fi
    start_service
    ;;

  stop)
    stop_service
    ;;

  restart)
    $0 stop
    $0 start
    ;;

  *)
    echo "Usage: $0 {start|stop|restart}"
    exit 1
esac

exit 0
