<template>
  <!-- 背景容器 -->
  <div class="login-container">
    <!-- 登录卡片 -->
    <el-card class="login-card" shadow="hover">
      <!-- 图标+标题 -->
      <div class="login-header">
        <!-- 替换为你的机器人图标实际路径（建议放 public 文件夹） -->
        <i class="iconfont icon-jiqirenfushi"></i>
        <div class="title-group">
          <div class="main-title">云也子AI机器人</div>
          <div class="sub-title">后台管理系统</div>
        </div>
      </div>
      <!-- 登录表单（带基础校验） -->
      <el-form 
        :model="loginForm" 
        :rules="loginRules" 
        ref="loginFormRef" 
        class="login-form"
      >
        <el-form-item prop="username" label="用户名" label-position="top">
          <el-input
            v-model="loginForm.username"
            placeholder="请输入用户名"
            :prefix-icon="User" 
            size="large"
           
          />
        </el-form-item>

        <el-form-item prop="password" label="密码" label-position="top">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="请输入密码"
            :prefix-icon="Lock"
            size="large"
          />
        </el-form-item>

        <!-- 登录按钮 -->
        <el-form-item>
          <el-button
            type="primary"
            class="login-btn"
            size="large"
            @click="submitLogin"
            
          >
            登 录
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { User, Lock } from '@element-plus/icons-vue'
import{useRouter} from 'vue-router'
// 1. 路由实例（用于跳转）
const router = useRouter()

// 1. 表单实例（用于校验）
const loginFormRef = ref()

// 2. 表单数据
const loginForm = ref({
  username: '',
  password: ''
})

// 3. 表单校验规则
const loginRules = ref({
  username: [
    {  message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在 3 到 20 个字符', trigger: 'blur' }
  ],
  password: [
    {  message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不少于 6 个字符', trigger: 'blur' }
  ]
})

// 4. 登录提交逻辑
const submitLogin = async () => {
  // if (!loginFormRef.value) return
  // try {
  //   // 先做表单校验
  //   const valid = await loginFormRef.value.validate()
  //   if (valid) {
  //     // 校验通过，执行登录逻辑（替换为你的实际接口）
  //     console.log('登录参数：', loginForm.value)

  //     router.push({name:'Dashboard'})
  //   }
  // } catch (error) {
  //   console.error('表单校验失败：', error)
  //   return false
  // }
  router.push({name:'Dashboard'})
}
</script>
<style scoped lang="scss">
/* 背景容器：铺满视口 + 背景图适配 */
.login-container {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  /* 替换为你的背景图路径（public 文件夹下直接写 /xxx.jpg） */
  background-image: url("./assets/loginbak.jpg");
  background-size: cover;    /* 铺满容器，保持比例 */
  background-position: center; /* 居中显示 */
  background-repeat: no-repeat;
  display: flex;
  justify-content: center;
  align-items: center;
  /* 可选：加一层半透明遮罩，提升表单可读性 */
  background-color: rgba(255, 255, 255, 0.1);
  background-blend-mode: overlay;
}

/* 登录卡片：轻量化设计，适配不同屏幕 */
.login-card {
  width: 28%;
  padding: 1% 3%;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.95); /* 半透明白底，提升对比 */
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
  border: none;
  min-width:20rem;
}

/* 头部图标+标题 */
.login-header {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  gap: 1rem;
  margin-bottom: 2rem;
  i{
    font-size:2rem;
 background: linear-gradient(to right, #409eff, #f97316);  border: none;
    color:#e6e6e6;
    padding: 1%;
    display: flex;
    justify-content: center;
    align-items: center;
    width: 3rem;
    height:3rem;
    border-radius: 50%;
  }
}

.title-group {
  text-align: center;
}
.main-title {
  font-size: 1.5rem;
  font-weight: 600;
  color: #1f2937;
}
.sub-title {
  font-size: 1rem;
  color: #6b7280;
  margin-top: 4px;
}

/* 表单样式 */
.login-form {
  margin-top: 1rem;
}

/* 登录按钮 */
.login-btn {
  width: 100%;
  height: 3rem;
  font-size: 1rem;
 background: linear-gradient(to right, #409eff, #f97316);  border: none;
  border-radius: 8px;
  margin-top: 5%;
}
.login-btn:hover {
 background: linear-gradient(to right, #409eff, #f97316);}

</style>