import sys
import os

# 获取 scripts/ 的父目录（即 ai/asr/）
asr_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
sys.path.append(asr_root)

from src.engine import asr_engine

if __name__ == "__main__":
    audio_path = os.path.join(asr_root, "data", "audio", "test.wav")
    if not os.path.exists(audio_path):
        raise FileNotFoundError(f"Audio file not found: {audio_path}")
    
    text = asr_engine.transcribe(audio_path)
    print(f"ASR Result: {text}")