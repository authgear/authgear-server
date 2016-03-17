#!/bin/sh

set -e

: ${DOCKER_COMPOSE_VERSION:=1.6.2}

DOCKER_ENGINE_PACKAGE=docker-engine=1.10.2-0~precise

# Update apt packages
echo "deb https://apt.dockerproject.org/repo ubuntu-precise main" > /etc/apt/sources.list.d/docker.list
apt-key adv --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys 58118E89F3A912897C070ADBF76221572C52609D
apt-get update
apt-get upgrade -y

# Reconfigure locales to get rid of warning when log in
locale-gen en_US.UTF-8
dpkg-reconfigure locales

# Install linux-image-extra so that Docker will use aufs
apt-get -y install linux-image-extra-$(uname -r)

# Install Docker Engine
apt-get install -y --no-install-recommends $DOCKER_ENGINE_PACKAGE
usermod -a -G docker ubuntu

# Install Docker Compose
curl -L https://github.com/docker/compose/releases/download/$DOCKER_COMPOSE_VERSION/docker-compose-`uname -s`-`uname -m` > /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Install CloudFormation Helper Scripts
apt-get install -y python-pip
pip install pyopenssl ndg-httpsclient pyasn1
pip install https://s3.amazonaws.com/cloudformation-examples/aws-cfn-bootstrap-latest.tar.gz

# Install jinja2-cli to modify config files
pip install jinja2-cli

# Install gitreceive
apt-get install -y git
wget https://raw.github.com/progrium/gitreceive/master/gitreceive
mv gitreceive /usr/local/bin/gitreceive
chmod +x /usr/local/bin/gitreceive
gitreceive init
usermod -aG docker git

# Install augtool for modifying skygear server config
apt-get install -y augeas-tools

# Move other files into place
cp /tmp/files/kickstart.sh /usr/local/bin/kickstart.sh
chmod +x /usr/local/bin/kickstart.sh
cp /tmp/files/motd /etc/motd
cp /tmp/files/skygear.aug /usr/share/augeas/lenses
cp /tmp/files/receiver /home/git/receiver
chmod +x /home/git/receiver
rm -rf /tmp/files
