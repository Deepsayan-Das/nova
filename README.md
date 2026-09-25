# Nova — GalactOS Command-Line Interface

`nova` is the primary command-line tool for **GalactOS**, a custom Linux operating system distribution. Built in Go using the Cobra framework, `nova` provides developer utilities, container management powered by Podman, Kubernetes manifest generation, system diagnostics, project templating, and automated tool installation.

---

## Key Features

- 🐳 **Container Execution & Management (Podman-native)**
  - Interactive multi-stage Dockerfile generator (`nova container dockerfile`).
  - Interactive `docker-compose.yml` generator (`nova container compose`).
  - Live container image builder with real-time log streaming (`nova container build`).
  - Container runner with foreground/detached support (`nova container run`).
  - Container lifecycle management (`nova container ps`, `nova container stop`, `nova container logs`).

- ☸️ **Kubernetes Cluster Manifests**
  - Interactive Kubernetes Deployment & Service generator writing clean YAML manifests to `./k8s/` (`nova cluster manifest`).

- 🩺 **Project Doctor & Environment Health**
  - Scans project requirements (Node.js, Python, Go, Rust, Java), verifies installed runtime versions against `.nova.yaml`, and recommends version fix commands (`nova project doctor`).

- 🛠️ **System Utilities & Package Installer**
  - Automated package installer (`nova install <tool>`).
  - System diagnostics and hardware benchmarking (`nova system`).
  - Project starter initialization (`nova init`).

---

## Installation & Setup

### Option 1: Install via APT Repository (Debian / Ubuntu / GalactOS)

Import the GPG key and add the official APT repository:

```bash
# 1. Import GPG Key
curl -fsSL https://deepsayan-das.github.io/nova/KEY.gpg | sudo gpg --dearmor -o /etc/apt/keyrings/galactos.gpg

# 2. Add APT Repository
echo "deb [signed-by=/etc/apt/keyrings/galactos.gpg] https://deepsayan-das.github.io/nova stable main" | sudo tee /etc/apt/sources.list.d/galactos.list

# 3. Update & Install
sudo apt update
sudo apt install nova
```

### Option 2: Build Debian Package (.deb) Locally

To build a `.deb` package locally using the automated build script:

```bash
chmod +x build_deb.sh
./build_deb.sh 1.0.0 1
sudo dpkg -i nova_1.0.0-1_amd64.deb
```

### Option 3: Building from Source

Clone the repository and compile using Go:

```bash
git clone https://github.com/Deepsayan-Das/nova.git
cd nova
go build -o nova main.go
```

To install globally on your system:

```bash
go install
```

---

## Command Reference

### 1. Container Commands (`nova container`)

`nova container` exposes subcommands for container lifecycle management and manifest generation:

#### `nova container dockerfile`
Interactively generates a single-stage or multi-stage `Dockerfile`.
```bash
nova container dockerfile
```
- **Options**: Select base image (`node`, `python`, `go`, `custom`), version tag, port, working directory, entry command, and multi-stage build flag.
- **Output**: Generates `./Dockerfile`.

#### `nova container compose`
Interactively configures services, ports, environment variables, and `depends_on` relationships.
```bash
nova container compose
```
- **Output**: Generates `./docker-compose.yml`.

#### `nova container build`
Builds a container image using Podman with live output streaming.
```bash
# Interactive tag prompt using default Dockerfile
nova container build

# Specify custom tag and Dockerfile
nova container build -t my-app:latest -f Dockerfile.prod .
```

#### `nova container run`
Runs a container from an image via Podman.
```bash
# Foreground execution with port mapping
nova container run node:20-alpine -p 3000:3000

# Detached background execution with custom name and environment variables
nova container run node:20-alpine -d --name web-api -e NODE_ENV=production -p 8080:8080
```

#### `nova container ps`
Lists running or all containers via Podman.
```bash
# List running containers
nova container ps

# List all containers (including stopped)
nova container ps -a
```

#### `nova container stop`
Stops a running container.
```bash
nova container stop web-api
```

#### `nova container logs`
Fetches logs from a container.
```bash
# Print recent logs
nova container logs web-api

# Stream/follow logs live
nova container logs web-api -f
```

---

### 2. Kubernetes Cluster Commands (`nova cluster`)

#### `nova cluster manifest`
Interactively generates production-ready Kubernetes Deployment and Service YAML manifests.
```bash
nova cluster manifest
```
- **Prompts**: Application name, container image, replica count, container port, and service type (`ClusterIP`, `NodePort`, `LoadBalancer`).
- **Output**: Generates `./k8s/deployment.yaml` and `./k8s/service.yaml`.

---

### 3. Development & Health Commands

#### `nova project doctor`
Inspects project runtime environments and compares active versions against `.nova.yaml` metadata.
```bash
nova project doctor
```

#### `nova install`
Installs system dependencies and toolchain packages.
```bash
# List available tools in registry
nova install

# Install a specific tool (e.g. podman, go, node)
nova install podman
```

#### `nova system`
Displays GalactOS system information, memory usage, CPU stats, and runs quick diagnostic checks.
```bash
nova system
```

#### `nova init`
Initializes a new GalactOS project workspace with `.nova.yaml` metadata.
```bash
nova init
```

#### `nova version`
Prints the current Nova CLI version.
```bash
nova version
```

---

## Architecture Principles

1. **Separation of Concerns**:
   - **Interactive Prompt Wizards**: Handled in command handlers (`Run: func(...)`) using `survey/v2`.
   - **Pure Generators**: Core logic functions (`GenerateDockerfile`, `GenerateCompose`, `GenerateK8sDeployment`, `BuildPodmanRunArgs`) accept configuration structs and return pure string/slice outputs.

2. **Podman-Native Runtime Integration**:
   - Podman is the primary container engine for GalactOS.
   - Real-time build/run logs stream directly to standard stdout/stderr streams.
   - Missing executable checks provide clear actionable installation steps (`nova install podman`).

3. **Comprehensive Testability**:
   - Generator functions and argument builders are 100% unit-testable without requiring a running Podman daemon or file IO side effects.

---

## Running Tests

Run the complete unit test suite:

```bash
go test -v ./...
```

Run static analysis:

```bash
go vet ./...
```

---

## License

Copyright © 2026 Deepsayan Das. Distributed under the project license terms.
