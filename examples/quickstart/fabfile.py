import yaml
from StringIO import StringIO
from fabric.operations import local, run, sudo, get, put
from fabric.api import env, settings
from fabric.decorators import task
from fabric.context_managers import cd
from fabric.utils import puts, error


@task
def uptime():
    """
    Show uptime information
    """
    run("uptime")


@task
def reboot():
    """
    Reboot the system
    """
    sudo("shutdown -r now")


def docker_start(service, should_recreate=False):
    with cd("myapp"):
        if should_recreate:
            sudo("docker-compose up -d --force-recreate {0}".format(service))
        else:
            sudo("docker-compose up -d {0}".format(service))


def docker_stop(service, warn_only=True):
    with cd("myapp"):
        with settings(warn_only=warn_only):
            sudo("docker-compose stop {0}".format(service))


def docker_restart(service, should_recreate=False):
    with cd("myapp"):
        if should_recreate:
            docker_start(service, should_recreate=True)
            sudo("docker-compose up -d {0}".format(service))
        else:
            sudo("docker-compose restart {0}".format(service))


def docker_pull(service):
    with cd("myapp"):
        sudo("docker-compose pull {0}".format(service))


def docker_build(service):
    with cd("myapp"):
        sudo("docker-compose build {0}".format(service))


def docker_set_image(service, image):
    with cd("myapp"):
        override = read_compose_override()
        services_dict = override.get('services', {})
        service_dict = services_dict.get(service, {})
        service_dict['image'] = image
        services_dict[service] = service_dict
        override['services'] = services_dict
        write_compose_override(override)


def get_string(path):
    fd = StringIO()
    get(path, fd)
    fd.seek(0)
    return fd.read()


def put_string(data, path):
    fd = StringIO()
    fd.write(data)
    fd.seek(0)
    put(fd, path)


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


def augtool(cmd):
    put_string(cmd, '/tmp/augtool.cmd')
    run("augtool -s -f /tmp/augtool.cmd")
    run('rm /tmp/augtool.cmd')


def plugin_services():
    with cd('myapp'):
        data = read_compose_override()
        if 'services' not in data:
            return []
        return [x for x in data['services'].keys() if x.startswith('plugin_')]


@task
def start_service(service, should_recreate=False):
    docker_start(service, should_recreate)


@task
def stop_service(service):
    docker_stop(service)


@task
def restart_service(service):
    docker_restart(service)


@task
def start_plugin(name, should_recreate=False):
    service = "plugin_{0}".format(name)
    docker_start(service, should_recreate)


@task
def stop_plugin(name):
    service = "plugin_{0}".format(name)
    docker_stop(service)


@task
def restart_plugin(name):
    service = "plugin_{0}".format(name)
    docker_restart(service)


@task
def restart(should_recreate=False):
    """
    Restart Skygear Server and plugins
    """
    services = plugin_services() + ['server']
    if should_recreate:
        docker_start(' '.join(services), should_recreate=should_recreate)
    else:
        docker_stop(' '.join(services))
        docker_start(' '.join(services))


@task
def upgrade(version="latest"):
    """
    Upgrade Skygear Server and plugins
    """
    docker_set_image('server', "skygeario/skygear-server:{0}".format(version))
    docker_pull('server')
    sudo('docker pull skygeario/py-skygear:onbuild')
    services = plugin_services() + ['server']
    docker_build(' '.join(services))
    restart(should_recreate=True)


@task
def logs(service):
    """
    Tail logs of the specified service
    """
    run("docker logs -f --tail=100 myapp_{0}_1".format(service))


@task
def add_upload_key(name, keyfile='~/.ssh/id_rsa.pub'):
    """
    Add a SSH public key to the server for uploading plugin
    """
    put(keyfile, '/tmp/keyfile.pub')
    sudo('gitreceive upload-key "{0}" < /tmp/keyfile.pub'.format(name))


@task
def remove_upload_key(name):
    """
    Remove a SSH public key from the server
    """
    sudo("sed -i '/run {0}/d' /home/git/.ssh/authorized_keys".format(name))


@task
def add_plugin(name, image=None, should_restart=True):
    """
    Add a new plugin by modifying skygear configuration

    If an image is specified, it will be treated as a Docker repository image
    and pulled from the repository. If an image is not specified, a build
    directory is configured where you should upload your plugin via git.

    Skygear Server is restarted automatically by default if an image is
    specified.
    """
    config_file = '/home/ubuntu/myapp/development.ini'
    service_name = "plugin_{0}".format(name)
    with cd("myapp"):
        data = read_compose_override()
        if 'services' not in data:
            data['services'] = {}
        if service_name in data['services']:
            error("Plugin '{0}' already exists.".format(name))
            return
        augtool(r"""
        set /files{0}/plugin\ \"{1}\"/transport http
        set /files{0}/plugin\ \"{1}\"/path http://{1}:8000
        """.format(config_file, service_name))
        service = {
                'restart': 'always',
                'command': 'py-skygear plugin.py --http',
                }
        if image is None:
            service['image'] = service_name
            service['build'] = {
                    'context': '/home/git/.sources/{0}'.format(name),
                }
        else:
            service['image'] = image
        data['services'][service_name] = service
        write_compose_override(data)
    if image is None:
        puts("""Plugin '{0}' is added to Skygear. To upload plugin, add
'git@<ip address>:{0}' as a git remote and push your code.""".format(name))
        return
    if should_restart:
        restart(should_recreate=True)



@task
def remove_plugin(name, should_restart=True):
    """
    Remove an existing plugin by modifying skygear configuration

    Skygear Server is restarted automatically by default.
    """
    config_file = '/home/ubuntu/myapp/development.ini'
    service_name = "plugin_{0}".format(name)
    with cd("myapp"):
        data = read_compose_override()
        if 'services' not in data:
            error("Plugin '{0}' does not exist.".format(name))
            return
    stop_plugin(name)
    with cd("myapp"):
        with settings(warn_only=True):
            sudo("docker-compose rm -f {0}".format(service_name))
        data['services'].pop(service_name, None)
        write_compose_override(data)
        augtool(r'rm /files{0}/plugin\ \"{1}\"'.format(config_file, name))
    if should_restart:
        restart(should_recreate=True)
