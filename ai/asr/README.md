# 语音识别

## 功能描述
基于 FunASR 实现语音识别功能。微调 Paraformer 模型，提高识别准确率。


## 目录结构
```
./ai/asr
├── data    # 语音数据目录
│   ├── audio 
│   │   └── test.wav
│   └── train.list
├── __init__.py
├── out     # 输出目录
├── README.md
├── scripts     # 脚本目录
│   ├── fine_tune.py  # 微调脚本
│   └── test.py  # 测试脚本
└── src
    ├── config.py  # 配置文件
    ├── engine.py  # 引擎文件
    ├── __init__.py
    └── utils.py  # 工具文件
```


## 环境配置

```bash
# 安装 FunASR（官方推荐方式）根据已有话你就嗯选择无冲突版本
pip install funasr==1.1.8 modelscope==1.17.0
pip install addict pyyaml
```

## 测试示例

```bash
05:44:06 (nlp) pp@bb Yunyez ±|feat/asr ✗|→ python ai/asr/scripts/test.py
funasr version: 1.1.8.
2025-12-15 17:47:03,213 - modelscope - WARNING - Using branch: master as version is unstable, use with caution
/home/pp/miniconda3/envs/nlp/lib/python3.10/site-packages/funasr/train_utils/load_pretrained_model.py:39: FutureWarning: You are using `torch.load` with `weights_only=False` (the current default value), which uses the default pickle module implicitly. It is possible to construct malicious pickle data which will execute arbitrary code during unpickling (See https://github.com/pytorch/pytorch/blob/main/SECURITY.md#untrusted-models for more details). In a future release, the default value for `weights_only` will be flipped to `True`. This limits the functions that could be executed during unpickling. Arbitrary objects will no longer be allowed to be loaded via this mode unless they are explicitly allowlisted by the user via `torch.serialization.add_safe_globals`. We recommend you start setting `weights_only=True` for any use case where you don't have full control of the loaded file. Please open an issue on GitHub for any issues related to this experimental feature.
  ori_state = torch.load(path, map_location=map_location)
  0%|                                                                                                                  | 0/1 [00:00<?, ?it/s]/home/pp/miniconda3/envs/nlp/lib/python3.10/site-packages/funasr/models/paraformer/model.py:251: FutureWarning: `torch.cuda.amp.autocast(args...)` is deprecated. Please use `torch.amp.autocast('cuda', args...)` instead.
  with autocast(False):
rtf_avg: 0.070: 100%|██████████████████████████████████████████████████████████████████████████████████████████| 1/1 [00:00<00:00,  2.55it/s]
ASR Result: 欢迎大家来体验达摩院推出的语音识别模型。
```