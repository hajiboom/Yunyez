<template>
  <div class="login-page">
    <!-- Canvas 粒子背景 -->
    <canvas ref="particleCanvas" class="particle-canvas" />

    <!-- 登录卡片 -->
    <div class="login-card">
      <!-- 头部 -->
      <div class="login-header">
        <div class="logo-wrap">
          <i class="iconfont icon-xianluomao" />
        </div>
        <h1 class="main-title">云也子AI机器人</h1>
        <p class="sub-title">后台管理系统</p>
      </div>

      <!-- 表单 -->
      <el-form
        ref="loginFormRef"
        :model="loginForm"
        :rules="loginRules"
        class="login-form"
        @keyup.enter="submitLogin"
      >
        <el-form-item prop="username">
          <el-input
            v-model="loginForm.username"
            autocomplete="off"
            placeholder="用户名"
            :prefix-icon="User"
            size="large"
            class="custom-input"
          />
        </el-form-item>

        <el-form-item prop="password">
          <el-input
            v-model="loginForm.password"
            autocomplete="off"
            type="password"
            placeholder="密码"
            :prefix-icon="Lock"
            size="large"
            class="custom-input"
            show-password
          />
        </el-form-item>

        <el-form-item>
          <button type="button" class="login-btn" @click="submitLogin">
            登 录
          </button>
        </el-form-item>
      </el-form>
    </div>

    <p class="footer-tag">YUNYEZ AI BOT</p>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useLoginStore } from '@/store/login'
import { User, Lock } from '@element-plus/icons-vue'

const loginFormRef = ref()
const particleCanvas = ref(null)

const loginForm = ref({
  username: '',
  password: '',
  remember: false
})

const loginStore = useLoginStore()

const loginRules = ref({
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在 3 到 20 个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不少于 6 个字符', trigger: 'blur' }
  ]
})

// ============ Canvas 粒子系统 ============
const COUNT = 60
const CONNECT = 140
const MOUSE_R = 180

let pts = []
let raf = null
let mx = -999
let my = -999

class Pt {
  constructor(w, h) {
    this.x = Math.random() * w
    this.y = Math.random() * h
    this.vx = (Math.random() - 0.5) * 0.35
    this.vy = (Math.random() - 0.5) * 0.35
    this.r = Math.random() * 1.3 + 0.7
  }
  step(w, h) {
    this.x += this.vx
    this.y += this.vy
    if (this.x < 0) this.x = w
    if (this.x > w) this.x = 0
    if (this.y < 0) this.y = h
    if (this.y > h) this.y = 0
  }
  render(ctx) {
    ctx.beginPath()
    ctx.arc(this.x, this.y, this.r, 0, Math.PI * 2)
    ctx.fillStyle = 'rgba(255,255,255,0.5)'
    ctx.fill()
  }
}

function init(w, h) {
  pts = Array.from({ length: COUNT }, () => new Pt(w, h))
}

function draw() {
  const cvs = particleCanvas.value
  if (!cvs) return
  const ctx = cvs.getContext('2d')
  const w = cvs.width
  const h = cvs.height

  ctx.clearRect(0, 0, w, h)

  for (let i = 0; i < pts.length; i++) {
    pts[i].step(w, h)
    pts[i].render(ctx)

    for (let j = i + 1; j < pts.length; j++) {
      const dx = pts[i].x - pts[j].x
      const dy = pts[i].y - pts[j].y
      const d = Math.sqrt(dx * dx + dy * dy)
      if (d < CONNECT) {
        const a = (1 - d / CONNECT) * 0.12
        ctx.beginPath()
        ctx.moveTo(pts[i].x, pts[i].y)
        ctx.lineTo(pts[j].x, pts[j].y)
        ctx.strokeStyle = `rgba(255,255,255,${a.toFixed(2)})`
        ctx.lineWidth = 2
        ctx.stroke()
      }
    }

    const mdx = mx - pts[i].x
    const mdy = my - pts[i].y
    const md = Math.sqrt(mdx * mdx + mdy * mdy)
    if (md < MOUSE_R && md > 0) {
      const a = (1 - md / MOUSE_R) * 0.18
      ctx.beginPath()
      ctx.moveTo(pts[i].x, pts[i].y)
      ctx.lineTo(mx, my)
      ctx.strokeStyle = `rgba(255,255,255,${a.toFixed(2)})`
      ctx.lineWidth = 2
      ctx.stroke()
      pts[i].x += (mdx / md) * 0.2
      pts[i].y += (mdy / md) * 0.2
    }
  }

  raf = requestAnimationFrame(draw)
}

function resize() {
  const cvs = particleCanvas.value
  if (!cvs) return
  cvs.width = window.innerWidth
  cvs.height = window.innerHeight
  init(cvs.width, cvs.height)
}

function mousemove(e) {
  mx = e.clientX
  my = e.clientY
}

onMounted(() => {
  resize()
  raf = requestAnimationFrame(draw)
  window.addEventListener('resize', resize)
  window.addEventListener('mousemove', mousemove)
})

onUnmounted(() => {
  cancelAnimationFrame(raf)
  window.removeEventListener('resize', resize)
  window.removeEventListener('mousemove', mousemove)
})

// ============ 登录逻辑 ============
async function submitLogin() {
  if (!loginFormRef.value) return
  try {
    const valid = await loginFormRef.value.validate()
    if (valid) {
      const transportData = ref({
        transUsername: '',
        transPassword: ''
      })
      loginStore.login(transportData.value)
    }
  } catch {
    // 校验失败
  }
}
</script>

<style scoped lang="scss">
/* ============ 页面 ============ */
.login-page {
  position: fixed;
  inset: 0;
  background: url("./assets/loginbak.jpg") center / cover no-repeat;
  display: flex;
  justify-content: center;
  align-items: center;
  overflow: hidden;
  font-family: 'PingFang SC', 'Microsoft YaHei', 'Helvetica Neue', sans-serif;

  &::before {
    content: '';
    position: absolute;
    inset: 0;
    z-index: 0;
    background: rgba(10, 20, 40, 0.55);
    pointer-events: none;
  }
}

.particle-canvas {
  position: absolute;
  inset: 0;
  z-index: 0;
}

/* ============ 卡片 ============ */
.login-card {
  position: relative;
  z-index: 10;
  width: 400px;
  padding: 52px 48px 44px;
  background: rgba(56, 80, 167, 0.3);
  backdrop-filter: blur(50px);
  -webkit-backdrop-filter: blur(50px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 20px;
  animation: cardIn 0.7s 0.1s cubic-bezier(0.22, 1, 0.36, 1) both;
}

@keyframes cardIn {
  from {
    opacity: 0;
    transform: translateY(24px) scale(0.98);
  }
}

/* ============ 头部 ============ */
.login-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 40px;
}

.logo-wrap {
  width: 52px;
  height: 52px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24px;
  background: rgba(255, 255, 255, 0.06);
  border-radius: 14px;

  i {
    font-size: 30px;
    color: #c8cdd8;
  }
}

.main-title {
  font-size: 20px;
  font-weight: 600;
  color: #eef0f5;
  margin: 0;
  letter-spacing: 2px;
}

.sub-title {
  font-size: 13px;
  color: #888c96;
  margin: 8px 0 0;
  letter-spacing: 3px;
}

/* ============ 表单 ============ */
.login-form {
  :deep(.el-form-item) {
    margin-bottom: 20px;
  }

  :deep(.el-form-item__error) {
    color: #e06060;
    font-size: 12px;
    padding-top: 4px;
    letter-spacing: 0.5px;
  }
}

:deep(.custom-input) {
  .el-input__wrapper {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 10px;
    box-shadow: none;
    padding: 2px 14px;
    transition: border-color 0.25s ease, background 0.25s ease;

    &:hover {
      border-color: rgba(255, 255, 255, 0.2);
      background: rgba(255, 255, 255, 0.07);
    }

    &.is-focus {
      border-color: rgba(180, 190, 210, 0.5);
      background: rgba(255, 255, 255, 0.08);
      box-shadow: none;
    }
  }

  .el-input__inner {
    color: #d8dbe2;
    font-size: 14px;

    &::placeholder {
      color: rgba(255, 255, 255, 0.25);
    }
  }

  .el-input__prefix {
    color: rgba(255, 255, 255, 0.28);
    margin-right: 6px;
    transition: color 0.25s ease;
  }

  .el-input__wrapper.is-focus .el-input__prefix {
    color: rgba(200, 205, 220, 0.6);
  }

  .el-input__suffix {
    color: rgba(255, 255, 255, 0.28);
    transition: color 0.25s ease;
    &:hover { color: rgba(200, 205, 220, 0.6); }
  }
}

/* ============ 按钮 ============ */
.login-btn {
  width: 100%;
  height: 46px;
  font-size: 15px;
  font-weight: 500;
  letter-spacing: 6px;
  color: #1a1d28;
  background: #dbdbdc;
  border: none;
  border-radius: 10px;
  cursor: pointer;
  margin-top: 4px;
  transition: background 0.2s ease, transform 0.2s ease;

  &:hover {
    background: #e0e4ed;
    transform: translateY(-1px);
  }

  &:active {
    transform: translateY(0) scale(0.985);
  }
}

/* ============ 底部 ============ */
.footer-tag {
  position: absolute;
  bottom: 36px;
  left: 50%;
  transform: translateX(-50%);
  font-size: 10px;
  letter-spacing: 5px;
  color: rgba(255, 255, 255, 0.3);
  pointer-events: none;
  z-index: 1;
}

/* ============ 响应式 ============ */
@media (max-width: 480px) {
  .login-card {
    width: calc(100% - 40px);
    padding: 40px 28px 36px;
    border-radius: 16px;
  }

  .main-title {
    font-size: 18px;
  }
}
</style>
