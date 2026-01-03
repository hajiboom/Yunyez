# Yunyez AI Components

This directory contains the AI components for the Yunyez project:
- ASR (Automatic Speech Recognition)
- NLU (Natural Language Understanding)
- TTS (Text-to-Speech)

## Development Setup

### Prerequisites
- Python 3.10.11
- Conda environment with NLP packages
- Docker (for containerized deployment)

### Starting All Services for Development

To start all AI services for development, run:

```bash
./start_all.sh
```

This will start:
- ASR service on http://127.0.0.1:8002
- NLU service on http://127.0.0.1:8001
- TTS service on http://127.0.0.1:8003

## Docker Deployment

### Building and Running with Docker Compose

To build and run all services with Docker:

```bash
cd ai
docker-compose up --build
```

### Individual Service Deployment

Each service can also be built and run individually:

```bash
# Build ASR service
docker build -t yunyez-asr ./asr

# Build NLU service
docker build -t yunyez-nlu ./nlu

# Build TTS service
docker build -t yunyez-tts ./tts
```

## Service Endpoints

### ASR (Automatic Speech Recognition)
- `/asr` - POST endpoint to transcribe audio data
- `/asr_test` - POST endpoint to upload and transcribe audio files

### NLU (Natural Language Understanding)
- `/nlu` - POST endpoint to analyze text and extract intent
- `/health` - GET endpoint to check service health

### TTS (Text-to-Speech)
- `/tts` - GET endpoint to synthesize text to speech
  - Parameters: `text`, `voice`, `rate`, `pitch`, `volume`