# Yunyez

> "Traveled through mountains and valleys, recorded into the clouds."

## 🎯 Core Values

> **3D Spatial Reconstruction**
- Outputs interactive 3D scenes, not just images/videos

> **Spatial Audio Synchronous Recording**
- Real-time capture of ambient sounds, human voices, and on-site audio
- Aligned with 3D space to restore the authentic atmosphere

> **Travel Digital Footprint**
- Integrated storage of GPS + trajectory + 3D scenes
- Forms a retrievable 3D travel route

> **Multi-Device Collaborative Rendering (Privacy-Safe Mode)**
- Devices in the same area can share reconstructed regions
- Avoids redundant computation, reduces device processing load
- Privacy-first: no face capture, excludes private spaces, no sharing by default

> **Lightweight Display & Sharing**
- Mobile/PC access for 3D scene roaming
- Suitable for personal memories, travel sharing, and scene archiving

## 📁 Project Structure

```
Yunyez/
├── cmd/                      # Application entry points (TODO)
├── configs/                  # Configuration files
│   ├── config.yaml           # Base configuration
│   ├── device.yaml           # Device configuration
│   └── dev/                  # Development environment configuration
│       ├── database.yaml     # Database configuration
│       ├── mqtt.yaml         # MQTT configuration
│       ├── default.yaml      # General configuration
│       ├── ai.yaml           # AI model configuration
│       └── rate_limit.yaml   # Rate limiting configuration
├── docker/                   # Docker related
│   ├── docker-compose.yml    # Infrastructure compose file
│   ├── device/               # Device-side Docker
│   ├── video/                # Video service
│   └── rtsp-server/          # RTSP server
├── internal/                 # Core code
│   ├── app/                  # Application layer
│   ├── common/               # Common modules (config, tools, constants)
│   ├── controller/           # HTTP controllers
│   ├── middleware/           # Middleware (auth, logging, CORS, etc.)
│   ├── model/                # Data models
│   ├── pkg/                  # Common packages
│   │   ├── agent/            # AI Agent (LLM, ASR, TTS, NLU)
│   │   ├── logger/           # Logging wrapper
│   │   ├── mqtt/             # MQTT client and protocol
│   │   ├── postgre/          # PostgreSQL client
│   │   ├── redis/            # Redis client
│   │   ├── transport/        # Transport layer (TCP/UDP)
│   │   ├── rtsp/             # RTSP protocol parsing
│   │   └── media/            # Media format processing
│   ├── service/              # Business service layer
│   ├── types/                # Type definitions
│   └── video/                # Video streaming service (RTSP Server)
├── sql/                      # SQL scripts
│   ├── default.sql
│   ├── agent/                # Agent related tables
│   └── device/               # Device related tables
├── storage/                  # Storage directories
│   ├── logs/                 # Log files
│   └── tmp/audio/            # Temporary audio files
└── example/                  # Example code
    ├── mock/                 # Mock devices (virtual audio/video capture)
    └── scripts/              # Example scripts
```

---

## 🔧 Environment Configuration

### Dependencies

The project depends on the following infrastructure services, which can be started with one click via Docker Compose:

| Service    | Port  | Username | Password | Description     |
|------------|-------|----------|----------|-----------------|
| PostgreSQL | 5432  | postgres | root     | Main database   |
| Redis      | 6379  | -        | -        | Cache service   |
| EMQX       | 1883  | root     | root123  | MQTT Broker     |
| EMQX Dashboard | 18083 | root | root123  | MQTT Web Admin  |

### Configuration Files

- `configs/config.yaml` - Project base configuration (environment, log paths)
- `configs/device.yaml` - Device related configuration
- `configs/dev/database.yaml` - Database connection configuration
- `configs/dev/mqtt.yaml` - MQTT connection configuration
- `configs/dev/default.yaml` - HTTP service configuration
- `configs/dev/ai.yaml` - AI model configuration (LLM, ASR, TTS, NLU)

---

## 🚀 Service Startup

### 1. Start Infrastructure

```bash
# One-click startup (recommended)
./setup-infra.sh

# Or manually
docker compose -f ./docker/docker-compose.yml up -d
```

After startup, access EMQX Dashboard: `http://localhost:18083` (user: root, password: root123)

### 2. Initialize Database

```bash
# Connect to PostgreSQL and execute SQL scripts
psql -h localhost -U postgres -d yunyez -f sql/default.sql
psql -h localhost -U postgres -d yunyez -f sql/device/device.sql
psql -h localhost -U postgres -d yunyez -f sql/agent/cost.sql
```

### 3. Start AI Services

AI services (ASR, NLU, TTS) support both **HTTP** and **gRPC** invocation methods, switched via environment variables.

#### Development Environment (Manual Startup)

```bash
# HTTP mode (default)
cd ai && ./start.sh

# gRPC mode
export YUNYEZ_AI_TRANSPORT_MODE=grpc
cd ai && ./start.sh
```

#### Production Environment (Docker Deployment)

```bash
# HTTP mode
docker compose -f ai/docker-compose.yml --profile http up -d

# gRPC mode
docker compose -f ai/docker-compose.yml --profile grpc up -d
```

| Service | HTTP Port | gRPC Port |
|---------|-----------|-----------|
| NLU     | 8001      | 50051     |
| ASR     | 8002      | 50052     |
| TTS     | 8003      | 50053     |

### 4. Run Backend Service

```bash
# Install Go dependencies
go mod tidy

# Start service
go run .
```

Service listens on: `http://127.0.0.1:8080`

---

## 📋 Environment Variables

The project specifies the environment via the `app.env` field in `configs/config.yaml`, defaulting to the `dev` environment.

Development environment configuration files are located in the `configs/dev/` directory.
