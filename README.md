# Docker Optimization Guide

A hands-on, step-by-step guide to reducing Docker image sizes by up to **98%**. Learn progressive optimization techniques with real **Go** and **Python** applications.

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

```
Go Image Size Reduction
Step 1  ███████████████████████████████████████  369 MB
Step 2  ██████████████                           134 MB
Step 3  █                                       8.27 MB
Step 4  ▏                                       4.23 MB

Python Image Size Reduction
Step 1  ███████████████████████████████████████  513 MB
Step 2  █████████████                            170 MB
Step 3  ███████                                 91.8 MB
Step 4  ██████                                  75.5 MB
```

<br/>

## Sample Applications

This guide uses two realistic applications rather than trivial "hello world" examples:

| | Go | Python |
|---|---|---|
| **App** | User REST API | ML Prediction API |
| **Framework** | gorilla/mux + Prometheus | FastAPI + pandas + numpy |
| **Endpoints** | CRUD `/users`, `/health`, `/metrics` | `/predict`, `/health` |
| **Port** | 8080 | 8000 |

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

---

### Step 2 - Use a Smaller Base Image

Switch to Alpine (Go) or Slim (Python) variants to drop most of the OS bloat.

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

---

### Step 3 - Multi-Stage Build (Go) / Virtual Build Deps (Python)

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

<br/>

## Quick Start

```bash
# Clone the repository
git clone https://github.com/somaz94/docker-optimization-guide.git
cd docker-optimization-guide

# Build and compare Go images
cd go
docker build -f Dockerfile.step1 -t go-step1 .
docker build -f Dockerfile.step2 -t go-step2 .
docker build -f Dockerfile.step3 -t go-step3 .
docker build -f Dockerfile.step4 -t go-step4 .

# Build and compare Python images
cd ../python
docker build -f Dockerfile.step1 -t py-step1 .
docker build -f Dockerfile.step2 -t py-step2 .
docker build -f Dockerfile.step3 -t py-step3 .
docker build -f Dockerfile.step4 -t py-step4 .

# Compare sizes
docker images | grep -E "(go|py)-step"
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
```

<br/>

## Optimization Techniques Summary

| Technique | Impact | Applicable To |
|-----------|--------|---------------|
| Use Alpine/Slim base images | High | All languages |
| Multi-stage builds | Very High | All languages |
| Strip debug symbols (`-ldflags="-w -s"`) | Medium | Go, C, C++, Rust |
| Static binary + scratch | Very High | Go, C, Rust |
| `--no-cache-dir` (pip) | Low-Medium | Python |
| Virtual build deps (`--virtual`) | Medium | Alpine-based images |
| `.dockerignore` | Low | All |
| Single RUN layer for install + cleanup | Medium | All |

<br/>

## Best Practices

- **Always use `.dockerignore`** to exclude `.git`, `node_modules`, logs, and other non-essential files
- **Order layers by change frequency** - put `COPY requirements.txt` before `COPY . .` to leverage Docker build cache
- **Use specific image tags** (e.g., `python:3.12-alpine`) instead of `latest` for reproducible builds
- **Run as non-root user** in production (add `USER nonroot` or use distroless images)
- **Scan your images** with `docker scout`, `trivy`, or `grype` to catch vulnerabilities

<br/>

## Project Structure

```
docker-optimization-guide/
├── go/
│   ├── main.go              # User REST API (gorilla/mux + Prometheus)
│   ├── go.mod
│   ├── go.sum
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
├── README.md
└── LICENSE
```

<br/>

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
