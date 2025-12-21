# NLU 服务
详细文档：：[意图识别指南](https://ccnlswpbwzhg.feishu.cn/wiki/IHPmwE9IVicutykukbJch7Jkncb)


## 环境
目前开发设备：
- 操作系统：Ubuntu 22.04
- 处理器：Intel Core i7-12700H
- 内存：16GB DDR4
- 显卡：NVIDIA GeForce RTX 3050 Laptop GPU

本项目使用 miniconda 环境管理依赖，当前环境为 `nlp`。
```bash
10:16:18 pp@bb Yunyez ±|feat/ai-nlu ✗|→ conda info --envs

# conda environments:
#
base                   /home/pp/miniconda3
nlp                    /home/pp/miniconda3/envs/nlp
```
当前使用的编辑器是 Visual Studio Code。  
这里需要安装 Python 扩展插件，以支持 `nlp` 环境。
在 vscode 中，按下 `Ctrl+Shift+P` 打开命令面板，输入 `Python: Select Interpreter` 并回车。系统会列出已检测到的 Python 解释器，包括全局安装的版本和虚拟环境路径。选择 `/home/pp/miniconda3/envs/nlp/bin/python` 即可。
> tips： 具体版本以你自己的环境为准。这里仅作参考。  

```bash
==============================
Python: 3.10.18
CUDA 可用: True
GPU 型号: NVIDIA GeForce RTX 3050 Laptop GPU
PyTorch 版本: 2.5.1+cu121
CUDA 版本: 12.1
==============================
```

## 模型训练

模型缓存：/home/pp/.cache/huggingface/hub（具体路径根据环境和配置而异）

> tips：更新训练数据后需要重新训练模型。
首先运行 `train.py` 训练模型时，会自动下载模型和数据集。如果网络环境不好，可能会失败。可以手动下载模型和数据集，放到缓存目录中。这里建议使用代理，或者使用镜像站。只需要下载一次，后续训练都不会再下载。

```bash
# 训练模型
python ./ai/nlu/src/train.py
# 启动服务
python -m uvicorn ai.nlu.src.server:app --host 0.0.0.0 --port 8001
```

## 使用调试
打开浏览器，访问 `http://localhost:8001/docs` 即可查看 API 文档。


## 补充

### 🧠 Sentence-BERT 是黑盒吗？其实是“灰盒”
虽然你没看到张量运算过程，但它的行为是高度可解释的：

| 组件                  | 可控性       | 你能做什么                                   |
|-----------------------|--------------|----------------------------------------------|
| Encoder（SBERT）      | 固定（frozen）| 选择不同模型（中文/多语言/小模型）|
| Embedding 向量        | 可见         | 打印 emb.shape，甚至可视化 t-SNE              |
| 分类器（LogisticRegression） | 完全透明     | 查看权重、概率、决策边界                     |

> ✅ 你完全不需要理解 BERT 内部怎么算 attention，只要知道：  
“它把句子变成一个 384 维向量，相似句子向量靠近，不同意图远离

### 🛠 你现在最该关注的：数据 > 模型
在 NLU 意图识别任务中：
90% 的效果提升来自更好的数据，10% 来自模型调优。  
给出合理粒度的训练数据，模型效果会有显著提升。

