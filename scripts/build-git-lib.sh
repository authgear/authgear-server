#!/bin/sh

if [  $# -lt 1 ]; then 
    echo "\nUsage: $0 <giturl> [<refspec>]\n"
    exit 1
fi 

REPO_NAME=`basename "${1%.*}"`

if [ ! -z $LD_INSTALL_PREFIX ]; then
    CONFIGURE_ARGS=--prefix=$LD_INSTALL_PREFIX
else
    CONFIGURE_ARGS=
fi

if [ ! -z $2 ]; then
    REFSPEC=$2
else
    REFSPEC=origin/master
fi

if [ ! -d "$REPO_NAME" ]; then
    echo "Repository $REPO_NAME not cloned. Cloning..."
    git clone $1
    cd $REPO_NAME
    git checkout $REFSPEC
    ./autogen.sh
    ./configure $CONFIGURE_ARGS
    make check
    make install
else
    echo "Found repository $REPO_NAME."
    cd $REPO_NAME
    git fetch
    if [ `git rev-parse HEAD` != `git rev-parse $REFSPEC` ]; then
        echo "Local hash `git rev-parse HEAD` and Remote hash `git rev-parse $REFSPEC` differ. Rebuilding..."
        git clean -fxd
        git checkout $REFSPEC
        ./autogen.sh
        ./configure $CONFIGURE_ARGS
        make check
        make install
    fi
fi

cd ..
