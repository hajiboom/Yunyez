from fastapi import FastAPI
from sentence_transformers import SentenceTransformer
import joblib
import numpy as np
import os


# ---------------------------------------------
# hugging face 镜像设置
# ---------------------------------------------
os.environ["HF_ENDPOINT"] = "https://hf-mirror.com"

# Load model on startup
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
MODEL_DIR = os.path.join(os.path.dirname(SCRIPT_DIR), "model")

print("Loading NLU model from local or HF (via mirror)...")
encoder = SentenceTransformer(os.path.join(MODEL_DIR, "encoder"))  # 先尝试本地
classifier = joblib.load(os.path.join(MODEL_DIR, "classifier.pkl"))

# Define command intents (update as needed)
COMMAND_INTENTS = {
    "play_music", # 播放音乐
    "set_temperature", # 设置温度
    "turn_on_light", # 打开灯
    "turn_off_light", # 关闭灯
    "query_weather" # 查询天气
    "chit_chat" # 闲聊
    "deny_action" # 拒绝操作
}


# ---------------------------------------------
# NLU 服务
# ---------------------------------------------
from pydantic import BaseModel


app = FastAPI(title="NLU Service", version="1.0")

# 定义请求体模型
class NLURequest(BaseModel):
    text: str  # 输入文本

# 定义响应模型（可选但推荐）
class NLUResponse(BaseModel):
    text: str  # 输入文本
    intent: str  # 预测意图
    confidence: float  # 置信度
    is_command: bool  # 是否为命令意图



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