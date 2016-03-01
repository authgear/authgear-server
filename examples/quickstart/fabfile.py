import yaml
from StringIO import StringIO
from fabric.operations import local, run, sudo, get, put
from fabric.api import env, settings
from fabric.decorators import task
from fabric.context_managers import cd
from fabric.utils import puts


@task
def uptime():
    run("uptime")


@task
def reboot():
    sudo("shutdown -r now")


def docker_start(service):
    with cd("myapp"):
        sudo("docker-compose up -d {0}", service)


def docker_stop(service):
    with cd("myapp"):
        sudo("docker-compose stop {0}", service)


def docker_restart(service):
    with cd("myapp"):
        sudo("docker-compose restart {0}", service)


def docker_recreate(service):
    with cd("myapp"):
        sudo("docker-compose up -d --force-recreate {0}", service)


def docker_pull(service):
    with cd("myapp"):
        sudo("docker-compose pull {0}", service)


def docker_set_image(service, image):
    with cd("myapp"):
        override = read_compose_override()
        services_dict = override.get('services', {})
        service_dict = services_dict.get(service, {})
        service_dict['image'] = image
        services_dict[service] = service_dict
        override['services'] = services_dict
        write_compose_override(override)


def read_compose_override():
    with settings(abort_exception=Exception):
        try:
            fd = StringIO()
            get('docker-compose.override.yml', fd)
            fd.seek(0)
            return yaml.load(fd.read()) or {}
        except Exception:
            return {}


def write_compose_override(data):
    data['version'] = '2'
    fd = StringIO()
    fd.write(yaml.dump(data, default_flow_style=False))
    fd.seek(0)
    put(fd, 'docker-compose.override.yml')


@task
def recreate(service):
    docker_recreate(service)


@task
def start(service):
    docker_start(service)


@task
def stop(service):
    docker_stop(service)


@task
def restart(service):
    docker_restart(service)


@task
def upgrade(version):
    docker_set_image('server', "skygeario/skygear:{0}".format(version))
    docker_pull('server')
    docker_recreate('server')
