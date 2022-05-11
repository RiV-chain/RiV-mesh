#!/bin/sh
CONF="/etc/config/mesh.conf"
QPKG_NAME="mesh"
QPKG_DIR=$(/sbin/getcfg $QPKG_NAME Install_Path -f $CONF)
CONFIG_DIR="/etc/config"

start_service ()
{
    #enable ipv6    
    sysctl -w net.ipv6.conf.all.disable_ipv6=0
    sysctl -w net.ipv6.conf.default.disable_ipv6=0
    echo sbin/getcfg $QPKG_NAME Install_Path -f $CONF > /tmp/mesh.log
    echo $SYS_QPKG_DIR/bin/mesh >> /tmp/mesh.log

    . /etc/init.d/vpn_common.sh && load_kernel_modules
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
