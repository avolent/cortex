---
layout: /src/layouts/BaseLayout.astro
author: avolent
title: Python development setup
date: May 2024
publish: "true"
---

## Summary

The following pages is how I setup my python environments for development. It is something I wish I knew before starting and allows you to get up and running quickly.

## Assumptions

- Basic Python knowledge
- You are using Mac OS (May be similar to other OS)

## Contents

1. [Useful Links](#useful-links)
2. [References](#references)

## Useful Links

- [pyenv]()
- [pipenv]()

## Pyenv

### Installation

```bash
sudo apt-get install -y make build-essential libssl-dev zlib1g-dev \
libbz2-dev libreadline-dev libsqlite3-dev wget curl llvm libncurses5-dev \
libncursesw5-dev xz-utils tk-dev libffi-dev liblzma-dev python-openssl
```

```bash
curl -L https://github.com/pyenv/pyenv-installer/raw/master/bin/pyenv-installer | bash
```

The script will output what to do next. It may be different for your host.

```bash
# Load pyenv automatically by adding
# the following to ~/.bashrc:

export PATH="/home/user/.pyenv/bin:$PATH"
eval "$(pyenv init -)"
eval "$(pyenv virtualenv-init -)"
```

## Pipenv

### Installation

```bash
echo 'export PATH="${HOME}/.local/bin:$PATH"' >> ~/.bashrc
```

```bash
python3 -m pip install --user pipenv
```

## References

<https://www.newline.co/courses/create-a-serverless-slackbot-with-aws-lambda-and-python/installing-python-3-and-pyenv-on-macos-windows-and-linux>  
<https://gist.github.com/planetceres/8adb62494717c71e93c96d8adad26f5c>

[^1]: [Example](https://example.com)
