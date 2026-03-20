"""TTS 服务入口 - 同时支持 HTTP 和 gRPC"""
import threading
import logging
from fastapi import FastAPI
from fastapi.responses import StreamingResponse
from pydantic import BaseModel
import uvicorn
import edge_tts

from .grpc_server import serve as serve_grpc

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="Edge TTS Service", version="1.0")

class TTSRequest(BaseModel):
    text: str = "你好呀，我是小云也"
    voice: str = "zh-CN-XiaoyiNeural"
    rate: str = "+0%"
    pitch: str = "+0Hz"
    volume: str = "+0%"


@app.post(
    "/tts",
    response_class=StreamingResponse,
    responses={
        200: {
            "content": {"audio/mpeg": {}},
            "description": "MP3 audio stream of synthesized speech.",
        }
    },
    response_description="MP3 audio binary stream"
)
async def tts(request: TTSRequest):
    communicate = edge_tts.Communicate(
        text=request.text,
        voice=request.voice,
        rate=request.rate,
        pitch=request.pitch,
        volume=request.volume,
    )

    async def audio_stream():
        async for chunk in communicate.stream():
            if chunk["type"] == "audio":
                yield chunk["data"]

    return StreamingResponse(
        audio_stream(),
        media_type="audio/mpeg",
        headers={
            "Content-Disposition": "inline; filename=tts_output.mp3"
        },
    )


@app.get("/health")
async def health():
    return {"status": "ok"}


def run_grpc_server(port=50053):
    """在独立线程中运行 gRPC 服务器"""
    serve_grpc(port)


def main():
    """启动 HTTP + gRPC 双协议服务"""
    # 启动 gRPC 服务器（独立线程）
    grpc_thread = threading.Thread(target=run_grpc_server, args=(50053,), daemon=True)
    grpc_thread.start()

    logger.info("Starting TTS service with HTTP (8003) + gRPC (50053)")
    
    # 启动 HTTP 服务器
    uvicorn.run(app, host="0.0.0.0", port=8003)


if __name__ == "__main__":
    main()
