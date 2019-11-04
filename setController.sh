#!/bin/sh

set -e
VERSION=`/appsody-controller --version`
#echo Verifying presence of appsody-controller version $VERSION
set +e
DOWNLOAD=0
if [ -f "/.appsody/appsody-controller" ]; then
    #echo Controller found - checking version
    CURRENT_VERSION=`/.appsody/appsody-controller --version`
    if [ $? -eq 0 ]; then
        #echo Found controller - version: $CURRENT_VERSION
        if [ "$CURRENT_VERSION" != "$VERSION" ]; then
            #echo Current controller version $CURRENT_VERSION does not match required version $VERSION - replacing controller...
            DOWNLOAD=1
        fi
    else
        #echo Old controller failed to execute - replacing it...
        DOWNLOAD=1
    fi
else
    #echo No controller found - copying controller...
    DOWNLOAD=1
fi
if [ $DOWNLOAD = 1 ]; then  
    chmod +x appsody-controller
    mv appsody-controller /.appsody
    #echo Done!
fi
CURRENT_VERSION=`/.appsody/appsody-controller --version`
echo Controller version: $CURRENT_VERSION
