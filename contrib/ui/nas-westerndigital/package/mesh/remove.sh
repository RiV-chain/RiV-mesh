#!/bin/sh

MESH_PACKAGE_LOG=/var/log/mesh.log
echo "remove.sh called" >> "$MESH_PACKAGE_LOG"
inst_path="$1"

rm -f /usr/bin/mesh
rm -f /usr/bin/meshctl
rm -fr /var/www/meshpkg
rm -f /usr/local/apache2/conf/extra/apache-mesh.conf
rm -fr "$inst_path"

( sleep 2 ; /usr/sbin/apache restart web ) &
