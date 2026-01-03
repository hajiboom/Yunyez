# 模型配置

ASR_CONFIG = {
    "model_name": "iic/speech_paraformer-large-vad-punc_asr_nat-zh-cn-16k-common-vocab8404-pytorch",
    "device": "cuda",  # or "cpu"
    "disable_update": True,
    "postprocess": {
        "remove_space": True,
        "auto_punctuate": True  # 未来可扩展
    }
}
