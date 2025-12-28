<template>
  <div class="common-layout">
    <el-container style="height: 100vh;"> <!-- 给容器加100%视口高度，避免aside高度不足 -->
      <!-- 1. 给el-aside加自定义类名 + 调整行内样式（移除margin-top，移到内部header） -->
      <el-aside 
        width="200px" 
        class="sidebar-aside" 
        style="border-right: 1px solid #eee;"
      >
        <!-- 2. 包裹头部+菜单区域（让footer能auto贴底） -->
        <div class="sidebar-main">
          <!-- 顶部 Logo + 应用名称区域 -->
          <div class="sidebar-header">
            <!-- 示例 Logo（可替换为你的图标） -->
            <i class="iconfont icon-jiqirenfushi"></i>
            <div class="sidebar-title">
              <div class="app-name">云也子</div>
              <div class="app-desc">管理后台</div>
            </div>
          </div>

          <!-- 侧边菜单主体（Element Plus el-menu） -->
          <el-menu
          mode="vertical"
           :default-active="$route.path"
           router
            class="sidebar-menu"
            background-color="transparent"
            text-color="#666"
            active-text-color="#409EFF"
            :unique-opened="true"
          >
            <!-- 总览 -->
            <el-menu-item index="/overview" class="sidebar-menu-item">
              
                <i class="iconfont icon-zonglan"></i>
              
              <span>总览</span>
            </el-menu-item>

            <!-- 设备管理 -->
                 <el-menu-item index="/deviceManage" class="sidebar-menu-item"  >
                <i class="iconfont icon-shouji"></i>
            
              <span>设备管理</span>
            </el-menu-item>
            <!-- 实时图像 -->
            <el-menu-item index="/image" class="sidebar-menu-item">
             
                <i class="iconfont icon-zhaoxiangji"></i>
             
              <span>实时图像</span>
            </el-menu-item>

            <!-- 实时语音 -->
            <el-menu-item index="/voice" class="sidebar-menu-item">
             
                <i class="iconfont icon-yuyin"></i>
             
              <span>实时语音</span>
            </el-menu-item>
          </el-menu>
        </div>

        <!-- 底部用户+退出区域（关键：通过auto贴底） -->
        <div class="sidebar-footer">
          <!-- 用户信息 -->
          <div class="user-info">
            <el-avatar :icon="UserFilled" style="width: 30px;height: 30px;"/>
            <span class="username">123123</span>
          </div>
          <!-- 退出登录 -->
          <div class="logout-btn" @click="handleLogout">
            <i class="iconfont icon-tuichu"></i>
            <span>退出登录</span>
          </div>
        </div>
      </el-aside>
      <el-main><router-view></router-view></el-main>
    </el-container>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { UserFilled } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router' // 补充导入router（否则handleLogout会报错）

const router = useRouter() // 初始化router
const isCollapse = ref(true)
const handleOpen = (key, keyPath) => {
  console.log(key, keyPath)
}
const handleClose = (key, keyPath) => {
  console.log(key, keyPath)
}
const handleLogout = () => {
  router.push({ name: 'Login' })
}
</script>

<style scoped lang="scss">
// 关键：给el-aside设置Flex垂直布局，占满高度
:deep(.sidebar-aside) {
  display: flex;
  flex-direction: column; // 垂直排列
  height: 100%; // 占满父容器（el-container）的高度
  box-sizing: border-box;
  padding: 0;
}

// 头部+菜单的容器（占中间区域）
.sidebar-main {
  flex: 1; // 撑满除footer外的剩余空间
  overflow: auto; // 菜单过多时可滚动
}

.el-menu-vertical-demo:not(.el-menu--collapse) {
  width: 200px;
  min-height: 400px;
}

/* 顶部 Logo + 名称区域 */
.sidebar-header {
  display: flex;
  align-items: center;
  padding: 0 1rem 1rem;
  border-bottom: 1px solid #eee;
  gap: 0.5rem;
  margin-top: 1rem; // 原来的el-aside的margin-top移到这里
  .iconfont {
    font-size: 1.5rem;
    padding: 4%;
    border-radius: 10px;
    background-color: #1729f1;
    color: #fff;
  }
  .sidebar-title {
    .app-name {
      font-size: 1rem;
      font-weight: 600;
      color: #333;
    }
    .app-desc {
      font-size: 0.9rem;
      color: #999;
    }
  }
}

/* 侧边菜单样式 */
.sidebar-menu {
  border-right: none; // 去掉默认右边框
  padding: 16px 0;
  width: 100%; // 确保菜单占满宽度
.sidebar-menu-item{
    display: flex;
    justify-content: flex-start;
    align-items: center;
    gap:10px;
    font-size: 16px;
}
}

/* 底部用户+退出区域（关键：margin-top auto 贴底） */
.sidebar-footer {
  padding: 0 16px 16px;
  border-top: 1px solid #eee;
  margin-top: auto; // 核心：自动填充上方空间，推到底部
  width: 100%;
  box-sizing: border-box;

  .user-info {
    display: flex;
    align-items: center;
    gap: 5px;
    padding: 8px 0;
    .username {
      font-size: 14px;
      color: #666;
    }
  }

  .logout-btn {
    display: flex;
    justify-content: flex-start;
    align-items: center;
    padding: 5px 0;
    cursor: pointer;
    color: #F56C6C; // 退出文字红色
    font-size: 14px;
    .iconfont {
      font-size: 20px;
      margin-right: 10px;
      color: #000000;
    }
  }
}

// 给外层容器加100%高度，确保aside能撑满
.common-layout {
  height: 100vh;
}
.el-main{
  background-color: #e8e8e8;
}
</style>