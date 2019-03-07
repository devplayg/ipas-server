#!/bin/sh

mode=$1
dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

addToService()
{
    command=$1
    dir=$2
    opt=$3
    sed 's|###COMMAND###|'"$command"'|g; s|###DIR###|'"$dir"'|g; s|###OPT###|'"$opt"'|g;' init.script > /etc/init.d/${command}
    chmod 755 /etc/init.d/${command}
    chkconfig $command on
}


case "$mode" in
    'install')
        addToService receiver ${dir}
        addToService classifier ${dir}
        addToService calculator ${dir}
        addToService scheduler ${dir}
        # addToService "generator" ${dir} "-loop -interval 60s"
        ;;

    'start')
        service receiver start
        service classifier start
        service calculator start
        service scheduler start
        # service generator start
        ;;

    'stop')
        service receiver stop
        service classifier stop
        service calculator stop
        service scheduler stop
        # service generator stop
        ;;
esac


