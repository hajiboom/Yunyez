"""NLU gRPC 服务"""
import grpc
from concurrent import futures
import logging
import numpy as np

from .pb import nlu_pb2
from .pb import nlu_pb2_grpc
from .pb import types_pb2

# 导入现有模型
from sentence_transformers import SentenceTransformer
import joblib
import os

logger = logging.getLogger(__name__)

# 加载模型
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
MODEL_DIR = os.path.join(os.path.dirname(SCRIPT_DIR), "model")

encoder = SentenceTransformer(os.path.join(MODEL_DIR, "encoder"))
classifier = joblib.load(os.path.join(MODEL_DIR, "classifier.pkl"))

COMMAND_INTENTS = {
    "play_music",
    "set_temperature",
    "turn_on_light",
    "turn_off_light",
    "query_weather",
    "chit_chat",
    "deny_action"
}


class NLUServicer(nlu_pb2_grpc.NLUServiceServicer):
    """NLU gRPC 服务实现"""

    def Predict(self, request, context):
        """意图识别"""
        try:
            text = request.text
            emb = encoder.encode([text])
            intent = classifier.predict(emb)[0]
            confidence = float(np.max(classifier.predict_proba(emb)))
            is_command = intent in COMMAND_INTENTS

            return nlu_pb2.PredictResponse(
                text=text,
                intent=intent,
                confidence=confidence,
                is_command=is_command
            )

        except Exception as e:
            logger.error(f"NLU predict error: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"NLU failed: {str(e)}")
            return nlu_pb2.PredictResponse(
                error=types_pb2.Error(code=500, message=f"NLU failed: {str(e)}")
            )

    def EmotionJudge(self, request, context):
        """情感识别"""
        # TODO: 实现情感识别
        return nlu_pb2.EmotionJudgeResponse(
            text=request.text,
            emotion="neutral",
            confidence=0.5
        )

    def Health(self, request, context):
        """健康检查"""
        return types_pb2.HealthResponse(status="ok", version="1.0.0")


def serve(port=50052):
    """启动 gRPC 服务"""
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    nlu_pb2_grpc.add_NLUServiceServicer_to_server(NLUServicer(), server)
    server.add_insecure_port(f'[::]:{port}')
    server.start()
    logger.info(f"NLU gRPC server started on port {port}")
    server.wait_for_termination()


if __name__ == '__main__':
    serve()
