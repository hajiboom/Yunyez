<template>
  <el-dialog
    title="语音接入"
    v-model="deviceVoiceVisible"
    width="560px"
    :close-on-click-modal="false"
    destroy-on-close
    class="voice-dialog"
    draggable
  >
    <!--对话框头部-->
    <div class="headDialog">
      <i class="iconfont icon-xianluomao"></i>
      <span>{{ deviceSn }}</span>
    </div>
    <!--消息主体-->
    <div
      class="chat-area"
      ref="chatAreaRef"
    >
      <TransitionGroup name="msg-fade">
        <div
          v-for="msg in messages"
          :key="msg.id"
          class="msgDialog"
          :class="msg.role"
        >
          <div class="avatar" v-if="msg.role === 'device'">
            <i class="iconfont icon-xianluomao"></i>
            <div class="bubble">
              <div class="msgText" v-if="msg.type === 'text'">
                {{ msg.content }}
              </div>
              <div
                class="msgVoice"
                v-if="msg.type === 'voice'"
                @click="playVoice(msg)"
                :class="{ playing: playingMsgId === msg.id }"
                :style="{ minWidth: voiceBubbleWidth(msg.duration) }"
              >
                <i class="iconfont icon-shengyin voice-icon"></i>
                <span class="duration">{{ msg.duration }}″</span>
                <span class="unread-dot" v-if="!msg.read"></span>
              </div>
            </div>
          </div>

          <div class="avatar" v-if="msg.role === 'user'">
            <div class="bubble">
              <div class="msgText" v-if="msg.type === 'text'">
                {{ msg.content }}
              </div>
              <div
                class="msgVoice"
                v-if="msg.type === 'voice'"
                @click="playVoice(msg)"
                :class="{ playing: playingMsgId === msg.id }"
                :style="{ minWidth: voiceBubbleWidth(msg.duration) }"
              >
                <span class="duration">{{ msg.duration }}″</span>
                <i class="iconfont icon-shengyin-copy voice-icon"></i>
              </div>
            </div>
            <el-avatar :src="userInfo.avatar" :size="30"></el-avatar>
          </div>
        </div>
      </TransitionGroup>
    </div>

    <!--消息输入-->
    <div class="input-area">
      <div class="input-wrapper">
        <el-button
          class="voiceSend-btn"
          @mouseenter="isHover = true"
          @mouseleave="isHover = false"
          @click="isVoiceView = !isVoiceView"
        >
          <i
            class="iconfont"
            :class="isHover ? 'icon-shengyin-copy' : 'icon-shengyin'"
          ></i>
        </el-button>
        <el-input
          v-model="inputText"
          placeholder="输入消息..."
          @keyup.enter="sendMessage"
          class="msg-input"
          v-if="isVoiceView"
        />
        <el-button
          v-else
          class="voiceButton"
          @mousedown="startPress"
          @mouseup="endPress"
        >
          <!--取消录音提示-->
          <div class="cancelVoiceRec" v-if="isRecording">上滑取消录音</div>
          按住说话
        </el-button>
        <el-button
          class="send-btn"
          @click="sendMessage"
          :disabled="!inputText.trim()"
        >
          发送
        </el-button>
      </div>
    </div>
   <audio ref="audioRef" style="display: none"></audio>
  </el-dialog>
</template>

<script setup>
import { Microphone, Mute, PhoneFilled } from "@element-plus/icons-vue";
import { computed, ref, nextTick, watch, onBeforeUnmount } from "vue";

const isHover = ref(false); //切换语音图标
const userInfo = JSON.parse(localStorage.getItem("userInfo"));
const props = defineProps({
  deviceSn: { type: String, default: "" },
  deviceVoiceVisible: { type: Boolean, default: false },
});

const emit = defineEmits(["update:deviceVoiceVisible"]);
const deviceVoiceVisible = computed({
  get: () => props.deviceVoiceVisible,
  set: (val) => emit("update:deviceVoiceVisible", val),
});
//控制发送语音显示
const isVoiceView = ref(true);

//长按录音
let pressTimer = null;
let isLongPressing = ref(false); // 是否正在长按中
let startY = 0;
let currentY = 0;
let mediaRecorder = null;
let audioChunks = [];
let startRecordingTime = ref(0);


//录音元素
const audioRef = ref(null)
const playingMsgId = ref(null)

//是否正在录音
const isRecording = ref(false);

const messages = ref([
  {
    id: 1,
    type: "text",
    role: "device",
    content: "你好这里是云也子设备,有什么可以帮助你的？",
  },
]);
const inputText = ref("");
// 添加新消息
const sendMessage = () => {
  const text = inputText.value.trim();
  if (!text) return;
  messages.value.push({
    id: Date.now(),
    role: "user",
    type: "text",
    content: text,
  });
  inputText.value = "";

  // 模拟回复
  setTimeout(() => {
    messages.value.push({
      id: Date.now() + 1,
      type: "text",
      role: "device",
      content: "收到你的消息了！",
      avatar: "/bot.png",
    });
  }, 500);
};

const chatAreaRef = ref("");
//监听消息变化触底到底部
watch(
  messages,
  async () => {
    await nextTick(); //等待DOM更新完成
    if (chatAreaRef.value) {
      chatAreaRef.value.scrollTop = chatAreaRef.value.scrollHeight;
    }
  },
  { deep: true }
);

// 监听对话框显示状态
watch(deviceVoiceVisible, async (newVal) => {
  if (newVal) {
    // 对话框打开时，等待 DOM 渲染完成
    await nextTick();
    // 这里要等消息渲染出来，可能需要多等一次
    setTimeout(async () => {
      if (chatAreaRef.value) {
        chatAreaRef.value.scrollTop = chatAreaRef.value.scrollHeight;
      }
    }, 100);
  }
});

//长按语音
const startPress = (e) => {
  startY = e.clientY; // 按下时记录起始位置
  if (pressTimer) clearTimeout(pressTimer);
  pressTimer = setTimeout(() => {
    isLongPressing.value = true;
    isRecording.value = true;
    //开始录音startRecording()
    window.addEventListener("mousemove", onMouseMove);
    startRecording();
  }, 300);
};
//鼠标移开
const onMouseMove = (e) => {
  currentY = e.clientY;
  const distance = startY - currentY; // 上滑距离（正数）
  if (distance > 30) {
    window.removeEventListener("mousemove", onMouseMove);
    //取消录音
    isRecording.value = false;
    isLongPressing.value = false;
    if (pressTimer) clearTimeout(pressTimer);
    if (mediaRecorder && mediaRecorder.state === 'recording') {
      mediaRecorder.stop();
      mediaRecorder.stream.getTracks().forEach((track) => track.stop());
    }
  }
};
// 结束监听（松开时调用）
const endPress = () => {
  if (pressTimer) {
    clearTimeout(pressTimer);
    pressTimer = null;
  }
  isRecording.value = false;
  window.removeEventListener("mousemove", onMouseMove);
  if (isLongPressing.value) {
    isLongPressing.value = false;
    stopRecordingAndSend();
  }
};
//开始录音
const startRecording = async () => {
  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
    // 如果在等待权限期间用户已松开鼠标，立即释放流
    if (!isLongPressing.value) {
      stream.getTracks().forEach((t) => t.stop());
      return;
    }
    startRecordingTime.value = Date.now();
    mediaRecorder = new MediaRecorder(stream);
    audioChunks = [];
    mediaRecorder.ondataavailable = (e) => {
      audioChunks.push(e.data);
    };
    mediaRecorder.start();
  } catch (err) {
    console.error("麦克风权限获取失败:", err);
    isRecording.value = false;
    isLongPressing.value = false;
  }
};

// 停止录音
const stopRecordingAndSend = () => {
  if (!mediaRecorder || mediaRecorder.state !== "recording") return;
  mediaRecorder.onstop = () => {
    const blob = new Blob(audioChunks, { type: "audio/webm" });
    const url = URL.createObjectURL(blob);
    const duration = Math.floor((Date.now() - startRecordingTime.value) / 1000);
    sendVoiceMessage(url, duration);
  };
  mediaRecorder.stop();
  mediaRecorder.stream.getTracks().forEach((track) => track.stop());
};

const sendVoiceMessage = (url, duration) => {
  messages.value.push({
    id: Date.now(),
    role: "user",
    type: "voice",
    voiceUrl: url, // 保存音频URL用于播放
    duration: duration, // 录音时长（秒）
  });

  // 自动滚动到底部
  nextTick(() => {
    if (chatAreaRef.value) {
      chatAreaRef.value.scrollTop = chatAreaRef.value.scrollHeight;
    }
  });
};
// 根据语音时长动态计算气泡宽度
const voiceBubbleWidth = (duration) => {
  const minW = 64;
  const maxW = 250;
  const minD = 1;
  const maxD = 60;
  const clamped = Math.max(minD, Math.min(maxD, duration));
  const width = minW + ((clamped - minD) / (maxD - minD)) * (maxW - minW);
  return `${Math.round(width)}px`;
};

const playVoice = (msg) => {
  const audio = audioRef.value

  // 如果正在播放当前消息，暂停
  if (playingMsgId.value === msg.id && !audio.paused) {
    audio.pause()
    playingMsgId.value = null
    return
  }

  // 切换播放源并播放
  audio.src = msg.voiceUrl
  audio.play()
  playingMsgId.value = msg.id
  msg.read = true

  audio.onended = () => {
    playingMsgId.value = null
  }
}
</script>
<style lang="scss" scoped>
// ============ Dialog Overrides ============
:deep(.el-dialog) {
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.12), 0 6px 16px rgba(0, 0, 0, 0.07);
}

:deep(.el-dialog__header) {
  padding: 0;
  margin: 0;
  border-bottom: none;
}

:deep(.el-dialog__title) {
  display: none;
}

:deep(.el-dialog__headerbtn) {
  top: 16px;
  right: 16px;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  color: rgba(255, 255, 255, 0.6);
  font-size: 16px;

  &:hover {
    background: rgba(255, 255, 255, 0.12);
    color: #fff;
  }
}

:deep(.el-dialog__body) {
  padding: 0;
}

// ============ Header ============
.headDialog {
  background: #1c1c1e;
  display: flex;
  align-items: center;
  padding: 16px 24px;
  color: #ffffff;
  position: relative;

  &::after {
    content: "";
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: 2px;
    background: linear-gradient(90deg, #e8785a, #f0a880, #e8785a);
    opacity: 0.7;
  }

  i {
    margin-right: 12px;
    font-size: 26px;
    color: #e8785a;
  }

  span {
    font-size: 15px;
    font-weight: 600;
    letter-spacing: 0.5px;
    color: #e4e4e4;
  }
}

// ============ Chat Area ============
.chat-area {
  background: #f3f0eb;
  padding: 18px 16px;
  height: 340px;
  overflow-y: auto;

  &::-webkit-scrollbar {
    width: 4px;
  }
  &::-webkit-scrollbar-track {
    background: transparent;
  }
  &::-webkit-scrollbar-thumb {
    background: rgba(0, 0, 0, 0.1);
    border-radius: 2px;
  }
}

// ============ Messages ============
.msgDialog {
  padding: 5px 0;
  display: flex;
  align-items: flex-end;

  .avatar {
    display: flex;
    align-items: flex-end;
    gap: 8px;

    > i {
      flex-shrink: 0;
      width: 34px;
      height: 34px;
      background: #fff;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 18px;
      color: #e8785a;
      box-shadow: 0 1px 6px rgba(0, 0, 0, 0.06);
    }
  }

  .bubble {
    max-width: 300px;
    padding: 12px 16px;
    font-size: 14px;
    line-height: 1.6;
    word-break: break-word;

    &:has(.msgVoice) {
      padding: 8px 14px;
      width: fit-content;
    }
  }
}

// Device messages (left)
.device {
  justify-content: flex-start;

  .bubble {
    background: #fff;
    color: #2c2c2c;
    border-radius: 16px 16px 16px 4px;
    box-shadow: 0 1px 8px rgba(0, 0, 0, 0.04);

    &:has(.msgVoice) {
      text-align: left;
    }
  }
}

// User messages (right)
.user {
  justify-content: flex-end;

  .bubble {
    background: #1e3a5f;
    color: #fff;
    border-radius: 16px 16px 4px 16px;

    &:has(.msgVoice) {
      text-align: right;
    }
  }

  .el-avatar {
    flex-shrink: 0;
    box-shadow: 0 1px 6px rgba(0, 0, 0, 0.1);
  }
}

// ============ Voice Bubbles ============
.msgVoice {
  display: flex;
  align-items: center;
  gap: 4px;
  cursor: pointer;
  user-select: none;
  position: relative;
  padding: 0;
  transition: min-width 0.25s ease;
 

  .duration {
    font-size: 12px;
    flex-shrink: 0;
  }

  .unread-dot {
    position: absolute;
    top: 0;
    right: -2px;
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background-color: #e8785a;
  }

  .voice-icon {
    font-size: 20px;
    flex-shrink: 0;
    transition: all 0.2s ease;
  }

  &.playing .voice-icon {
    animation: voicePulse 0.65s ease-in-out infinite;
  }
}

.device .msgVoice {
  justify-content: flex-start;

  .duration {
    color: #7a8a9a;
  }
}

.user .msgVoice {
  justify-content: flex-end;

  .voice-icon {
    transform: scaleX(-1);
  }
  .duration {
    color: rgba(255, 255, 255, 0.75);
  }
}

// ============ Input Area ============
.input-area {
  padding: 14px 18px;
  background: #fff;
  border-top: 1px solid #eae6df;
}

.input-wrapper {
  display: flex;
  align-items: center;
  gap: 10px;
}

// Voice toggle button
.voiceSend-btn {
  width: 38px !important;
  height: 38px !important;
  min-width: 38px;
  border-radius: 50% !important;
  border: none !important;
  background: #f5f3ef !important;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: all 0.25s ease;

  i {
    font-size: 18px;
    color: #999;
    transition: all 0.25s;
  }

  &:hover {
    background: #fef0ec !important;

    i {
      color: #e8785a;
    }
  }
}

// Text input
.msg-input {
  flex: 1;

  :deep(.el-input__wrapper) {
    border-radius: 20px;
    padding: 6px 16px;
    background: #f5f3ef;
    border: 1px solid transparent;
    box-shadow: none;
    transition: all 0.25s;

    &:hover {
      background: #efede8;
    }

    &.is-focus {
      background: #fff;
      border-color: #e8785a;
      box-shadow: 0 0 0 3px rgba(232, 120, 90, 0.08);
    }
  }

  :deep(.el-input__inner) {
    font-size: 14px;
    color: #2c2c2c;

    &::placeholder {
      color: #b8b4ad;
    }
  }
}

// Voice recording button
.voiceButton {
  flex: 1;
  height: 42px;
  border-radius: 20px;
  margin: 0;
  position: relative;
  background: #f5f3ef;
  border: 1px solid #e8e5df;
  color: #8c8c8c;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s;

  &:hover {
    background: #efede8;
  }

  &:active {
    background: #fef0ec;
    border-color: #e8785a;
    color: #e8785a;
  }
}

// Cancel recording tooltip
.cancelVoiceRec {
  position: absolute;
  bottom: calc(100% + 14px);
  left: 50%;
  transform: translateX(-50%);
  width: 90%;
  max-width: 240px;
  background: rgba(28, 28, 30, 0.88);
  backdrop-filter: blur(10px);
  color: #fff;
  text-align: center;
  padding: 10px 20px;
  border-radius: 12px;
  font-size: 13px;
  font-weight: 500;
  z-index: 100;
  animation: slideUp 0.2s ease;
  white-space: nowrap;

  &::after {
    content: "";
    position: absolute;
    bottom: -6px;
    left: 50%;
    transform: translateX(-50%);
    width: 0;
    height: 0;
    border-left: 6px solid transparent;
    border-right: 6px solid transparent;
    border-top: 6px solid rgba(28, 28, 30, 0.88);
  }
}

// Send button
.send-btn {
  flex-shrink: 0;
  height: 38px;
  width: 60px;
  padding: 0;
  border-radius: 19px;
  font-size: 14px;
  font-weight: 600;
  background: #1e3a5f;
  border: none;
  color: #fff;
  letter-spacing: 1px;
  transition: all 0.25s;
  margin: 0;

  &:hover {
    background: #264a78;
    transform: translateY(-1px);
    box-shadow: 0 4px 14px rgba(30, 58, 95, 0.3);
  }

  &:active {
    transform: translateY(0);
  }

  &.is-disabled {
    background: #e8e5df;
    color: #c4c0b8;
    box-shadow: none;

    &:hover {
      transform: none;
      box-shadow: none;
    }
  }
}

// ============ Message Transition ============
.msg-fade-enter-active {
  transition: all 0.35s ease-out;
}
.msg-fade-leave-active {
  transition: all 0.2s ease-in;
}
.msg-fade-enter-from {
  opacity: 0;
  transform: translateY(10px) scale(0.97);
}
.msg-fade-leave-to {
  opacity: 0;
}

// ============ Animations ============
@keyframes voicePulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.25;
  }
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateX(-50%) translateY(6px);
  }
  to {
    opacity: 1;
    transform: translateX(-50%) translateY(0);
  }
}
</style>
