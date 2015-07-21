# fabric 1.9.0
from fabric.operations import run
from fabric.api import env


'''
This file is collection of commands regarding deployment
'''

env.user = 'ourd'
env.roledefs.update({
    'pandawork': ['ourd.pandawork.com'], #54.159.147.211
})

print(env.host)

# Heaven will execute fab -R edge deploy:branch_name=edge
def deploy(branch_name):
    print("Executing on %s as %s" % (env.host, env.user))
    run('docker pull oursky/ourd:%s' % branch_name)
    run('docker stop ourd')
    run('docker rm ourd')
    init_ourd()


def restart():
    run('docker restart ourd')


### Below is init image as refs. Normal flow will only need above command.
### Assume the machine with docker installed,
### ubuntu: wget -qO- https://get.docker.com/ | sh
### adduser ourd
### passwd -l ourd
### adduser ourd docker

def init_ourd():
    '''
    Run ourd containter
    '''
    run('''docker run --name ourd \
-v /home/ourd:/etc/ourd:ro \
-v /home/ourd/data:/etc/data:rw \
--expose 3000 \
-p 3000:3000 \
-e "OD_CONFIG=/etc/ourd/docker.ini" \
--link ourpg:postgres \
--restart=always \
-d oursky/ourd''')


def start_db():
    run('''docker run --name ourpg \
-v /home/ourd/pgdata:/var/lib/postgresql/data \
-p 5432:5432 \
-d postgres''')


def start_nginx():
    sudo('''docker run -d --name nginx --restart=always \
-v /var/lib/docker/data/nginx/etc:/etc/nginx \
-v /var/lib/docker/data/nginx/log:/var/log/nginx \
-p 80:80 -p 443:443 \
nginx:1.9.2''')
