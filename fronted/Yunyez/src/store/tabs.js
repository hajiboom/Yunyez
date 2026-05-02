
import { defineStore } from 'pinia'
import router from '@/router'

export const useTabsStore = defineStore('tabs', {
  state: () => ({
    visitedViews: [],        // { path, title, name, query, params }
    activeTabPath: '',
  }),
  actions: {
    // 添加标签（路由守卫调用）
    addView(view) {
      const { path, title, name, query, params } = view
      const exist = this.visitedViews.find(v => v.path === path)
      if (!exist) {
        this.visitedViews.push({ path, title: title || name || '未命名', query, params })
      }
      this.activeTabPath = path
    },
    // 移除标签
    removeView(path) {
      const index = this.visitedViews.findIndex(v => v.path === path)
      if (index === -1) return
      // 如果关闭的是当前激活的标签，需要切换到相邻标签
      if (path === this.activeTabPath) {
        const nextTab = this.visitedViews[index + 1] || this.visitedViews[index - 1]
        if (nextTab) {
          this.activeTabPath = nextTab.path
          router.push(nextTab.path)
        } else {
          // 如果删完了，跳转到首页（根据业务处理）
          router.push('/')
          this.activeTabPath = ''
        }
      }
      this.visitedViews.splice(index, 1)
    },
    // 点击标签切换
    setActiveTab(path) {
      this.activeTabPath = path
      router.push(path)
    },
    // 清空所有标签
    clearTabs() {
      this.visitedViews = []
      this.activeTabPath = ''
    },
    
  },
  // 可选：持久化
  // persist: true
})