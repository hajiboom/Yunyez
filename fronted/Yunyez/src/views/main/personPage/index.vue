<template>
  <div class="person-page">
    <!-- 个人信息卡片 -->
    <div class="person-info">
      <div class="person-info-item">
        <div class="avatar">黄</div>
        <div class="name">
          <h3>黄sir</h3>
          <div class="user-role">超级管理员</div>
        </div>
      </div>

      <!-- 详细信息列表 -->
      <div class="info-list">
        <div class="info-row">
          <span class="info-label">手机号</span>
          <span class="info-value">138****1234</span>
        </div>
        <div class="info-row">
          <span class="info-label">邮箱</span>
          <span class="info-value">huang.xiaoming@example.com</span>
        </div>
        <div class="info-row">
          <span class="info-label">部门</span>
          <span class="info-value">技术研发中心</span>
        </div>
        <div class="info-row">
          <span class="info-label">入职日期</span>
          <span class="info-value">2023-04-01</span>
        </div>
      </div>

      <!-- 修改密码按钮 -->
      <button class="change-pwd-btn" @click="openChangePwdModal">修改密码</button>
    </div>

    <!-- 修改密码弹窗 -->
    <FixPassword :modalVisible="modalVisible" @closeModal="modalVisible = false" />

  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import FixPassword from './assets/fix-password.vue'

// 用户信息
const userInfo = reactive({
  name: '黄小明',
  role: '超级管理员',
  phone: '138****1234',
  email: 'huang.xiaoming@example.com',
  department: '技术研发中心',
  joinDate: '2023-04-01'
})

const modalVisible = ref(false)

// 打开修改密码弹窗
const openChangePwdModal = () => {
  modalVisible.value = true
}

</script>

<style scoped lang="scss">
.person-page {
  padding: 24px;
  background-color: #f5f7fa;
  min-height: 100vh;
  display: flex;
  justify-content: center;
  align-items: flex-start;
}

.person-info {
  
  width: 100%;
  background: #fff;
  border-radius: 24px;
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.05);
  padding: 32px;
  transition: all 0.2s;
}

.person-info-item {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-bottom: 32px;
  padding-bottom: 24px;
  border-bottom: 1px solid #eef2f6;
 
}

.avatar {
  width: 80px;
  height: 80px;
  background: linear-gradient(135deg, #4f46e5, #7c3aed);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  font-weight: 600;
  color: white;
  box-shadow: 0 4px 12px rgba(79, 70, 229, 0.3);
}

.name h3 {
  font-size: 24px;
  font-weight: 600;
  margin: 0 0 8px 0;
  color: #1e293b;
}

.user-role {
  display: inline-block;
  background: #eef2ff;
  color: #4f46e5;
  font-size: 13px;
  font-weight: 500;
  padding: 4px 12px;
  border-radius: 30px;
}

.info-list {
  margin-bottom: 32px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 0;
  border-bottom: 1px solid #f1f5f9;
  
  &:last-child {
    border-bottom: none;
  }
}

.info-label {
  font-size: 15px;
  color: #64748b;
  font-weight: 500;
}

.info-value {
  font-size: 15px;
  color: #0f172a;
  font-weight: 500;
}

.change-pwd-btn {
  background: #4f46e5;
  color: white;
  border: none;
  border-radius: 40px;
  padding: 12px 24px;
  font-size: 15px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  width: 20%;
  display: block;
  margin: 0 auto;
  &:hover {
    background: #4338ca;
    transform: translateY(-1px);
  }
}


</style>
