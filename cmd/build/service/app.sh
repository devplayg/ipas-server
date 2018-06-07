#!/bin/sh

mode=$1
dir=/root/go/src/github.com/devplayg/ipas-server/cmd/build

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
        addToService "generator" ${dir} "-loop -interval 60s"
        ;;

    'start')
        service receiver start
        service classifier start
        service calculator start
        service generator start
        ;;

    'stop')
        service receiver stop
        service classifier stop
        service calculator stop
        service generator stop
        ;;
esac


