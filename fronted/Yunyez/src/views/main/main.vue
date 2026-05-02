<template>
  <div class="common-layout">
    <el-container style="height: 100vh;">
      <el-aside :width="isFold ? '60px' : '200px'" style="transition: width 0.3s ease; overflow: hidden;">
        <main-menu :isFold="isFold"></main-menu>
      </el-aside>
      <el-container>
        <el-header>
          <main-header @fold-change="handleFoldChange"></main-header>
        </el-header>
        <MainTabs />
        <el-main>
          <router-view v-slot="{ Component }">
            <keep-alive :include="cachedViewNames">
              <component :is="Component" />
            </keep-alive>
          </router-view>
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import MainMenu from '@/components/main-menu/main-menu.vue'
import MainTabs from '@/components/main-tabs/main-tabs.vue'
import MainHeader from '@/components/main-header/main-header.vue'
import { useTabsStore } from '@/store/tabs'

const tabsStore = useTabsStore()

const isFold = ref(false)
const handleFoldChange = (flag) => {
  isFold.value = flag
}
const cachedViewNames = computed(() => {
  return tabsStore.visitedViews.map(v => v.name).filter(Boolean)
})

</script>

<style scoped lang="scss">

</style>