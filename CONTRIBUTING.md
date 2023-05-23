## Local Development

### First Time Set Up

1. Download Taskfile by following the instructions at: https://taskfile.dev/installation/

2. Run the following command from the root of the repo to build the CLI, this will create an executable file called `amp` in the bin folder.

```
task build
```

Create a symlink in `/usr/local/bin` that points to this executable. Run this command from the root of the repo. This command names the symlinked executable `lamp`, which stands for `local amp`, in order the differentiate between the local and production versions of the amp CLI, so that you can have both running on your computer. But you can feel free to name it anything you want.

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

### Steps for setup on Windows Operating System.

1. Install task on windows. Documentation:  https://taskfile.dev/installation/

2. Run the following command from the root of the repo to build the CLI, this will create an executable file called `amp` in the bin folder.
    `task build`

3. Open command prompt as administrator. You can choose a directory such as C:\Windows or C:\Windows\System32 for system-wide availability, or you can create a new directory and add it to the "Path" variable.

4. Run the following commands.
    `echo @echo off > lamp.bat`
    `echo "C:\Users\YourUsername\Documents\cli\bin\amp.exe" %* >> lamp.bat`

    Replace "C:\Users\YourUsername\Documents\cli\bin\amp.exe" with the actual path to the `amp.exe` file. Make sure to keep the double quotes around the file path if it contains spaces.

5. You can now run the command `lamp` from anywhere in the command line, and it will execute the `amp.exe` file.


### Ongoing Development

Whenever you make code changes, and want to test it locally, run 

```
task build
```

Now you can run `lamp` commands from anywhere on your computer and it'll use your latest code!
