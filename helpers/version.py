#!/usr/bin/env python
import os
import re

_VERSION_FILE_REGEX = re.compile(r'CliqueAgentVersion\s*=\s*"([0-9\.]+)"',
                                 re.MULTILINE)
_VERSION_REGEX = re.compile(r'([0-9]+)\.([0-9]+)\.([0-9]+)', re.MULTILINE)


def read_version(file_path, default_version=None):
    if not os.path.isfile(file_path):
        return default_version

    m = re.search(_VERSION_FILE_REGEX, open(file_path, 'r').read())
    if m is None:
        return default_version
    return m.group(1)


def write_version(file_path, version):
    if not os.path.isfile(file_path):
        return False

    old_data = open(file_path, 'r').read()
    data = re.sub(
        _VERSION_FILE_REGEX, r'CliqueAgentVersion = "{}"'.format(version),
        old_data
    )

    f = open(file_path, 'w')
    f.write(data)


class Version(object):
    def __init__(self, major=0, minor=0, patch=0):
        self.major = major
        self.minor = minor
        self.patch = patch

    def __str__(self):
        return '{:d}.{:d}.{:d}'.format(self.major, self.minor, self.patch)

    def bump_major(self):
        self.major += 1

    def bump_minor(self):
        self.minor += 1

    def bump_patch(self):
        self.patch += 1

    @classmethod
    def parse_str(cls, version_str):
        vm = re.match(_VERSION_REGEX, version_str)
        if vm is None:
            return None
        return Version(int(vm.group(1)), int(vm.group(2)), int(vm.group(3)))
