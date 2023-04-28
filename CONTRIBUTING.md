## Local Development

### First Time Set Up

1. Download Taskfile by following the instructions at: https://taskfile.dev/installation/

2. Run the following command from the root of the repo to build the CLI, this will create an executable file called `amp` in the bin folder.

```
task build
```

Create a symlink in `/usr/local/bin` that points to this executable. Run this command from the root of the repo. This command names the symlinked executable `lamp`, which stands for `local amp`, in order the differentiate between the production version of the amp CLI, so that you can have both running on your computer. But you can feel free to name it anything you want.

```
sudo ln -s $PWD/bin/amp /usr/local/bin/lamp
```

You can test if this worked by running the following:
```
which lamp
```
It should print `/usr/local/bin/lamp`

Now you can run the following from anywhere on your computer:
```
lamp
```
This should print a list of available commands in the CLI.

Now you can run `lamp` commands from anywhere on your computer, and it'll use your local version!

If this doesn't work, then you'll need to add `/usr/local/bin` to your PATH variable.

### Ongoing Development

Whenever you make code changes, and want to test it locally, run 

```
task build
```

Now you can run `lamp` commands from anywhere on your computer and it'll use your latest code!
