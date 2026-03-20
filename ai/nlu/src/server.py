"""NLU 服务入口 - 同时支持 HTTP 和 gRPC"""
import threading
import logging
from fastapi import FastAPI
from fastapi import HTTPException
from pydantic import BaseModel
import uvicorn
import numpy as np
import os

from .grpc_server import serve as serve_grpc
from sentence_transformers import SentenceTransformer
import joblib

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# ---------------------------------------------
# hugging face 镜像设置
# ---------------------------------------------
os.environ["HF_ENDPOINT"] = "https://hf-mirror.com"

# Load model on startup
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
MODEL_DIR = os.path.join(os.path.dirname(SCRIPT_DIR), "model")

logger.info("Loading NLU model from local or HF (via mirror)...")
encoder = SentenceTransformer(os.path.join(MODEL_DIR, "encoder"))
classifier = joblib.load(os.path.join(MODEL_DIR, "classifier.pkl"))

COMMAND_INTENTS = {
    "play_music",
    "set_temperature",
    "turn_on_light",
    "turn_off_light",
    "query_weather",
    "chit_chat",
    "deny_action"
}

app = FastAPI(title="NLU Service", version="1.0")

class NLURequest(BaseModel):
    text: str

class NLUResponse(BaseModel):
    text: str
    intent: str
    confidence: float
    is_command: bool


@app.post("/nlu", response_model=NLUResponse)
def predict(request: NLURequest):
    text = request.text
    emb = encoder.encode([text])
    intent = classifier.predict(emb)[0]
    confidence = float(np.max(classifier.predict_proba(emb)))
    is_command = intent in COMMAND_INTENTS
    return {
        "text": text,
        "intent": intent,
        "confidence": round(confidence, 4),
        "is_command": is_command
    }


@app.get("/health")
def health():
    return {"status": "ok"}


def run_grpc_server(port=50052):
    """在独立线程中运行 gRPC 服务器"""
    serve_grpc(port)


def main():
    """启动 HTTP + gRPC 双协议服务"""
    # 启动 gRPC 服务器（独立线程）
    grpc_thread = threading.Thread(target=run_grpc_server, args=(50052,), daemon=True)
    grpc_thread.start()

    logger.info("Starting NLU service with HTTP (8001) + gRPC (50052)")
    
    # 启动 HTTP 服务器
    uvicorn.run(app, host="0.0.0.0", port=8001)


if __name__ == "__main__":
    main()
