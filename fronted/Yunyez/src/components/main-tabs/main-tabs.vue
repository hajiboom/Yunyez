<template>
  <el-tabs
    v-model="activeTabPath"
    type="card"
    closable
    class="tab-bar"
    @tab-remove="handleRemove"
  >
    <el-tab-pane
      v-for="tab in visitedViews"
      :key="tab.path"
      :name="tab.path"
      :label="tab.title"
    />
  </el-tabs>
</template>

<script setup>
import { computed } from 'vue'
import { useTabsStore } from '@/store/tabs'

const tabsStore = useTabsStore()

// 从 store 获取标签列表
const visitedViews = computed(() => tabsStore.visitedViews)
console.log(visitedViews.value,"visitedViews");

// 双向绑定当前激活的标签路径
const activeTabPath = computed({
  get: () => tabsStore.activeTabPath,
  set: (path) => tabsStore.setActiveTab(path)
})

// 关闭标签
const handleRemove = (path) => {
  tabsStore.removeView(path)
}
</script>

<style scoped lang="scss">
.tab-bar {
  // 按需调整样式，不设 padding 或内容区高度
  :deep(.el-tabs__content) {
    display: none; /* 内容区完全隐藏，由外部 router-view 渲染 */
  }
}
</style>