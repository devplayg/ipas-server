#!/bin/sh

dir=/root/go/src/github.com/devplayg/ipas-server/cmd/build


addToService()
{
    command=$1
    dir=$2
    sed 's|###COMMAND###|'"$command"'|g; s|###DIR###|'"$dir"'|g' init.script > /etc/init.d/${command}
    chmod 755 /etc/init.d/${command}
    chkconfig $command on
    service $command restart
}

start()
{
    service start receiver
    service start classifier
    service start calculator
    service start generator
}

stop()
{
    service stop receiver
    service stop classifier
    service stop calculator
    service stop generator
}




addToService receiver ${dir}
addToService classifier ${dir}
addToService calculator ${dir}
addToService generator ${dir}