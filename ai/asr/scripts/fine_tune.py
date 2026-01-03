# ai/asr/scripts/finetune.py

"""
【预留】微调脚本
- 数据格式：每行 `audio_path|transcript_with_punctuation`
- 依赖: funasr-train
- 当前为空，待数据积累后启用
"""

def prepare_finetune_data():
    # TODO: 从标注平台导出数据
    pass

def run_finetune():
    # TODO: 调用 funasr-train
    pass

if __name__ == "__main__":
    print("ASR 微调功能暂未启用，请先收集足够标注数据。")