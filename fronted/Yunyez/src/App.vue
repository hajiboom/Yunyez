<script setup>
import { ElConfigProvider } from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
</script>

<template>
  <el-config-provider :locale="zhCn">
    <!-- 你的页面内容，比如 router-view -->
    <router-view />
  </el-config-provider>
</template>

<style scoped>
</style>
