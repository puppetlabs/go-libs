#!/bin/bash

function validate_choice(){
    choice=$1

   if [ "$choice" != "y" ] && [ "$choice" != "n" ]; then
      echo "Invalid input $choice"
      exit 1
   fi
}

echo "Choose directory to generate the certificate to:(Defaults to ${PWD})"
read certdir < /dev/stdin
if [ "$certdir" == "" ]; then
    certdir=${PWD}
else
    mkdir -p $certdir
    DIR_ARGS="-directory $certdir"
fi

if [ ! -d "$certdir" ]
then
    echo "Directory $certdir could not be created."
    exit 1
fi

echo "Do you want the CA certificate and key generated?(y|n)"
read generateca
validate_choice $generateca
if [ "$generateca" == "y" ];then
    CA_ARGS="-cafiles true"
fi

echo "Enter the comman name(cn) you want to generate the certificate for(localhost):"
read commonname < /dev/stdin

if [ "$commonname" == "" ]; then
    commonname="localhost"
fi
CN_ARGS="-cn $commonname"

echo "Enter DNS name to generate cert for(localhost):"
while read line
do
    if [ "$line" == "" ]; then
        break
    fi
    host_array=("${host_array[@]}" $line)
    echo "Enter DNS name to generate cert for:(press enter when finished)"
done

if [ ${#host_array[@]} -eq 0 ]; then
    HOST_ARGS="-hostname localhost"
else
    for host in "${host_array[@]}"
    do
        echo $host
        if [ "HOST_ARGS" == "" ]; then
            HOST_ARGS="-hostname $host"
        else
            HOST_ARGS="$HOST_ARGS -hostname $host"
        fi
    done
fi

ARGS="$DIR_ARGS $CA_ARGS $CN_ARGS $HOST_ARGS"
go run cmd/cert-generator/main.go $ARGS
