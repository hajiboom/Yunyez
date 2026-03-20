"""ASR gRPC 服务"""
import grpc
from concurrent import futures
import logging
import tempfile
import os

from .pb import asr_pb2
from .pb import asr_pb2_grpc
from .pb import types_pb2
from .engine import asr_engine

logger = logging.getLogger(__name__)


class ASRServicer(asr_pb2_grpc.ASRServiceServicer):
    """ASR gRPC 服务实现"""

    def Recognize(self, request, context):
        """单次语音识别"""
        try:
            # 判断是否是 WAV（通过 magic header）
            is_wav = request.audio_content.startswith(b"RIFF") and b"WAVE" in request.audio_content[:12]

            # 写入临时文件
            with tempfile.NamedTemporaryFile(delete=False, suffix=".wav" if is_wav else ".pcm") as tmp:
                tmp.write(request.audio_content)
                tmp_path = tmp.name

            try:
                text = asr_engine.transcribe(tmp_path)
                return asr_pb2.RecognizeResponse(
                    transcript=text,
                    confidence=0.95  # 假设置信度
                )
            finally:
                os.unlink(tmp_path)

        except Exception as e:
            logger.error(f"ASR recognize error: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"ASR failed: {str(e)}")
            return asr_pb2.RecognizeResponse(
                error=types_pb2.Error(code=500, message=f"ASR failed: {str(e)}")
            )

    def StreamingRecognize(self, request_iterator, context):
        """流式语音识别"""
        config = None
        audio_chunks = []

        try:
            for request in request_iterator:
                if request.HasField('config'):
                    config = request.config
                    logger.info(f"Streaming ASR config: {config.encoding}, {config.sample_rate}Hz")
                elif request.audio_content:
                    audio_chunks.append(request.audio_content)

            # 合并音频数据
            audio_data = b''.join(audio_chunks)

            # 写入临时文件
            with tempfile.NamedTemporaryFile(delete=False, suffix=".pcm") as tmp:
                tmp.write(audio_data)
                tmp_path = tmp.name

            try:
                text = asr_engine.transcribe(tmp_path)

                # 发送最终结果
                yield asr_pb2.StreamingRecognizeResponse(
                    recognition_result=asr_pb2.StreamingRecognitionResult(
                        transcript=text,
                        confidence=0.95,
                        is_final=True,
                        stability=1.0
                    )
                )
            finally:
                os.unlink(tmp_path)

        except Exception as e:
            logger.error(f"Streaming ASR error: {e}")
            yield asr_pb2.StreamingRecognizeResponse(
                error=types_pb2.Error(code=500, message=f"ASR failed: {str(e)}")
            )

    def Health(self, request, context):
        """健康检查"""
        return types_pb2.HealthResponse(status="ok", version="1.0.0")


def serve(port=50051):
    """启动 gRPC 服务"""
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    asr_pb2_grpc.add_ASRServiceServicer_to_server(ASRServicer(), server)
    server.add_insecure_port(f'[::]:{port}')
    server.start()
    logger.info(f"ASR gRPC server started on port {port}")
    server.wait_for_termination()


if __name__ == '__main__':
    serve()
