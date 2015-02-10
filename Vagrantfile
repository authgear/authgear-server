# -*- mode: ruby -*-
# vi: set ft=ruby :

$script = <<EOF
apt-get install -y git
curl -s https://storage.googleapis.com/golang/go1.4.1.linux-amd64.tar.gz | tar -C /usr/local -zxf -
ln -s /usr/local/go/bin/go /usr/local/bin/go

echo "\nexport GOPATH=/vagrant" >> /home/vagrant/.bashrc
cd /vagrant
GOPATH=/vagrant go get github.com/tools/godep

EOF

Vagrant.configure(2) do |config|
  config.vm.box = "trusty64"
  config.vm.provision "shell", inline: $script
  config.vm.synced_folder ".", "/vagrant/src/github.com/oursky/ourd"
end
