"""TTS gRPC 服务"""
import grpc
from concurrent import futures
import logging
import asyncio

from .pb import tts_pb2
from .pb import tts_pb2_grpc
from .pb import types_pb2

import edge_tts

logger = logging.getLogger(__name__)


class TTSServicer(tts_pb2_grpc.TTSServiceServicer):
    """TTS gRPC 服务实现"""

    def Synthesize(self, request, context):
        """单次语音合成"""
        try:
            # 将 rate/pitch/volume 从 gRPC float 转换为 edge-tts 字符串格式
            rate_str = f"{int((request.rate - 1.0) * 100):+d}%" if request.rate != 1.0 else "+0%"
            pitch_str = f"{int((request.pitch - 1.0) * 100):+d}Hz" if request.pitch != 1.0 else "+0Hz"
            volume_str = f"{int((request.volume - 1.0) * 100):+d}%" if request.volume != 1.0 else "+0%"

            # 使用 edge-tts 合成
            communicate = edge_tts.Communicate(
                text=request.text,
                voice=request.voice if request.voice else "zh-CN-XiaoyiNeural",
                rate=rate_str,
                pitch=pitch_str,
                volume=volume_str,
            )

            # 收集音频数据
            audio_chunks = []
            asyncio.run(self._collect_audio(communicate, audio_chunks))

            audio_data = b''.join(audio_chunks)

            # 计算时长（假设 24kHz 16bit mono: 48KB/s）
            duration_ms = len(audio_data) * 1000 // 48000

            return tts_pb2.SynthesizeResponse(
                audio_content=audio_data,
                duration_ms=duration_ms
            )

        except Exception as e:
            logger.error(f"TTS synthesize error: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"TTS failed: {str(e)}")
            return tts_pb2.SynthesizeResponse(
                error=types_pb2.Error(code=500, message=f"TTS failed: {str(e)}")
            )

    async def _collect_audio(self, communicate, audio_chunks):
        """收集音频数据"""
        async for chunk in communicate.stream():
            if chunk["type"] == "audio":
                audio_chunks.append(chunk["data"])

    def StreamingSynthesize(self, request_iterator, context):
        """流式语音合成"""
        config = None

        try:
            text_chunks = []

            for request in request_iterator:
                if request.HasField('config'):
                    config = request.config
                    logger.info(f"Streaming TTS config: {config.voice}")
                elif request.text:
                    text_chunks.append(request.text)

            # 合并文本
            full_text = ''.join(text_chunks)

            if not full_text:
                yield tts_pb2.StreamingSynthesizeResponse(
                    error=types_pb2.Error(code=400, message="Empty text")
                )
                return

            # 转换为 edge-tts 格式
            rate_str = f"{int((config.rate - 1.0) * 100):+d}%" if config and config.rate != 1.0 else "+0%"
            pitch_str = f"{int((config.pitch - 1.0) * 100):+d}Hz" if config and config.pitch != 1.0 else "+0Hz"
            volume_str = f"{int((config.volume - 1.0) * 100):+d}%" if config and config.volume != 1.0 else "+0%"

            communicate = edge_tts.Communicate(
                text=full_text,
                voice=config.voice if config and config.voice else "zh-CN-XiaoyiNeural",
                rate=rate_str,
                pitch=pitch_str,
                volume=volume_str,
            )

            # 收集并发送音频数据
            audio_chunks = []
            asyncio.run(self._collect_audio(communicate, audio_chunks))

            for chunk in audio_chunks:
                yield tts_pb2.StreamingSynthesizeResponse(
                    audio_content=chunk
                )

        except Exception as e:
            logger.error(f"Streaming TTS error: {e}")
            yield tts_pb2.StreamingSynthesizeResponse(
                error=types_pb2.Error(code=500, message=f"TTS failed: {str(e)}")
            )

    def Health(self, request, context):
        """健康检查"""
        return types_pb2.HealthResponse(status="ok", version="1.0.0")


def serve(port=50053):
    """启动 gRPC 服务"""
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    tts_pb2_grpc.add_TTSServiceServicer_to_server(TTSServicer(), server)
    server.add_insecure_port(f'[::]:{port}')
    server.start()
    logger.info(f"TTS gRPC server started on port {port}")
    server.wait_for_termination()


if __name__ == '__main__':
    serve()
