# Setup for Linux Debian / Ubuntu

## Install Required Dependencies

```
apt-get update
apt-get install libsodium18 libzmq5 postgresql-9.5-postgis-2.2 redis-server
```

## Setup Database

```
service postgresql start
sudo -u postgres psql postgres
\password postgres
<input new password>
\q
```

## Download and Start Skygear

```
mkdir skygear
cd skygear/
touch .env
# <configurate .env file>
wget https://github.com/$(wget https://github.com/SkygearIO/skygear-server/releases/latest -O - | egrep '/.*/.*/.*linux-amd64.tar.gz' -o)
tar zxf skygear-server-linux-amd64.tar.gz
mv skygear-server-linux-amd64 skygear-server
./skygear-server
```
