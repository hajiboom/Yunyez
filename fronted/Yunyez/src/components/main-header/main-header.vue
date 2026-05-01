<template>
  <div class="header">
    <div class="header-icon">
        <i class="iconfont icon-zhedie2" v-if="!isFold" @click="handleFoldChange(true)"></i>
        <i class="iconfont icon-zhedie" v-else @click="handleFoldChange(false)"></i>
    </div>
    <el-dropdown>
       <div class="userCenter" >
          <div class="userInfo">
              <el-avatar :src="userData.avatar" :size="30"></el-avatar>
              <span v-show="!isFold">{{ userData.name }}</span>
              <el-icon class="el-icon--right" style="color: black;">
              <arrow-down />
      </el-icon>
      </div>
    </div>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item @click="handleLogout">
          <el-icon><User /></el-icon>
          <span>退出登录</span>
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
  </div>
    
  
</template>

<script setup>
import { ref } from 'vue'
import { ArrowDown, User } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const userData = JSON.parse(localStorage.getItem('userInfo'))

const emit = defineEmits(['fold-change'])
const isFold = ref(false)
const handleFoldChange = (flag) => {
    isFold.value = !isFold.value
    emit('fold-change', flag)
}
const handleLogout = () => {
    localStorage.removeItem('menuData')
    localStorage.removeItem('userInfo')
    router.push('/login')
}
</script>

<style scoped lang="scss">
.header{
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.userCenter {
  margin: 10px 0;
 outline: none;
  display: flex;
  align-items: center;
  cursor: pointer;
  .userInfo {
    display: flex;
    align-items: center;
    justify-content: flex-start;
    
    .el-avatar {
      margin-left: 0px;
    }
    span {
      margin-left: 8px;
      font-size: 14px;
      color: #555555;
    }
  }

}
</style>