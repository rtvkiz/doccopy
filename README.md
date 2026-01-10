# DocCopy

DocCopy is one of the experimental projects that utilize an interesting runc feature. `runc` which is a lower level container runtime. To provide more context on this project and why this is interesting - whenever we run `docker exec` command - the request is passed to `containerd` which then calls runc through shim. 
If we look closely at the runc implementation - https://github.com/opencontainers/runc/blob/main/exec.go#L188, we pass in the flag "CT_ACT_RUN" - basically we are running a new process. 
Now for the trick - `runc` relies on a `state.json` file which helps to identify the container where we want to exec. 
State file has all the information about the root file system, capabilities, seccomp values, namespace symlink path. It is located at the path `/var/run/docker/runtime-runc/moby/<container-id>/state.json`.

https://github.com/opencontainers/runc/blob/main/libcontainer/factory_linux.go#L21


## What we are doing with DocCopy
This is a simple tool which basically creates a clone for your container by using a trick where we pass in and copy the `state.json` file of the given container to a newly started lightweight alpine container with the name `cloned-cont`.
What you will observe on running this script is a new running container which is a copy of your given container and any changes you make will reflect in the existing running container. Any process you start in either of the containers will reflect in the other.


## Prerequisites

- Go 1.23.0 or later
- Docker daemon running
- Root privileges or docker group membership

## Installation

1. Clone the repository:
```bash
git clone https://github.com/rtvkiz/doccopy.git
cd doccopy
```

2. Download dependencies:
```bash
go mod tidy
```

## Usage

1. Build the program:
```bash
go build -o doccopy main.go
```

2. Run with sudo (required for Docker state access):
```bash
sudo ./doccopy [flags] [target-container-id]
```

### Command Line Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-name` | `cloned-cont` | Name for the new container |
| `-image` | `alpine` | Image to use for the new container |
| `-cmd` | `sleep infinity` | Command to run in the container |
| `-it` | `false` | Run container in interactive mode with TTY |
| `-target` | | Target container ID to clone namespaces from |

### Examples

```bash
# Basic usage - prompts for container ID
sudo ./doccopy

# Specify target container as positional argument
sudo ./doccopy e2a04deed1c2

# Create a sidecar container with interactive shell
sudo ./doccopy -name sidecar -cmd "/bin/sh" -it -target e2a04deed1c2

# Use busybox image instead of alpine
sudo ./doccopy -name mycontainer -image busybox -target e2a04deed1c2
```

## What it does

- Inspects the specified Docker container
- Creates a new container sharing PID, network, and IPC namespaces with the target
- The new container can see processes, network interfaces, and IPC resources of the target

## Note

The script will fail if run as a normal user for the sole purpose of reading the 'state.json' file. You might need to run this with elevated privileges.

This won't work on WSL which run docker desktop as that abstracts the implementation.
