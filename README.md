# Docker Optimization Guide

A hands-on, step-by-step guide to reducing Docker image sizes by up to **98%**. Learn progressive optimization techniques with real **Go**, **Python**, **Node.js**, **Java**, and **Rust** applications.

<br/>

## Results at a Glance

<br/>

### Go (User REST API)

| Step | Base Image | Technique | Size | Reduction |
|------|-----------|-----------|------|-----------|
| 1 | `golang:1.24` | Naive build | **369 MB** | - |
| 2 | `golang:1.24-alpine` | Alpine base | **134 MB** | -63.7% |
| 3 | `golang:1.24-alpine` + `alpine` | Multi-stage build | **8.27 MB** | -97.8% |
| 4 | `golang:1.24-alpine` + `scratch` | Scratch + stripped binary | **4.23 MB** | -98.9% |

<br/>

### Python (ML Prediction API)

| Step | Base Image | Technique | Size | Reduction |
|------|-----------|-----------|------|-----------|
| 1 | `python:3.12` | Naive build | **513 MB** | - |
| 2 | `python:3.12-slim` | Slim base + no-cache-dir | **170 MB** | -66.9% |
| 3 | `python:3.12-alpine` | Alpine + virtual build deps | **91.8 MB** | -82.1% |
| 4 | `python:3.12-alpine` (multi-stage) | Multi-stage build | **75.5 MB** | -85.3% |

<br/>

### Node.js (Task Management API)

| Step | Base Image | Technique | Size | Reduction |
|------|-----------|-----------|------|-----------|
| 1 | `node:22` | Naive build | **410 MB** | - |
| 2 | `node:22-slim` | Slim base + production deps | **82.1 MB** | -80.0% |
| 3 | `node:22-alpine` | Alpine + npm ci | **59.4 MB** | -85.5% |
| 4 | `node:22-alpine` (multi-stage) | Multi-stage + non-root user | **58 MB** | -85.9% |

<br/>

### Java (Book REST API - Spring Boot)

| Step | Base Image | Technique | Size | Reduction |
|------|-----------|-----------|------|-----------|
| 1 | `maven:3.9-eclipse-temurin-21` | Build + runtime in one | **311 MB** | - |
| 2 | `eclipse-temurin:21-jre` | Multi-stage + JRE only | **121 MB** | -61.1% |
| 3 | `eclipse-temurin:21-jre-alpine` | Alpine JRE | **94.8 MB** | -69.5% |
| 4 | `alpine` + custom JRE (jlink) | jlink + Spring Boot layers | **64.8 MB** | -79.2% |

<br/>

### Rust (Note API - Actix Web)

| Step | Base Image | Technique | Size | Reduction |
|------|-----------|-----------|------|-----------|
| 1 | `rust:1.94` | Naive build | **755 MB** | - |
| 2 | `rust:1.94-slim` | Slim base | **486 MB** | -35.6% |
| 3 | `rust:1.94-slim` + `debian-slim` | Multi-stage build | **94.8 MB** | -87.4% |
| 4 | `rust:1.94-alpine` + `scratch` | Static musl build + scratch | **64.8 MB** | -91.4% |

<br/>

```
Go Image Size Reduction
Step 1  ██████████████████████████████████████████████████  369 MB
Step 2  ██████████████████                                  134 MB
Step 3  █                                                  8.27 MB
Step 4  ▏                                                  4.23 MB

Python Image Size Reduction
Step 1  ██████████████████████████████████████████████████  513 MB
Step 2  █████████████████                                   170 MB
Step 3  █████████                                          91.8 MB
Step 4  ███████                                            75.5 MB

Node.js Image Size Reduction
Step 1  ██████████████████████████████████████████████████  410 MB
Step 2  ██████████                                         82.1 MB
Step 3  ███████                                            59.4 MB
Step 4  ███████                                              58 MB

Java Image Size Reduction
Step 1  ██████████████████████████████████████████████████  311 MB
Step 2  ███████████████████                                 121 MB
Step 3  ███████████████                                    94.8 MB
Step 4  ██████████                                         64.8 MB

Rust Image Size Reduction
Step 1  ██████████████████████████████████████████████████  755 MB
Step 2  ████████████████████████████████                    486 MB
Step 3  ██████                                             94.8 MB
Step 4  ████                                               64.8 MB
```

<br/>

## Sample Applications

This guide uses realistic applications rather than trivial "hello world" examples:

| | Go | Python | Node.js | Java | Rust |
|---|---|---|---|---|---|
| **App** | User REST API | ML Prediction API | Task Manager API | Book REST API | Note API |
| **Framework** | gorilla/mux + Prometheus | FastAPI + pandas + numpy | Express + Helmet | Spring Boot + Actuator | Actix Web |
| **Endpoints** | CRUD `/users`, `/health`, `/metrics` | `/predict`, `/health` | CRUD `/tasks`, `/health` | CRUD `/api/books`, `/api/health` | CRUD `/notes`, `/health` |
| **Port** | 8080 | 8000 | 3000 | 8080 | 8080 |

<br/>

## Optimization Steps

<br/>

### Step 1 - The Unoptimized Baseline

Start with the official full-size base image. This is what most tutorials show you.

<details>
<summary><b>Go - Dockerfile.step1</b> (369 MB)</summary>

```dockerfile
FROM golang:1.24

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

EXPOSE 8080
CMD ["./main"]
```

**Problem**: The final image includes the entire Go toolchain, source code, and build cache - none of which are needed at runtime.
</details>

<details>
<summary><b>Python - Dockerfile.step1</b> (513 MB)</summary>

```dockerfile
FROM python:3.12

WORKDIR /app

COPY requirements.txt .
RUN pip install -r requirements.txt

COPY . .

EXPOSE 8000
CMD ["python", "main.py"]
```

**Problem**: `python:3.12` is based on Debian and includes compilers, man pages, and many tools unnecessary for running the app.
</details>

<details>
<summary><b>Node.js - Dockerfile.step1</b> (410 MB)</summary>

```dockerfile
FROM node:22

WORKDIR /app

COPY package*.json ./
RUN npm install

COPY . .

EXPOSE 3000
CMD ["npm", "start"]
```

**Problem**: `node:22` includes the full Debian OS, build tools, yarn, and npm cache. Using `npm start` adds unnecessary overhead vs running `node` directly.
</details>

<details>
<summary><b>Java - Dockerfile.step1</b> (311 MB)</summary>

```dockerfile
FROM maven:3.9-eclipse-temurin-21

WORKDIR /app

COPY pom.xml .
COPY src ./src

RUN mvn package -DskipTests

EXPOSE 8080
CMD ["java", "-jar", "target/book-api-1.0.0.jar"]
```

**Problem**: The final image contains Maven, the full JDK, all build dependencies, and the `.m2` cache - only the JAR is needed at runtime.
</details>

<details>
<summary><b>Rust - Dockerfile.step1</b> (755 MB)</summary>

```dockerfile
FROM rust:1.94

WORKDIR /app

COPY Cargo.toml ./
COPY Cargo.lock* ./
COPY src ./src

RUN cargo build --release

EXPOSE 8080
CMD ["./target/release/note-api"]
```

**Problem**: The Rust toolchain, all compiled dependencies, and build artifacts remain in the image. The toolchain alone is ~500 MB.
</details>

---

### Step 2 - Use a Smaller Base Image

Switch to Alpine or Slim variants to drop most of the OS bloat.

<details>
<summary><b>Go - Dockerfile.step2</b> (134 MB, -63.7%)</summary>

```dockerfile
FROM golang:1.24-alpine

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

EXPOSE 8080
CMD ["./main"]
```

**Key change**: `golang:1.24-alpine` uses musl libc and a minimal Alpine filesystem instead of the full Debian base.
</details>

<details>
<summary><b>Python - Dockerfile.step2</b> (170 MB, -66.9%)</summary>

```dockerfile
FROM python:3.12-slim

WORKDIR /app

RUN apt-get update && apt-get install -y \
    --no-install-recommends \
    gcc \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

EXPOSE 8000
CMD ["python", "main.py"]
```

**Key changes**:
- `python:3.12-slim` removes docs, man pages, and dev tools from the base
- `--no-cache-dir` prevents pip from storing downloaded packages
- `rm -rf /var/lib/apt/lists/*` cleans up apt cache
</details>

<details>
<summary><b>Node.js - Dockerfile.step2</b> (82.1 MB, -80.0%)</summary>

```dockerfile
FROM node:22-slim

WORKDIR /app

COPY package*.json ./
RUN npm install --omit=dev

COPY . .

EXPOSE 3000
CMD ["node", "src/index.js"]
```

**Key changes**:
- `node:22-slim` drops build tools, docs, and extra utilities
- `--omit=dev` skips devDependencies (test frameworks, linters, etc.)
- `node src/index.js` instead of `npm start` avoids spawning an extra process
</details>

<details>
<summary><b>Java - Dockerfile.step2</b> (121 MB, -61.1%)</summary>

```dockerfile
FROM maven:3.9-eclipse-temurin-21-alpine AS builder

WORKDIR /app

COPY pom.xml .
RUN mvn dependency:go-offline

COPY src ./src
RUN mvn package -DskipTests

# Runtime stage
FROM eclipse-temurin:21-jre

WORKDIR /app
COPY --from=builder /app/target/book-api-1.0.0.jar app.jar

EXPOSE 8080
CMD ["java", "-jar", "app.jar"]
```

**Key changes**:
- Multi-stage: Maven + JDK for building, JRE-only for running
- `dependency:go-offline` caches dependencies for better layer reuse
- JRE is ~60% smaller than the full JDK
</details>

<details>
<summary><b>Rust - Dockerfile.step2</b> (486 MB, -35.6%)</summary>

```dockerfile
FROM rust:1.94-slim

WORKDIR /app

RUN apt-get update && apt-get install -y \
    --no-install-recommends \
    pkg-config libssl-dev \
    && rm -rf /var/lib/apt/lists/*

COPY Cargo.toml ./
COPY Cargo.lock* ./
COPY src ./src

RUN cargo build --release

EXPOSE 8080
CMD ["./target/release/note-api"]
```

**Key change**: `rust:1.94-slim` removes docs and extra tools, but the Rust toolchain is still large.
</details>

---

### Step 3 - Multi-Stage Build

Separate the build environment from the runtime environment.

<details>
<summary><b>Go - Dockerfile.step3</b> (8.27 MB, -97.8%)</summary>

```dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main .

# Runtime stage
FROM alpine:latest

WORKDIR /root/

RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /app/main .

EXPOSE 8080
CMD ["./main"]
```

**Key changes**:
- **Multi-stage build**: Build in one image, run in another
- `CGO_ENABLED=0`: Static binary with no C dependencies
- `-ldflags="-w -s"`: Strip debug info and symbol tables
- Runtime uses `alpine:latest` (~8 MB) instead of the Go toolchain image
</details>

<details>
<summary><b>Python - Dockerfile.step3</b> (91.8 MB, -82.1%)</summary>

```dockerfile
FROM python:3.12-alpine

WORKDIR /app

RUN apk add --no-cache \
    libstdc++ \
    openblas

COPY requirements.txt .
RUN apk add --no-cache --virtual .build-deps \
    gcc g++ musl-dev linux-headers \
    gfortran openblas-dev \
    && pip install --no-cache-dir -r requirements.txt \
    && apk del .build-deps

COPY . .

EXPOSE 8000
CMD ["python", "main.py"]
```

**Key changes**:
- Alpine base instead of Slim (much smaller filesystem)
- `--virtual .build-deps`: Groups build-only packages for easy removal
- Compilers are installed, used for pip install, then deleted in a **single RUN layer**
- Only runtime libs (`libstdc++`, `openblas`) remain
</details>

<details>
<summary><b>Node.js - Dockerfile.step3</b> (59.4 MB, -85.5%)</summary>

```dockerfile
FROM node:22-alpine

WORKDIR /app

COPY package*.json ./
RUN npm ci --omit=dev

COPY . .

USER node

EXPOSE 3000
CMD ["node", "src/index.js"]
```

**Key changes**:
- `node:22-alpine` is ~5x smaller than `node:22-slim`
- `npm ci` ensures reproducible installs from lockfile (faster, stricter)
- `USER node` runs as non-root for better security
</details>

<details>
<summary><b>Java - Dockerfile.step3</b> (94.8 MB, -69.5%)</summary>

```dockerfile
FROM maven:3.9-eclipse-temurin-21-alpine AS builder

WORKDIR /app

COPY pom.xml .
RUN mvn dependency:go-offline

COPY src ./src
RUN mvn package -DskipTests

# Runtime stage with Alpine JRE
FROM eclipse-temurin:21-jre-alpine

WORKDIR /app

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/target/book-api-1.0.0.jar app.jar

USER appuser

EXPOSE 8080
CMD ["java", "-jar", "app.jar"]
```

**Key changes**:
- `eclipse-temurin:21-jre-alpine` combines JRE-only with Alpine (~95 MB vs ~121 MB)
- Non-root user for production security
</details>

<details>
<summary><b>Rust - Dockerfile.step3</b> (94.8 MB, -87.4%)</summary>

```dockerfile
FROM rust:1.94-slim AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y \
    --no-install-recommends \
    pkg-config libssl-dev \
    && rm -rf /var/lib/apt/lists/*

COPY Cargo.toml ./
COPY Cargo.lock* ./
COPY src ./src

RUN cargo build --release

# Runtime stage
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/target/release/note-api /usr/local/bin/

EXPOSE 8080
CMD ["note-api"]
```

**Key changes**:
- Multi-stage: Rust toolchain is left in the builder stage
- Runtime uses `debian:bookworm-slim` (~80 MB) - only libc and CA certs needed
- Binary is dynamically linked against glibc
</details>

---

### Step 4 - Maximum Optimization

Push to the smallest possible image for each language.

<details>
<summary><b>Go - Dockerfile.step4</b> (4.23 MB, -98.9%)</summary>

```dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o main .

# Final stage: scratch (empty image)
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/main /main

EXPOSE 8080
ENTRYPOINT ["/main"]
```

**Key changes**:
- `FROM scratch`: Absolutely empty base image - 0 bytes
- Manually copy only what's needed: binary, SSL certs, timezone data
- `-a -installsuffix cgo`: Force rebuild all packages for clean static build
- Final image is literally just the binary + certs

> **Note**: `scratch` has no shell, so you can't `docker exec` into the container for debugging. Use `alpine` (Step 3) if you need that capability.
</details>

<details>
<summary><b>Python - Dockerfile.step4</b> (75.5 MB, -85.3%)</summary>

```dockerfile
FROM python:3.12-alpine AS builder

WORKDIR /app

RUN apk add --no-cache \
    gcc g++ musl-dev linux-headers \
    gfortran openblas-dev

COPY requirements.txt .
RUN pip install --no-cache-dir --user -r requirements.txt

# Runtime stage
FROM python:3.12-alpine

WORKDIR /app

RUN apk add --no-cache libstdc++

COPY --from=builder /root/.local /root/.local
COPY . .

ENV PATH=/root/.local/bin:$PATH

EXPOSE 8000
CMD ["python", "main.py"]
```

**Key changes**:
- Multi-stage: build stage has all compilers, runtime stage has none
- `pip install --user`: Installs to `/root/.local` for clean copying
- Only compiled packages and runtime libs are in the final image
- Build tools (`gcc`, `g++`, `gfortran`) are left behind in the builder stage

> **Note**: Python can't use `scratch` because it needs the Python interpreter. Alpine-based multi-stage is the practical minimum.
</details>

<details>
<summary><b>Node.js - Dockerfile.step4</b> (58 MB, -85.9%)</summary>

```dockerfile
FROM node:22-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm ci --omit=dev

# Runtime stage
FROM node:22-alpine

WORKDIR /app

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/node_modules ./node_modules
COPY src/ ./src/
COPY package.json ./

USER appuser

ENV NODE_ENV=production

EXPOSE 3000
CMD ["node", "src/index.js"]
```

**Key changes**:
- Multi-stage separates dependency installation from runtime
- Only `node_modules`, source code, and `package.json` are copied (no lockfile, no npm cache)
- Non-root user with dedicated app group
- `NODE_ENV=production` enables Node.js production optimizations

> **Note**: Node.js can't use `scratch` because it needs the V8 runtime. Alpine is the practical minimum.
</details>

<details>
<summary><b>Java - Dockerfile.step4</b> (64.8 MB, -79.2%)</summary>

```dockerfile
FROM maven:3.9-eclipse-temurin-21-alpine AS builder

WORKDIR /app

COPY pom.xml .
RUN mvn dependency:go-offline

COPY src ./src
RUN mvn package -DskipTests

# Extract Spring Boot layers for better caching
RUN java -Djarmode=layertools -jar target/book-api-1.0.0.jar extract --destination extracted

# Create custom JRE with jlink
FROM eclipse-temurin:21-jdk-alpine AS jre-builder

RUN jlink \
    --add-modules java.base,java.logging,java.naming,java.net.http,java.security.jgss,java.sql,java.management,jdk.unsupported,java.desktop,java.instrument \
    --strip-debug \
    --compress zip-6 \
    --no-header-files \
    --no-man-pages \
    --output /custom-jre

# Minimal runtime
FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache libstdc++ \
    && addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=jre-builder /custom-jre /opt/java
ENV PATH="/opt/java/bin:${PATH}"

COPY --from=builder /app/extracted/dependencies/ ./
COPY --from=builder /app/extracted/spring-boot-loader/ ./
COPY --from=builder /app/extracted/snapshot-dependencies/ ./
COPY --from=builder /app/extracted/application/ ./

USER appuser

EXPOSE 8080
CMD ["java", "org.springframework.boot.loader.launch.JarLauncher"]
```

**Key changes**:
- **3-stage build**: Maven build → jlink JRE → Alpine runtime
- `jlink` creates a custom JRE with only the modules your app needs (~45 MB vs ~100 MB full JRE)
- **Spring Boot layered extraction**: Dependencies are separated for better Docker cache hits
- `--strip-debug --compress zip-6`: Minimize JRE size
- `--no-header-files --no-man-pages`: Remove unnecessary files

> **Note**: This is the most complex Dockerfile but yields the best results. For simpler apps, Step 3 is often good enough.
</details>

<details>
<summary><b>Rust - Dockerfile.step4</b> (64.8 MB, -91.4%)</summary>

```dockerfile
FROM rust:1.94-alpine AS builder

WORKDIR /app

RUN apk add --no-cache musl-dev

COPY Cargo.toml ./
COPY Cargo.lock* ./
COPY src ./src

RUN cargo build --release --target x86_64-unknown-linux-musl || \
    (rustup target add x86_64-unknown-linux-musl && cargo build --release --target x86_64-unknown-linux-musl)

# Minimal runtime
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/target/x86_64-unknown-linux-musl/release/note-api /note-api

EXPOSE 8080
ENTRYPOINT ["/note-api"]
```

**Key changes**:
- `rust:1.94-alpine` + musl target for fully static binary
- `--target x86_64-unknown-linux-musl`: Links against musl libc for `scratch` compatibility
- `FROM scratch`: Zero-byte base image, same as Go Step 4
- Release profile in `Cargo.toml` enables `opt-level = "z"`, `lto = true`, `strip = true`, and `panic = "abort"` for smallest binary

> **Note**: Like Go, `scratch` means no shell access for debugging. Use `debian:bookworm-slim` (Step 3) if you need that.
</details>

<br/>

## Quick Start

```bash
# Clone the repository
git clone https://github.com/somaz94/docker-optimization-guide.git
cd docker-optimization-guide

# Build and compare Go images
cd go
for i in 1 2 3 4; do docker build -f Dockerfile.step$i -t go-step$i .; done

# Build and compare Python images
cd ../python
for i in 1 2 3 4; do docker build -f Dockerfile.step$i -t py-step$i .; done

# Build and compare Node.js images
cd ../nodejs
for i in 1 2 3 4; do docker build -f Dockerfile.step$i -t node-step$i .; done

# Build and compare Java images
cd ../java
for i in 1 2 3 4; do docker build -f Dockerfile.step$i -t java-step$i .; done

# Build and compare Rust images
cd ../rust
for i in 1 2 3 4; do docker build -f Dockerfile.step$i -t rust-step$i .; done

# Compare all sizes
docker images | grep -E "(go|py|node|java|rust)-step" | sort
```

<br/>

### Run the Applications

```bash
# Go - User API
docker run -p 8080:8080 go-step4
curl http://localhost:8080/health
curl http://localhost:8080/users

# Python - ML Prediction API
docker run -p 8000:8000 py-step4
curl http://localhost:8000/health
curl -X POST http://localhost:8000/predict \
  -H "Content-Type: application/json" \
  -d '{"features": [1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0]}'

# Node.js - Task Manager API
docker run -p 3000:3000 node-step4
curl http://localhost:3000/health
curl http://localhost:3000/tasks

# Java - Book API
docker run -p 8080:8080 java-step4
curl http://localhost:8080/api/health
curl http://localhost:8080/api/books

# Rust - Note API
docker run -p 8080:8080 rust-step4
curl http://localhost:8080/health
curl http://localhost:8080/notes
```

<br/>

## Optimization Techniques Summary

| Technique | Impact | Applicable To |
|-----------|--------|---------------|
| Use Alpine/Slim base images | High | All languages |
| Multi-stage builds | Very High | All languages |
| Strip debug symbols (`-ldflags="-w -s"`) | Medium | Go, C, C++, Rust |
| Static binary + scratch | Very High | Go, Rust |
| Custom JRE with jlink | High | Java |
| Spring Boot layered JARs | Medium | Java (Spring Boot) |
| `npm ci --omit=dev` | Medium | Node.js |
| `--no-cache-dir` (pip) | Low-Medium | Python |
| Virtual build deps (`--virtual`) | Medium | Alpine-based images |
| `.dockerignore` | Low | All |
| Single RUN layer for install + cleanup | Medium | All |
| Non-root user (`USER`) | Security | All |

<br/>

## Best Practices

- **Always use `.dockerignore`** to exclude `.git`, `node_modules`, `target/`, logs, and other non-essential files
- **Order layers by change frequency** - put dependency files (`package.json`, `go.mod`, `pom.xml`) before source code to leverage Docker build cache
- **Use specific image tags** (e.g., `python:3.12-alpine`) instead of `latest` for reproducible builds
- **Run as non-root user** in production (add `USER nonroot` or use distroless images)
- **Scan your images** with `docker scout`, `trivy`, or `grype` to catch vulnerabilities
- **Use `npm ci`** instead of `npm install` for deterministic builds in CI/CD
- **Use `jlink`** for Java apps to include only the JRE modules your application needs

<br/>

## Project Structure

```
docker-optimization-guide/
├── go/
│   ├── main.go              # User REST API (gorilla/mux + Prometheus)
│   ├── go.mod / go.sum
│   ├── .dockerignore
│   ├── Dockerfile.step1     # Full golang base (369 MB)
│   ├── Dockerfile.step2     # Alpine base (134 MB)
│   ├── Dockerfile.step3     # Multi-stage + alpine runtime (8.27 MB)
│   └── Dockerfile.step4     # Multi-stage + scratch (4.23 MB)
├── python/
│   ├── main.py              # ML Prediction API (FastAPI + numpy + pandas)
│   ├── requirements.txt
│   ├── .dockerignore
│   ├── Dockerfile.step1     # Full python base (513 MB)
│   ├── Dockerfile.step2     # Slim base (170 MB)
│   ├── Dockerfile.step3     # Alpine + virtual deps (91.8 MB)
│   └── Dockerfile.step4     # Multi-stage alpine (75.5 MB)
├── nodejs/
│   ├── src/index.js         # Task Manager API (Express + Helmet)
│   ├── package.json / package-lock.json
│   ├── .dockerignore
│   ├── Dockerfile.step1     # Full node base (410 MB)
│   ├── Dockerfile.step2     # Slim base (82.1 MB)
│   ├── Dockerfile.step3     # Alpine + npm ci (59.4 MB)
│   └── Dockerfile.step4     # Multi-stage alpine (58 MB)
├── java/
│   ├── src/                 # Book REST API (Spring Boot + Actuator)
│   ├── pom.xml
│   ├── .dockerignore
│   ├── Dockerfile.step1     # Maven + JDK (311 MB)
│   ├── Dockerfile.step2     # Multi-stage + JRE (121 MB)
│   ├── Dockerfile.step3     # Alpine JRE (94.8 MB)
│   └── Dockerfile.step4     # jlink custom JRE (64.8 MB)
├── rust/
│   ├── src/main.rs          # Note API (Actix Web)
│   ├── Cargo.toml
│   ├── .dockerignore
│   ├── Dockerfile.step1     # Full rust base (755 MB)
│   ├── Dockerfile.step2     # Slim base (486 MB)
│   ├── Dockerfile.step3     # Multi-stage + debian-slim (94.8 MB)
│   └── Dockerfile.step4     # musl static + scratch (64.8 MB)
├── README.md
└── LICENSE
```

<br/>

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
