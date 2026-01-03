# 核心推理引擎
from funasr import AutoModel
from .config import ASR_CONFIG
from .utils import postprocess_text
import logging

logger = logging.getLogger(__name__)

class ASREngine:
    _instance = None

    def __new__(cls):
        if cls._instance is None:
            cls._instance = super().__new__(cls)
            cls._instance._initialized = False
        return cls._instance

    def __init__(self):
        if not self._initialized:
            logger.info("Initializing ASR engine...")
            self.model = AutoModel(
                model=ASR_CONFIG["model_name"],
                device=ASR_CONFIG["device"],
                disable_update=ASR_CONFIG["disable_update"]
            )
            self._initialized = True

    def transcribe(self, audio_path: str) -> str:
        """
        音频转文本（带后处理）
        """
        try:
            result = self.model.generate(input=audio_path)
            print("FunASR result:", result)  # 👈 关键：打印原始结果
            
            if not result or len(result) == 0:
                return ""
            
            raw_text = result[0].get("text", "")
            print("Raw text:", repr(raw_text))
            
            processed = postprocess_text(raw_text, ASR_CONFIG["postprocess"])
            print("Processed text:", repr(processed))
            return processed
            
        except Exception as e:
            import traceback
            print("Transcribe error:", str(e))
            traceback.print_exc()
            raise

# 全局单例（避免重复加载模型）
asr_engine = ASREngine()