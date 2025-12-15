
def postprocess_text(text: str, config: dict) -> str:
    if config.get("remove_space", False):
        text = text.replace(" ", "")
    
    if config.get("auto_punctuate", False):
        # 简单兜底：句子结尾无标点则加句号
        if text and text[-1] not in "。！？!?":
            text += "。"
    
    return text.strip()