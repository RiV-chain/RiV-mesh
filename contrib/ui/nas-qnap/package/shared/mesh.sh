#!/bin/sh
CONF="/etc/config/mesh.conf"
QPKG_NAME="mesh"
QPKG_DIR=$SYS_QPKG_DIR
CONFIG_DIR="/etc/config"

start_service ()
{
    #enable ipv6    
    sysctl -w net.ipv6.conf.all.disable_ipv6=0
    sysctl -w net.ipv6.conf.default.disable_ipv6=0

    . /etc/init.d/vpn_common.sh && load_kernel_modules

    if [ ! -f '/etc/config/apache/extra/apache-mesh.conf' ] ; then
      ln -sf $SYS_QPKG_DIR/apache-mesh.conf /etc/config/apache/extra/
      apache_reload=1
    fi    
    
    if ! grep '/etc/config/apache/extra/apache-mesh.conf' /etc/config/apache/apache.conf ; then
      echo 'Include /etc/config/apache/extra/apache-mesh.conf' >> /etc/config/apache/apache.conf
      apache_reload=1
    fi

    if [ -n "$apache_reload" ] ; then
      /usr/local/apache/bin/apachectl -k graceful
    fi
    
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
