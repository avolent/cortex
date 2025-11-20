---
layout: /src/layouts/BaseLayout.astro
author: avolent
title: Setup git with extra tips and tricks
date: May 2024
publish: 'true'
---

## Summary

This page will go over how I setup git and some of the useful tip/tricks I wish I knew before.

## Assumptions

- You use GitHub for your git activities
- You are using Linux/WSL/MacOS

## Contents

1. [Useful Links](#useful-links)
2. [Setting up github SSH Keys](#setting-up-github-ssh-keys)
3. [Commands](#commands)
4. [References](#references)

## Useful Links

- [Git Docs](https://git-scm.com/docs)
- [Github Cheat Sheet](https://training.github.com/downloads/github-git-cheat-sheet/)

## Setting Up Github SSH Keys

### Linux / WSL / MacOS

**Generate a SSH Key[^2] or use a pre-existing one [^3]**  
In the terminal use the following command to generate a key.

```shell
ssh-keygen -t ed25519 -C "your_email@example.com"
```

**Start the ssh-agent in the background.**

```shell
eval "$(ssh-agent -s)"
> Agent pid 59566
```

**Add your SSH private key to the ssh-agent.**

```shell
ssh-add ~/.ssh/key_name
```

**Now add the key to your github account**  
Copy the contents of your key and add to your github account under - [New SSH Key](https://github.com/settings/ssh/new)

```shell
cat ~/.ssh/key_name.pub
```

![add_ssh_key.png](/images/tools/git/add_ssh_key.png)

**You can test your connection with the following:**

```shell
$ ssh -T git@github.com
> Hi USERNAME! You've successfully authenticated, but GitHub does not
> provide shell access.
```

## Commands

Official cheat sheet [^1]

### Setup

```bash
git config --global user.name “name” #Set the name that will be used in commits, Normally your username.
git config --global user.email “[valid-email]” #Set a email that will be used in commits.
git config --global color.ui auto #Add colouring to git commands.
```

### Setup & Init

```bash
git init #Initialise an existing directory as a Git repository.
git clone [url] #Retreive entire repository from a hosted location via URL.
```

### Stage & Snapshot

```bash
git status #Show modified files in working directory, staged for your next commit.
git add [file] #Add a file as it looks not to your next commit (Stage).
git reset [file] #Unstage a file while retaining the changes in working directory.
git diff #Diff of what is changed but not staged.
git diff --staged #Diff of what is staged but not yet commited.
git commit -m "[descriptive message]" #Commit your staged content as a new commit snapshot.
```

### Branch & Merge

### Inspect & Compare

### Tracking Path Changes

### Ignoring Patterns

### Share & Update

### Rewrite History

### Temporary Commits

## References

[^1]: [GitHub Cheat Sheet](https://education.github.com/git-cheat-sheet-education.pdf)

[^2]: [Generating SSH keys for Github](https://docs.github.com/en/authentication/connecting-to-github-with-ssh/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent)

[^3]: [Checking for existing SSH Key](https://docs.github.com/en/authentication/connecting-to-github-with-ssh/checking-for-existing-ssh-keys)
