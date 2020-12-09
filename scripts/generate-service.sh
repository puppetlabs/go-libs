#!/bin/bash

function validate_choice(){
    choice=$1

   if [ "$choice" != "y" ] && [ "$choice" != "n" ]; then
      echo "Invalid input $choice"
      exit 1
   fi
}

function validate_number(){
    number=$1
    re="^([0-9]+)$"

   if ! [[ $number =~ $re ]] ; then
        echo "Input $number must be number"
        exit 1
   fi
}

#Read input from command line
echo "What is the service name?"
read name < /dev/stdin
BUILDARGS='-ldflags "'"-X main.name=${name}"


echo "Choose directory to generate service code to:(N.B. If it doesn't exist it will be created)"
read servicedir < /dev/stdin
mkdir -p $servicedir
if [ ! -d "$servicedir" ]
then
    echo "Directory $servicedir could not be created."
    exit 1
fi
BUILDARGS="${BUILDARGS} -X main.serviceDir=${servicedir}"

echo "What is the listen address of your service([IP address]:<port>)?"
read serviceaddress < /dev/stdin
re="^([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})?:([0-9]+)$"
if ! [[ $serviceaddress =~ $re ]] ; then
   echo "Listen address must be of the format [IP address]:<port>"
   exit 1
fi
port=${BASH_REMATCH[2]}
BUILDARGS="${BUILDARGS} -X main.listenAddress=${serviceaddress} -X main.listenPort=${port}"

echo "Are you setting up TLS?(y|n)"
read tlssetup < /dev/stdin
validate_choice $tlssetup
BUILDARGS="${BUILDARGS} -X main.tlsSetup=${tlssetup}"

echo "Do you want cors enabled?(y|n)"
read cors < /dev/stdin
validate_choice $cors
BUILDARGS="${BUILDARGS} -X main.corsEnabled=${cors}"

echo "Do you want a readiness check enabled?(y|n)"
read readiness < /dev/stdin
validate_choice $readiness
BUILDARGS="${BUILDARGS} -X main.readinessCheckEnabled=${readiness}"

echo "Do you want to setup rate limiting?(y|n)"
read ratelimiting < /dev/stdin
validate_choice $ratelimiting
if [ "$ratelimiting" = "y" ] ; then
    echo "What would you like the rate interval to be(seconds)?"
    read rateinterval < /dev/stdin
    validate_number $rateinterval
    echo "What would you like the rate limit to be for that interval?"
    read ratelimit < /dev/stdin
    validate_number $ratelimit
    BUILDARGS="${BUILDARGS} -X main.rateLimit=${ratelimit} -X main.rateInterval=${rateinterval}"
fi

echo "Would you like the default prometheus metrics?(y|n)"
read metrics < /dev/stdin
validate_choice $metrics
BUILDARGS="${BUILDARGS} -X main.metricsEnabled=${metrics}"'"'

eval "go run $BUILDARGS cmd/service-generator/main.go"
