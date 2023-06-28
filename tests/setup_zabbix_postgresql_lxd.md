# Set up Zabbix with PostgreSQL in a LXD Ubuntu 22.04 LTS container

* Install steps references
     - LXD: [Getting started - LXD documentation](https://linuxcontainers.org/lxd/docs/latest/getting_started/)
         (Japanese: [LXD を使い始めるには - LXD ドキュメント](https://lxd-ja.readthedocs.io/ja/latest/getting_started/))
     - PostgreSQL: https://wiki.postgresql.org/wiki/Apt
     - Zabbix: https://www.zabbix.com/download?zabbix=6.0&os_distribution=ubuntu&os_version=22.04&components=server_frontend_agent&db=pgsql&ws=nginx

```bash
lxc launch ubuntu:22.04 zabbix
```

```bash
cat <<'EOF' > setup_zabbix.sh
#!/bin/bash

DB_PASSWORD=zbxpass

apt-get -y install curl ca-certificates
curl -sSL -o /etc/apt/keyrings/pgdg.asc 'https://keyserver.ubuntu.com/pks/lookup?op=get&search=0x7FCC7D46ACCC4CF8'
echo "deb [signed-by=/etc/apt/keyrings/pgdg.asc] http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list

apt-get update
apt-get -y install postgresql-15

curl -sSLO https://repo.zabbix.com/zabbix/6.0/ubuntu/pool/main/z/zabbix-release/zabbix-release_6.0-4+ubuntu22.04_all.deb
dpkg -i zabbix-release_6.0-4+ubuntu22.04_all.deb
apt-get update 

apt-get -y install zabbix-server-pgsql zabbix-frontend-php php8.1-pgsql zabbix-nginx-conf zabbix-sql-scripts zabbix-agent

sudo -iu postgres psql -c "CREATE USER zabbix PASSWORD '$DB_PASSWORD';"
sudo -iu postgres createdb -O zabbix zabbix 

echo "*:*:*:zabbix:$DB_PASSWORD" > /var/lib/postgresql/.pgpass
chown postgres: /var/lib/postgresql/.pgpass
chmod 600 /var/lib/postgresql/.pgpass

zcat /usr/share/zabbix-sql-scripts/postgresql/server.sql.gz | sudo -iu postgres psql -h localhost -U zabbix

cp -p /etc/zabbix/zabbix_server.conf /etc/zabbix/zabbix_server.conf.orig
sed '/^# DBPassword=/a\
\
DBPassword='"$DB_PASSWORD" /etc/zabbix/zabbix_server.conf.orig > /etc/zabbix/zabbix_server.conf

rm /etc/nginx/sites-enabled/default

cp -p /etc/zabbix/nginx.conf /etc/zabbix/nginx.conf.orig

sed 's/^# *listen  *8080;/        listen 80;/;s/^# *server_name  *example.com;/        server_name example.com;/' /etc/zabbix/nginx.conf.orig > /etc/zabbix/nginx.conf

systemctl stop apache2
systemctl mask apache2

systemctl restart zabbix-server zabbix-agent nginx php8.1-fpm
systemctl enable zabbix-server zabbix-agent nginx php8.1-fpm
EOF
```

```bash
chmod +x setup_zabbix.sh
lxc file push setup_zabbix.sh zabbix/usr/local/sbin/
lxc exec zabbix /usr/local/sbin/setup_zabbix.sh
```

```bash
ipv4addr=$(lxc exec zabbix -- ip -4 -br addr show dev eth0 | awk '{sub(/\/.*/, "", $3); print $3}')
echo open http://$ipv4addr and complete web interface installation
```

https://www.zabbix.com/documentation/current/en/manual/installation/frontend

* Configure DB connection
    - Password にsetup_zabbix.sh内のDB_PASSWORDと同じ値を入力
    - Database TLS encryptionのチェックを外す

* Settings
    - Zabbix server name にsetup_zabbix.sh内で/etc/zabbix/nginx.conf内の
      server_nameに指定したホスト名(上記の例ではexample.com)を入力
    - Default time zone を (UTC+09:00) Asia/Tokyo に設定

https://www.zabbix.com/documentation/current/en/manual/quickstart/login#:~:text=This%20is%20the%20Zabbix%20welcome,in%20as%20a%20Zabbix%20superuser.

* Administrator username: Admin
* Initial password:       zabbix

## Add group hosts and hosts

lxc project create zabbix --config features.images=false --config features.profiles=false
lxc stop zabbix
lxc move zabbix zabbix --project default --target-project zabbix

lxc project switch zabbix

CLIENTS=$(echo sv0{1,2}-grp{1,2} | tr ' ' '\n')

echo "$CLIENTS" | xargs -I % -P 0 lxc launch ubuntu:22.04 %

ZBX_SERVER_IP=$(lxc info zabbix | sed -n '/^ *eth0:$/,/^ *inet:/{/^ *inet:/{s/^ *inet: *\([^/]*\).*/\1/;p}}')
cat <<EOF > zabbix_server.conf
Server=$ZBX_SERVER_IP
ServerActive=$ZBX_SERVER_IP
EOF

cat <<EOF > setup_zabbix_client.sh
#!/bin/bash
curl -sSLO https://repo.zabbix.com/zabbix/6.0/ubuntu/pool/main/z/zabbix-release/zabbix-release_6.0-4+ubuntu22.04_all.deb
dpkg -i zabbix-release_6.0-4+ubuntu22.04_all.deb
apt-get update
rm zabbix-release_6.0-4+ubuntu22.04_all.deb
apt-get -y install zabbix-agent
EOF

chmod +x setup_zabbix_client.sh

lxc file push setup_zabbix_client.sh sv01-grp1/usr/local/sbin/
lxc exec sv01-grp1 -- /usr/local/sbin/setup_zabbix_client.sh

lxc file push zabbix_server.conf sv01-grp1//etc/zabbix/zabbix_agentd.d/

echo "$CLIENTS" | xargs -I % -P 0 lxc file push setup_zabbix_client.sh %/usr/local/sbin/
echo "$CLIENTS" | xargs -I % -P 0 lxc exec % -- /usr/local/sbin/setup_zabbix_client.sh

echo "$CLIENTS" | xargs -I % -P 0 lxc file push zabbix_server.conf %/etc/zabbix/zabbix_agentd.d/
echo "$CLIENTS" | xargs -I % -P 0 lxc exec % -- systemctl restart zabbix-agent
