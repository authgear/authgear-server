#Setup for Linux Debian / Ubuntu

##Install Required Dependency

```
apt-get install postgresql
apt-get install golang
apt-get install postgis
apt-get install postgresql-client
apt-get install postgresql-contrib
apt-get install software-properties-common
apt-get install redis-server
apt-get install libsodium-dev
apt-get install libghc-zeromq4-haskell-dev
```

##Setup Database
```
service postgresql start
sudo -u postgres psql postgres
\password postgres
<input new password>
\q
```

##Download and Start Skygear
```
mkdir skygear
cd skygear/
touch .env
<configurate .env file>
wget https://github.com/SkygearIO/skygear-server/releases/download/v0.22.1/skygear-server-linux-amd64
chmod +x skygear-server-linux-amd64 
./skygear-server-linux-amd64 
```
