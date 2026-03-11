"""ASR 服务入口 - 同时支持 HTTP 和 gRPC"""
import threading
import logging
from fastapi import FastAPI, Request
import uvicorn

from .grpc_server import serve as serve_grpc
from .engine import asr_engine
import tempfile
import os

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="Yunyez ASR Service")

@app.post("/asr")
async def transcribe(request: Request):
    """
    语音识别接口，接收 POST 请求，返回识别文本
    """
    try:
        body = await request.body()
        if len(body) == 0:
            from fastapi import HTTPException
            raise HTTPException(status_code=400, detail="Empty audio data")

        # 判断是否是 WAV（通过 magic header）
        is_wav = body.startswith(b"RIFF") and b"WAVE" in body[:12]

        # 写入临时文件
        with tempfile.NamedTemporaryFile(delete=False, suffix=".wav" if is_wav else ".pcm") as tmp:
            tmp.write(body)
            tmp_path = tmp.name

        try:
            text = asr_engine.transcribe(tmp_path)
            return {"text": text}
        finally:
            os.unlink(tmp_path)

    except Exception as e:
        from fastapi import HTTPException
        raise HTTPException(status_code=500, detail=f"ASR failed: {str(e)}")


@app.post("/asr_test")
async def transcribe_test(audio: bytes):
    """
    测试语音识别接口
    """
    try:
        with tempfile.NamedTemporaryFile(delete=False, suffix=".pcm") as tmp:
            tmp.write(audio)
            tmp_path = tmp.name

        try:
            text = asr_engine.transcribe(tmp_path)
            return {"text": text}
        finally:
            os.unlink(tmp_path)
    except Exception as e:
        from fastapi import HTTPException
        raise HTTPException(status_code=500, detail=f"ASR failed: {str(e)}")


@app.get("/health")
async def health():
    return {"status": "ok"}


def run_grpc_server(port=50051):
    """在独立线程中运行 gRPC 服务器"""
    serve_grpc(port)


def main():
    """启动 HTTP + gRPC 双协议服务"""
    # 启动 gRPC 服务器（独立线程）
    grpc_thread = threading.Thread(target=run_grpc_server, args=(50051,), daemon=True)
    grpc_thread.start()

    logger.info("Starting ASR service with HTTP (8002) + gRPC (50051)")
    
    # 启动 HTTP 服务器
    uvicorn.run(app, host="0.0.0.0", port=8002)


if __name__ == "__main__":
    main()
