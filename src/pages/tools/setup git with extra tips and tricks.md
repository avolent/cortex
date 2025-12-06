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

![add_ssh_key.png](/images/tools/setup%20git%20with%20extra%20tips%20and%20tricks/add_ssh_key.png)

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

```bash
git branch #List your branches, a * will appear next to the currently active branch.
git branch [branch-name] #Create a new branch at the current commit.
git checkout [branch-name] #Switch to another branch and check it out into your working directory.
git checkout -b [branch-name] #Create a new branch and switch to it in one command.
git merge [branch] #Merge the specified branch's history into the current one.
git log #Show all commits in the current branch's history.
git branch -d [branch-name] #Delete a local branch (use -D to force delete).
git push origin --delete [branch-name] #Delete a remote branch.
```

### Inspect & Compare

```bash
git log #Show the commit history for the currently active branch.
git log --oneline #Show commit history in condensed format.
git log --graph --oneline --all #Visual representation of branch structure.
git log --follow [file] #Show the commits that changed file, even across renames.
git diff [branch-a]...[branch-b] #Show the diff of what is in branch-a that is not in branch-b.
git show [commit] #Show any object in Git in human-readable format.
git log --author="[name]" #Show commits by a specific author.
git log --since="2 weeks ago" #Show commits from a specific time period.
git blame [file] #Show who changed what and when in a file.
```

### Tracking Path Changes

```bash
git rm [file] #Delete the file from project and stage the removal for commit.
git mv [existing-path] [new-path] #Change an existing file path and stage the move.
git log --stat -M #Show all commit logs with indication of any paths that moved.
git rm --cached [file] #Remove file from version control but preserve locally.
```

### Ignoring Patterns

```bash
# Create a .gitignore file to prevent staging unwanted files
git config --global core.excludesfile [file] #System wide ignore pattern for all local repositories.
```

**Common .gitignore patterns:**

```gitignore
# OS files
.DS_Store
Thumbs.db

# IDE files
.vscode/
.idea/
*.swp

# Dependencies
node_modules/
vendor/

# Build outputs
dist/
build/
*.log

# Environment files
.env
.env.local
```

### Share & Update

```bash
git remote add [alias] [url] #Add a git URL as an alias.
git fetch [alias] #Fetch down all the branches from that Git remote.
git merge [alias]/[branch] #Merge a remote branch into your current branch.
git push [alias] [branch] #Push local branch commits to remote repository.
git pull #Fetch and merge any commits from the tracking remote branch.
git push -u origin [branch] #Push branch to remote and set upstream tracking.
git remote -v #List all configured remotes.
git remote show origin #Show information about the remote.
```

### Rewrite History

```bash
git rebase [branch] #Apply any commits of current branch ahead of specified one.
git reset --hard [commit] #Clear staging area, rewrite working tree from specified commit.
git reset --soft HEAD~ #Undo last commit but keep changes staged.
git reset --mixed HEAD~ #Undo last commit and unstage changes (default).
git commit --amend #Modify the last commit (message or content).
git rebase -i HEAD~[n] #Interactive rebase for last n commits.
git reflog #Show a log of changes to HEAD (useful for recovering lost commits).
```

**Warning:** Never rewrite history on shared/public branches!

### Temporary Commits

```bash
git stash #Save modified and staged changes for later.
git stash list #List stack-order of stashed file changes.
git stash pop #Write working from top of stash stack and remove it.
git stash apply #Apply stashed changes without removing from stash.
git stash drop #Discard the changes from top of stash stack.
git stash save "message" #Stash with a descriptive message.
git stash show #Show the changes in the latest stash.
git stash branch [branch-name] #Create a new branch from a stash.
```

## Advanced Tips & Tricks

### Git Aliases

Save time with custom shortcuts:

```bash
git config --global alias.co checkout
git config --global alias.br branch
git config --global alias.ci commit
git config --global alias.st status
git config --global alias.unstage 'reset HEAD --'
git config --global alias.last 'log -1 HEAD'
git config --global alias.visual 'log --graph --oneline --all'
git config --global alias.amend 'commit --amend --no-edit'
```

### Useful Workflows

**Undo a commit that was pushed:**

```bash
git revert [commit-hash] #Creates a new commit that undoes changes
```

**Cherry-pick a commit from another branch:**

```bash
git cherry-pick [commit-hash] #Apply specific commit to current branch
```

**Find which commit introduced a bug (binary search):**

```bash
git bisect start
git bisect bad #Mark current commit as bad
git bisect good [commit-hash] #Mark a known good commit
# Git will checkout commits for you to test, mark each as good/bad
git bisect reset #When done
```

**Clean up untracked files:**

```bash
git clean -n #Dry run - show what would be deleted
git clean -fd #Force delete untracked files and directories
```

**Work with patches:**

```bash
git format-patch [branch] #Create patch files for commits
git apply [patch-file] #Apply a patch file
git am [patch-file] #Apply a patch and create a commit
```

### Performance & Maintenance

```bash
git gc #Cleanup unnecessary files and optimize local repository
git prune #Remove unreachable objects
git fsck #Verify the connectivity and validity of objects
git count-objects -vH #Show repository size statistics
```

### Searching & Finding

```bash
git grep [pattern] #Search working directory for pattern
git log -S [string] #Search commits that added or removed a string
git log -G [regex] #Search commits with changes matching regex
git show [commit]:[file] #Show file contents at specific commit
```

### Configuration Tips

**View all settings:**

```bash
git config --list #Show all git configuration
git config --list --show-origin #Show settings and their source files
```

**Helpful global configurations:**

```bash
git config --global pull.rebase true #Use rebase instead of merge for pulls
git config --global core.editor "vim" #Set your preferred editor
git config --global init.defaultBranch main #Set default branch name
git config --global rerere.enabled true #Remember how you resolved merge conflicts
git config --global core.autocrlf input #Handle line endings (input for Mac/Linux)
```

### Debugging & Troubleshooting

**Check what changed in a file:**

```bash
git diff HEAD~2..HEAD [file] #Compare file between two commits ago and now
git log -p [file] #Show patch/diff for each commit affecting file
```

**Find who deleted a file:**

```bash
git log --all --full-history -- [file-path]
```

**Recover deleted branch:**

```bash
git reflog #Find the commit where branch was
git checkout -b [branch-name] [commit-hash]
```

**Unstage all files:**

```bash
git reset
```

**Discard all local changes:**

```bash
git restore . #Git 2.23+
# or
git checkout -- . #Older Git versions
```

### Working with Submodules

```bash
git submodule add [url] [path] #Add a submodule
git submodule update --init --recursive #Initialize and update submodules
git submodule update --remote #Update submodules to latest remote commits
```

### Best Practices

1. **Commit often, perfect later** - Make small, logical commits
2. **Write meaningful commit messages** - Use imperative mood ("Add feature" not "Added feature")
3. **Pull before you push** - Always sync with remote before pushing
4. **Use branches** - Keep main/master stable, develop in branches
5. **Review before committing** - Use `git diff --staged` before commit
6. **Don't commit sensitive data** - Use .gitignore and environment variables
7. **Keep commits atomic** - One logical change per commit
8. **Test before pushing** - Ensure code works before sharing

### Commit Message Convention

```text
<type>: <subject>

<body>

<footer>
```

Types: feat, fix, docs, style, refactor, test, chore

Example:

```text
feat: add user authentication

Implement JWT-based authentication system with login and logout endpoints.

Closes #123
```

## References

[^1]: [GitHub Cheat Sheet](https://education.github.com/git-cheat-sheet-education.pdf)

[^2]: [Generating SSH keys for Github](https://docs.github.com/en/authentication/connecting-to-github-with-ssh/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent)

[^3]: [Checking for existing SSH Key](https://docs.github.com/en/authentication/connecting-to-github-with-ssh/checking-for-existing-ssh-keys)
