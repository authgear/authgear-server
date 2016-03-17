# Postgresql README

## Installation

### Mac OS X

The following examples assumed you had Homebrew and (homebrew-services|https://github.com/gapple/homebrew-services) installed.

* `brew install postgres`
* Stop the server
  * `brew services stop postgresql`
* Delete the cluster created by homebrew by default since it might be created with arbitrary locale
  * `rm -rf /usr/local/var/postgres`
* Create a new cluster with correct locale
  * `initdb --pgdata=/usr/local/var/postgres --locale=en_US.UTF-8`
* Now start the postgres server
  * `brew services start postgresql`
* Create a default database to avoid `psql` complaining
  * `createdb`
* Verify database is created correctly w.r.t cluster locale
  * `psql`, then;
  * `\l`
    ```
    ...
       Name    |  Owner   | Encoding |   Collate   |    Ctype    |   Access privileges
     limouren  | limouren | UTF8     | en_US.UTF-8 | en_US.UTF-8 |
    ```
* Now create a new database for Ourd
  * `createdb skygear`
* Done!

### Ubuntu

*Precise*

1. `sudo sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt/ $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list'`
2. `sudo apt-get install wget ca-certificates`
3. `wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo apt-key add -`
4. `sudo apt-get update`
5. `sudo apt-get install postgresql-9.4 pgadmin3`
6. Done!

See (Apt|https://wiki.postgresql.org/wiki/Apt) in the Postgresql's wiki for more information.

*Trusty*

1. `sudo apt-get install postgresql-9.4 pgadmin3`
2. Done!

After postgresql installed:

Let's say we are running skygear server under the user account named `skyuser`

1. `sudo su postgres`
2. `createuser --superuser skyuser`
3. `exit`
4. `sudo su skyuser`
4. `createdb`; stop psql from complaining
5. `createdb skygear`
6. `psql`
7. `\l`; verify that the database is created correctly

```
Name    | Owner    | Encoding |   Collate   |    Ctype    |   Access privileges
--------+----------+----------+-------------+-------------+-----------------------
skygear | skyuser  | UTF8     | en_US.UTF-8 | en_US.UTF-8 |
```

8. Done!

## PostGIS

Ourd uses PostGIS extension to handle geometry stroage and query.

*To install:*

### Mac OS X

```shell
$ brew install postgis
```

### Deb

```shell
$ sudo apt-get postgis-2.1
```

*To Enable:*

```shell
$ psql -c 'CREATE EXTENSION postgis;' -d skygear
```
