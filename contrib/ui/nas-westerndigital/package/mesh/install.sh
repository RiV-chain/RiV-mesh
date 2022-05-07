#!/bin/sh

INST_PATH=/mnt/HD/HD_a2/Nas_Prog/mesh
MESH_PACKAGE_LOG=/var/log/mesh.log
echo "install.sh called" >> "$MESH_PACKAGE_LOG"

path_src="$1"
path_dst="$2"

rm -rf "$INST_PATH"
mv "$path_src" "$path_dst"
#LOG_SYMLINK
ln -s "$INST_PATH"/var/log/mesh.log "$INST_PATH"/www/log
ln -s "$INST_PATH"/apache-mesh.conf /usr/local/apache2/conf/extra
( sleep 2 ; /usr/sbin/apache restart web ) &
