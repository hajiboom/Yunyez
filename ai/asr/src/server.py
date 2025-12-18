## 语音识别服务

from fastapi import FastAPI, File, UploadFile, HTTPException
from .engine import asr_engine
import uvicorn
import tempfile
import os


app = FastAPI(title="Yunyez ASR Service")

@app.post("/v1/transcribe")
async def transcribe(audio: UploadFile = File(...)):
    if not audio.filename.endswith(('.wav', '.mp3', '.pcm')):
        raise HTTPException(status_code=400, detail="Only WAV/MP3/PCM supported")
    
    # 保存临时文件
    with tempfile.NamedTemporaryFile(delete=False, suffix=os.path.splitext(audio.filename)[1]) as tmp:
        tmp.write(await audio.read())
        tmp_path = tmp.name

    try:
        text = asr_engine.transcribe(tmp_path)
        return {"text": text}
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"ASR failed: {str(e)}")
    finally:
        os.unlink(tmp_path)

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8002)