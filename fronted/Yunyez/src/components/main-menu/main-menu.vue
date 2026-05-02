<template>
  <div class="sidebar-container" >
    <!-- 头部 Logo 区域 -->
    <div class="headMenuIcon">
      <i class="iconfont icon-xianluomao"></i>
      <h5 v-show="!isFold">云也子后台管理系统</h5>
    </div>

    <!-- 菜单区域：占据剩余空间，支持滚动 -->
    <el-menu
      :collapse="isFold"
      text-color="#b7bdc3"
      active-text-color="#fff"
      background-color="#0a1f2c"
      :default-active="defaultActive + ''"
      class="menu-wrapper"
      
    >
      <template v-for="item in userMenus" :key="item.id">
        <el-menu-item :index="item.id + ''"  @click="handleItemClick(item)">
          <i :class="'iconfont ' + item.icon" style="font-size: 20px; margin-right: 8px;"></i>
          <template #title>
            <span >{{ item.name }}</span>
          </template>
        </el-menu-item>
      </template>
    </el-menu>

  
    
  </div>
</template>

<script setup>
import { ref,computed } from 'vue'
import { useRouter } from 'vue-router'
import {mapMenusToRoutes,mapPathToMenu} from '@/utils/map-menus.js'
import { useRoute } from 'vue-router'

defineProps({
  isFold: Boolean
})

const router = useRouter()
const userMenus = JSON.parse(localStorage.getItem('menuData'))


const routes = mapMenusToRoutes(userMenus)
const route = useRoute()

const defaultActive = computed(() => {
  const pathMenu = mapPathToMenu(route.path, userMenus)
  // 如果找不到匹配的菜单，返回一个空字符串（ElMenu 会不高亮任何项）
  return pathMenu?.id ? pathMenu.id + '' : ''
})

const handleItemClick = (item) => { 
  router.push(item.url)
}
</script>

<style scoped lang="scss">

.sidebar-container {
  height: 100%;
  background-color: #153b52;
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
}

.headMenuIcon {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  padding: 20px 0;
  flex-shrink: 0;
  .iconfont.icon-xianluomao {
    font-size: 28px;
    color: #fff;
  }
  h5 {
    color: #fff;
    margin-top: 8px;
    font-size: 16px;
    font-weight: normal;
  }
}

.menu-wrapper {
  flex: 1;
  overflow-y: auto;
  border-right: none;
  background-color: #153b52;
  width: 100%;

 &.el-tooltip__trigger{
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: #fff;
 }

  &::-webkit-scrollbar {
    width: 4px;
  }
  &::-webkit-scrollbar-thumb {
    background-color: rgba(255,255,255,0.3);
    border-radius: 2px;
  }
}


</style>