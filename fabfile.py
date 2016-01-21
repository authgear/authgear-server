# fabric 1.9.0
from fabric.operations import local
from fabric.api import env


'''
This file is collection of commands regarding deployment
'''

env.roledefs.update({
    'hub-migrate': ['hub.docker.com'],
})

# Heaven will execute fab -R hub-migrate deploy:branch_name=master
def deploy(branch_name):
    print("Executing on %(host)s as %(user)s" % env)
    if env.host == 'hub.docker.com':
        build_migrate()
    else:
        raise ValueError('Not supported deployment target')


def build_migrate():
    local('docker build -f Dockerfile-migrate -t oursky/skygear-migrate .')
    local('docker push oursky/skygear-migrate')
