#!/bin/bash

# Script to start all AI components (ASR, NLU, TTS) for development
# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Starting Yunyez AI components..."

# Function to start ASR service
start_asr() {
    echo "Starting ASR service..."
    cd "$SCRIPT_DIR/asr"
    python -m uvicorn src.server:app --host 127.0.0.1 --port 8002 &
    ASR_PID=$!
    echo "ASR service started with PID $ASR_PID"
}

# Function to start NLU service
start_nlu() {
    echo "Starting NLU service..."
    cd "$SCRIPT_DIR/nlu"
    python -m uvicorn src.server:app --host 127.0.0.1 --port 8001 &
    NLU_PID=$!
    echo "NLU service started with PID $NLU_PID"
}

# Function to start TTS service
start_tts() {
    echo "Starting TTS service..."
    cd "$SCRIPT_DIR/tts"
    python -m uvicorn src.edgeTTS:app --host 127.0.0.1 --port 8003 &
    TTS_PID=$!
    echo "TTS service started with PID $TTS_PID"
}

# Start all services
start_asr
sleep 2  # Wait for ASR to start
start_nlu
sleep 2  # Wait for NLU to start
start_tts

echo "All services started!"
echo "ASR: http://127.0.0.1:8002"
echo "NLU: http://127.0.0.1:8001"
echo "TTS: http://127.0.0.1:8003"

# Wait for all background processes
wait