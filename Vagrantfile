# -*- mode: ruby -*-
# vi: set ft=ruby :

$script = <<EOF

echo "Installing go..."

curl -s https://storage.googleapis.com/golang/go1.4.linux-amd64.tar.gz | sudo tar -C /usr/local -zxf -
sudo ln -s /usr/local/go/bin/go /usr/local/bin/go

echo "Installing postgres..."

sudo apt-get install -y postgresql augeas-tools
sudo augtool <<-AUGEAS
set /files/etc/postgresql/9.3/main/pg_hba.conf/*[address="127.0.0.1/32"][last()]/method trust
set /files/etc/postgresql/9.3/main/pg_hba.conf/*[address="::1/128"][last()]/method trust
save
quit
AUGEAS
sudo -u postgres createdb ourd_test
sudo -u postgres createuser vagrant
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE ourd_test to vagrant;"
sudo service postgresql restart

echo "Installing git..."

sudo apt-get install -y git

# Fix ownership on /vagrant so root permission is not needed to install
# go dependencies.
sudo find /vagrant -xdev -exec chown vagrant:vagrant {} \\;

# Set environment variables.
export GOPATH=/vagrant
export PATH=/vagrant/bin:$PATH
echo "\nexport GOPATH=/vagrant" >> /home/vagrant/.bashrc
echo "\nexport PATH=/vagrant/bin:\\$PATH" >> /home/vagrant/.bashrc

echo "Installing go dependencies. This can take a while..."

cd /vagrant
go get github.com/tools/godep
go get github.com/smartystreets/goconvey/convey
go get github.com/pilu/fresh
godep go install github.com/oursky/ourd/...

echo "   done."

echo "Performing tests..."

godep go test github.com/oursky/ourd/...

EOF

Vagrant.configure(2) do |config|
  config.vm.box = "ubuntu/trusty64"
  config.vm.provision "shell", inline: $script, privileged: false
  config.vm.synced_folder ".", "/vagrant/src/github.com/oursky/ourd"
  config.vm.synced_folder "Godeps", "/vagrant/Godeps"
  config.vm.network "forwarded_port", guest: 5432, host: 5432
  config.vm.network "forwarded_port", guest: 3000, host: 3000
end
