from fastapi import FastAPI
from fastapi.responses import StreamingResponse
from pydantic import BaseModel
import edge_tts

app = FastAPI(title="Edge TTS Service", version="1.0")

class TTSRequest(BaseModel):
    text: str = "你好呀,我是小云也"
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
        }
    )