#!/bin/sh

#cp calculator.init /etc/init.d/calculator
#cp classifier.init /etc/init.d/classifier
#cp generator.init /etc/init.d/generator
#cp receiver.init /etc/init.d/receiver

addToService() 
{
    command=$1
    dir=$2
    sed 's|###COMMAND###|'"$command"'|g; s|###DIR###|'"$dir"'|g' init.script > /etc/init.d/${command}
    chmod 755 /etc/init.d/${command}
    chkconfig $command on
    service $command restart
}

dir=/root/go/src/github.com/devplayg/ipas-server/cmd/build
addToService receiver ${dir}
addToService classifier ${dir}
addToService calculator ${dir}
addToService generator ${dir}


