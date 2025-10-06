# DocCopy

A Go program that copies Docker container state between containers.

## Prerequisites

- Go 1.23.0 or later
- Docker daemon running
- Root privileges or docker group membership

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd doccopt
```

2. Download dependencies:
```bash
go mod tidy
```

## Usage

1. Build the program:
```bash
go build -o doccopt main.go
```

2. Run with sudo (required for Docker state access):
```bash
sudo ./doccopt
```

3. Enter a container ID when prompted

## What it does

- Inspects the specified Docker container
- Reads the container's state.json file
- Creates a new Alpine container named "cloned-cont"
- Copies the state from the original container to the new one
- Starts the cloned container

## Note

The script will fail if run as a normal user for the sole purpose of reading the 'state.json' file. You might need to run this with elevated privileges.

This won't work on WSL which run docker desktop as that abstracts the implementation.
