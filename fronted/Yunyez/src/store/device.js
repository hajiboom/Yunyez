// src/store/device.js
import { defineStore } from 'pinia'
import { getDeviceListApi } from '@/service/device'
import { ElMessage } from 'element-plus'

// 定义Store：id为唯一标识（deviceStore）
export const useDeviceStore = defineStore('deviceStore', {
  // 1. 状态：存储设备列表、分页、查询条件、加载状态
  state: () => ({
    // 设备列表数据
    deviceList: [],
    // 分页参数（和页面分页器绑定）
    pagination: {
      pageNum: 1,    // 当前页码
      pageSize: 10,  // 每页条数
      total: 0       // 总条数（后端返回）
    },
    // 查询条件（和搜索栏绑定）
    searchForm: {
      sn: '', // 设备序列号
      vendorName: '', // 厂商名称
      status: '', // 设备状态
      createTime: [] // 录入时间
    },
    // 加载状态
    loading: false
  }),

  // 2. 计算属性（可选，比如筛选后的列表）
  getters: {
    // 示例：获取已激活的设备数量
    activatedDeviceCount: (state) => {
      return state.deviceList.filter(item => item.status === 'activated').length
    }
  },

  // 3. 方法：封装业务逻辑（查询设备、重置条件等）
  actions: {
    /**
     * 查询设备列表（封装分页+查询条件）
     */
    async fetchDeviceList() {
      try {
        this.loading = true
        // 构造请求参数（拼接分页+查询条件）
        const params = {
          pageNum: this.pagination.pageNum,
          pageSize: this.pagination.pageSize,
          sn: this.searchForm.sn,
          vendorName: this.searchForm.vendorName,
          status: this.searchForm.status,
          startTime: this.searchForm.createTime[0],
          endTime: this.searchForm.createTime[1]
        }
        console.log('查询参数', params)
        const res = await getDeviceListApi()

        if (res.Code === 200) {
          this.deviceList = res.Data?.list || []
          this.pagination.total = res.Data?.total || 0
        } else {
          ElMessage.error(res.Message || '获取失败')
          this.deviceList = []
          this.pagination.total = 0
        }
      } catch (error) {
        ElMessage.error('设备列表查询失败：' + (error.msg || error.message))
        return res // 包含 Code, Data, Message
      } finally {
        this.loading = false
      }
    },

    /**
     * 重置查询条件
     */
    resetSearchForm() {
      this.searchForm = {
        sn: '',
        vendorName: '',
        status: '',
        createTime: []
      }
      // 重置后默认查第一页
      this.pagination.pageNum = 1
      // 重新查询
      this.fetchDeviceList()
    },

    /**
     * 更新分页参数（页码/页大小）
     * @param {Object} payload - { pageNum?, pageSize? }
     */
    updatePagination(payload) {
      if (payload.pageNum !== undefined) {
        this.pagination.pageNum = payload.pageNum
      }
      if (payload.pageSize !== undefined) {
        this.pagination.pageSize = payload.pageSize
      }
      // 分页变化后重新查询
      this.fetchDeviceList()
    }
  }
})

export default useDeviceStore