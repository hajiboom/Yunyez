import pandas as pd
from sentence_transformers import SentenceTransformer
from sklearn.linear_model import LogisticRegression
import joblib
import os

# ---------------------------------------------
# hugging face 镜像设置
# ---------------------------------------------
os.environ["HF_ENDPOINT"] = "https://hf-mirror.com"


# ---------------------------------------------
# 训练模型
# ---------------------------------------------

# 确认执行路径
script_dir = os.path.dirname(os.path.abspath(__file__))
nlu_root_dir = os.path.dirname(script_dir) 

print("NLU root dir:", nlu_root_dir)

data_dir = os.path.join(nlu_root_dir, "data")
model_dir = os.path.join(nlu_root_dir, "model")
print("model dir:", model_dir)
# 确保模型目录存在
os.makedirs(model_dir, exist_ok=True)

df = pd.read_csv(os.path.join(data_dir, "train.csv"))
sentences = df["text"].tolist()
labels = df["intent"].tolist()

print("Loading SBERT model...")
model = SentenceTransformer('paraphrase-multilingual-MiniLM-L12-v2')

print("Encoding sentences...")
embeddings = model.encode(sentences)

print("Training classifier...")
clf = LogisticRegression(max_iter=1000)
clf.fit(embeddings, labels)

print("Saving model...")
joblib.dump(clf, os.path.join(model_dir, "classifier.pkl"))
model.save(os.path.join(model_dir, "encoder"))
print(f"✅ Done! Model saved to {model_dir}/")