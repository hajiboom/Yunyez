## 语音识别服务

from fastapi import FastAPI, File, UploadFile, HTTPException, Request
from .engine import asr_engine
import uvicorn
import tempfile
import os


app = FastAPI(title="Yunyez ASR Service")

@app.post("/asr")
async def transcribe(request: Request):
    """
    语音识别接口，接收 POST 请求，返回识别文本
    
    :param request: 音频二进制数据
    :type request: Request
    :return: 包含识别文本的 JSON 响应
    :rtype: dict
    """
    try:
        body = await request.body()
        if len(body) == 0:
            raise HTTPException(status_code=400, detail="Empty audio data")

        # 判断是否是 WAV（通过 magic header）
        is_wav = body.startswith(b"RIFF") and b"WAVE" in body[:12]

        # 写入临时文件 后缀根据是否是 WAV 确定
        with tempfile.NamedTemporaryFile(delete=False, suffix=".wav" if is_wav else ".pcm") as tmp:
            if is_wav:
                tmp.write(body)
                tmp.flush()
                tmp_path = tmp.name
            else:
                # 假设是 raw PCM（需约定格式：16kHz, 16bit, mono）
                tmp.write(body)
                tmp.flush()
                tmp_path = tmp.name + ".pcm"

        try:
            print("[tmp_path]: ",tmp_path)
            text = asr_engine.transcribe(tmp_path)
            return {"text": text}
        finally:
            os.unlink(tmp_path)

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"ASR failed: {str(e)}")



@app.post("/asr_test")
async def transcribe_test(audio: UploadFile = File(...)):
    """
    测试语音识别接口，上传音频文件进行识别
    
    :param audio: 上传的音频文件，支持 WAV/MP3/PCM 格式
    :type audio: UploadFile 
    :return: 包含识别文本的 JSON 响应
    :rtype: dict
    """
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
    uvicorn.run(app, host="127.0.0.1", port=8002)