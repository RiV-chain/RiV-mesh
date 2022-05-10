#!/bin/sh
#

##!!
. /etc/service.subr

prog_dir=`dirname \`realpath $0\``
base_dir=/mnt/DroboFS/Shares/DroboApps/mesh
config_dir=$base_dir/config

name="mesh"
framework_version="2.1"
description="Secure cloud backup solution optimized for the Drobo platform"
depends=""
webui="WebUI"
pidfile=/tmp/DroboApps/mesh/pid.txt
daemon=$base_dir/mesh

errorfile=/tmp/DroboApps/mesh/error.txt
statusfile=/tmp/DroboApps/mesh/status.txt
edstatusfile=$base_dir/var/lib/mesh/status

start()
{
	mkdir -p /tmp/DroboApps/mesh
	# delete edstatufile before starting daemon to delete previous status
	rm -f $edstatusfile
	rm -f $errorfile

    if [ -f $config_file ]; then
       mkdir -p /var/backups
       echo "Backing up configuration file to /var/backups/mesh.conf.`date +%Y%m%d`"
       cp $config_file /var/backups/mesh.conf.`date +%Y%m%d`
       echo "Normalising and updating /etc/mesh.conf"
       $daemon -useconf -normaliseconf < /var/backups/mesh.conf.`date +%Y%m%d` > $config_file
    else
       mkdir -p $config_dir
       echo "Generating initial configuration file $config_file"
       echo "Please familiarise yourself with this file before starting RiV-mesh"
       sh -c "umask 0027 && $daemon -genconf > '$config_file'"
    fi
    
    

	if [ -z $(pidof mesh) ]; then
		echo 1 > $errorfile
		echo "Application starting error" > $edstatusfile
	fi
	sleep 1
	update_status
}

update_status()
{

	# wait until file appears
	i=60

	while [ -z $(pidof mesh) ] 
	do
		sleep 1
		i=$((i-1))

		if [ $i -eq 0 ] 
		then
			break
		fi
	done

	# if we don't have file here. throw error into status and return
	if [ -z $(pidof mesh) ] 
	then
		echo 1 > "${errorfile}"
		echo "Configuration required" > $statusfile
		return
        else
        	echo 0 > "${errorfile}"
		echo "Application is running" > $statusfile
	fi

}

stop()
{
		killall $name
		echo 0 > "${errorfile}"
		echo "Application is stopped" > $statusfile
}

case "$1" in
	update_status)
		update_status
		exit $?
		;;
esac

main "$@"
