#!/bin/bash

# Script to start all AI components (ASR, NLU, TTS) for development
# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Transport mode: http or grpc (default: http)
TRANSPORT_MODE="${YUNYEZ_AI_TRANSPORT_MODE:-http}"

echo "Starting Yunyez AI components..."
echo "Transport mode: $TRANSPORT_MODE"

# Function to start ASR service
start_asr() {
    echo "Starting ASR service ($TRANSPORT_MODE)..."
    cd "$SCRIPT_DIR/asr"
    if [ "$TRANSPORT_MODE" = "grpc" ]; then
        python -c "from src.grpc_server import serve; serve(50052)" &
    else
        python -m uvicorn src.server:app --host 127.0.0.1 --port 8002 &
    fi
    ASR_PID=$!
    echo "ASR service started with PID $ASR_PID"
}

# Function to start NLU service
start_nlu() {
    echo "Starting NLU service ($TRANSPORT_MODE)..."
    cd "$SCRIPT_DIR/nlu"
    if [ "$TRANSPORT_MODE" = "grpc" ]; then
        python -c "from src.grpc_server import serve; serve(50051)" &
    else
        python -m uvicorn src.server:app --host 127.0.0.1 --port 8001 &
    fi
    NLU_PID=$!
    echo "NLU service started with PID $NLU_PID"
}

# Function to start TTS service
start_tts() {
    echo "Starting TTS service ($TRANSPORT_MODE)..."
    cd "$SCRIPT_DIR/tts"
    if [ "$TRANSPORT_MODE" = "grpc" ]; then
        python -c "from src.grpc_server import serve; serve(50053)" &
    else
        python -m uvicorn src.edgeTTS:app --host 127.0.0.1 --port 8003 &
    fi
    TTS_PID=$!
    echo "TTS service started with PID $TTS_PID"
}

# Start all services
start_asr
sleep 2  # Wait for ASR to start
start_nlu
sleep 2  # Wait for NLU to start
start_tts

if [ "$TRANSPORT_MODE" = "grpc" ]; then
    echo "All services started in gRPC mode!"
    echo "NLU gRPC: 127.0.0.1:50051"
    echo "ASR gRPC:  127.0.0.1:50052"
    echo "TTS gRPC:  127.0.0.1:50053"
else
    echo "All services started in HTTP mode!"
    echo "ASR: http://127.0.0.1:8002"
    echo "NLU: http://127.0.0.1:8001"
    echo "TTS: http://127.0.0.1:8003"
fi

# Wait for all background processes
wait