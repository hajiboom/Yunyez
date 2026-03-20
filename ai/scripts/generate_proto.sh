#!/bin/bash
# 生成 Python Protobuf 代码脚本
# 在 ai 目录下执行：./scripts/generate_proto.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
PROTO_DIR="$PROJECT_ROOT/api/proto"
PYTHON_PB_DIR="$PROJECT_ROOT/ai"

echo "Generating Python Protobuf code..."

# ASR
python -m grpc_tools.protoc \
    -I"$PROTO_DIR" \
    --python_out="$PYTHON_PB_DIR/asr/src/pb" \
    --grpc_python_out="$PYTHON_PB_DIR/asr/src/pb" \
    "$PROTO_DIR/ai/asr.proto" \
    "$PROTO_DIR/common/types.proto"

echo "Generated ASR protobuf files"

# NLU
python -m grpc_tools.protoc \
    -I"$PROTO_DIR" \
    --python_out="$PYTHON_PB_DIR/nlu/src/pb" \
    --grpc_python_out="$PYTHON_PB_DIR/nlu/src/pb" \
    "$PROTO_DIR/ai/nlu.proto" \
    "$PROTO_DIR/common/types.proto"

echo "Generated NLU protobuf files"

# TTS
python -m grpc_tools.protoc \
    -I"$PROTO_DIR" \
    --python_out="$PYTHON_PB_DIR/tts/src/pb" \
    --grpc_python_out="$PYTHON_PB_DIR/tts/src/pb" \
    "$PROTO_DIR/ai/tts.proto" \
    "$PROTO_DIR/common/types.proto"

echo "Generated TTS protobuf files"

echo "Python Protobuf generation completed!"
