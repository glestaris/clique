import os
import json
import string
import tempfile
from fabric import api as fab
import ice


@ice.Runner
def deploy_agents(hosts, clique_agent_bin):
    """Orcherstate the deployment of agents.

    :param clique_agent_bin str: The agent Linux binary path.
    """
    if not os.path.isfile(clique_agent_bin):
        print 'ERROR: `{}` is not a path to a file!'.format(clique_agent_bin)

    fab.execute(copy_agent, hosts, clique_agent_bin)
    fab.execute(copy_config, hosts)

    res = fab.execute(run_start, hosts)
    print_outcomes(res)


@ice.Runner
def check_agents(hosts):
    """Checks if the agents are running in the iCE instances.
    """
    res = fab.execute(run_check, hosts)
    print_outcomes(res)


@ice.Runner
def stop_agents(hosts):
    """Stop running agents.
    """
    res = fab.execute(run_stop, hosts)
    print_outcomes(res)


def copy_agent(hosts, clique_agent_bin):
    fab.put(clique_agent_bin, '/home/ec2-user/clique-agent')
    fab.run('chmod +x /home/ec2-user/clique-agent')

    d = os.path.dirname(__file__)
    fab.put(os.path.join(d, 'ctl.sh'), '/home/ec2-user/ctl.sh')
    fab.run('chmod +x /home/ec2-user/ctl.sh')


def copy_config(hosts):
    remote_hosts = []
    for host in hosts.values():
        if host.get_host_string() != fab.env.host_string:
            if len(host.networks) == 0:
                continue
            (ip, net_class) = string.split(host.networks[0]['addr'], '/')
            remote_hosts.append("{}:5000".format(ip))

    cfg = {
        "transfer_port": 5000,
        "remote_hosts": remote_hosts,
        "init_transfer_size": 10*1024*1024  # 10MB
    }

    f = tempfile.NamedTemporaryFile(delete=False)
    f.write(json.dumps(cfg))
    f.close()

    fab.put(f.name, '/home/ec2-user/config.json')
    os.remove(f.name)


def run_start(hosts):
    with fab.cd('/home/ec2-user'):
        return fab.run('./ctl.sh start', warn_only=True, pty=False)


def run_check(hosts):
    with fab.cd('/home/ec2-user'):
        return fab.run('./ctl.sh check', warn_only=True)


def run_stop(hosts):
    with fab.cd('/home/ec2-user'):
        return fab.run('./ctl.sh stop', warn_only=True)


def print_outcomes(res):
    for key, value in res.items():
        outcome = '[OK]'
        if value.failed:
            outcome = '[FAIL]'
        print "{}\t\t{}".format(key, outcome)
