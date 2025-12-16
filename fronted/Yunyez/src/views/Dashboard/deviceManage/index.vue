<template>
  <div class="device-manage-container">
    <!-- 搜索筛选栏 -->
    <div class="search-bar">
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="设备序列号">
          <el-input
            v-model="searchForm.sn"
            placeholder="请输入设备序列号"
            clearable
            style="width: 200px"
          />
        </el-form-item>
        <el-form-item label="厂商名称">
          <el-input
            v-model="searchForm.vendorName"
            placeholder="请输入厂商名称"
            clearable
            style="width: 200px"
          />
        </el-form-item>
        <el-form-item label="设备状态">
          <el-select
            v-model="searchForm.status"
            placeholder="请选择状态"
            clearable
            style="width: 150px"
          >
            <el-option
              v-for="(item, key) in deviceStatusMap"
              :key="key"
              :label="item.label"
              :value="key"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="录入时间">
          <el-date-picker
            v-model="searchForm.createTime"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
          <el-button type="success" @click="handleAdd">新增设备</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 设备列表 -->
    <el-table
      :data="deviceList"
      border
      stripe
      v-loading="loading"
      style="width: 100%; margin-top: 10px"
      @selection-change="handleSelectionChange"
    >
      <el-table-column type="selection" width="55" />
      <el-table-column
        prop="sn"
        label="设备序列号"
        min-width="220"
        show-overflow-tooltip
      />
      <el-table-column prop="deviceType" label="设备类型" width="120" align="center">
        <template #default="scope">
          {{ scope.row.deviceType || '未知类型' }}
        </template>
      </el-table-column>
      <el-table-column
        prop="vendorName"
        label="厂商名称"
        width="150"
        align="center"
      />
      <el-table-column
        prop="productModel"
        label="产品型号"
        width="150"
        align="center"
      />
      <el-table-column prop="status" label="设备状态" width="120" align="center">
        <template #default="scope">
          <el-tag
            :type="deviceStatusMap[scope.row.status]?.color || 'default'"
            size="small"
          >
            {{ deviceStatusMap[scope.row.status]?.label || '未知状态' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column
        prop="createTime"
        label="录入时间"
        width="200"
        align="center"
      >
        <template #default="scope">
          {{ formatTime(scope.row.createTime) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" align="center">
        <template #default="scope">
          <el-button
          class="contral-btn"
            type="primary"
            size="small"
            @click="handleDetail(scope.row)"
          >
            详情
          </el-button>
          <el-button
          class="contral-btn"
            type="warning"
            size="small"
           
            @click="handleEdit(scope.row)"
          >
            编辑
          </el-button>
          <el-button
          class="contral-btn"
            type="danger"
            size="small"
           
            @click="handleDelete(scope.row)"
          >
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页器 -->
    <div class="pagination-container">
      <el-pagination
        v-model:current-page="pagination.pageNum"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        layout="total, sizes, prev, pager, next, jumper"
        :page-sizes="[10, 20, 50, 100]"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import dayjs from 'dayjs'
import { useRouter } from 'vue-router'
import { useDeviceStore } from '@/store/device'
const deviceStore = useDeviceStore()
const router = useRouter()

// 加载状态
const loading = ref(false)

// 设备状态映射表（核心：将英文状态转为中文+样式）
const deviceStatusMap = reactive({
  activated: { label: '已激活', color: 'success' },
  inactivated: { label: '未激活', color: 'warning' },
  offline: { label: '离线', color: 'danger' },
  online: { label: '在线', color: 'success' },
  disabled: { label: '禁用', color: 'info' },
})

// 搜索表单
const searchForm = reactive({
  sn: '',
  vendorName: '',
  status: '',
  createTime: [],
})

// 分页参数
const pagination = reactive({
  pageNum: 1,
  pageSize: 10,
  total: 0,
})
const deviceList = ref([])
// 3. 核心：查询设备列表（封装成通用方法，供查询/重置/分页调用）
const getDeviceList = async () => {
  try {
    loading.value = true
    // 构造请求参数：搜索条件 + 分页参数
    const params = {
      // 分页参数
      pageNum: pagination.pageNum,
      pageSize: pagination.pageSize,
      // 搜索条件
      sn: searchForm.sn.trim(), // 去除首尾空格
      vendorName: searchForm.vendorName.trim(),
      status: searchForm.status,
      // 时间范围拆分成startTime/endTime（适配后端接口）
      startTime: searchForm.createTime[0] || '',
      endTime: searchForm.createTime[1] || '',
    }
    // 调用封装的接口，获取筛选后的列表数据
    await deviceStore.fetchDeviceList(params)

    pagination.total = deviceStore.pagination.total
    deviceList.value = deviceStore.deviceList

  } catch (error) {
    ElMessage.error('查询失败：' + (error.msg || error.message))
    deviceList.value = [] // 失败时清空列表
    pagination.total = 0
  } finally {
    loading.value = false
  }
}


// 初始化：获取设备列表
onMounted(() => {
  getDeviceList()
})

// 时间格式化：ISO8601 → YYYY-MM-DD HH:mm:ss
const formatTime = (time) => {
  if (!time) return '-'
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

// 搜索
const handleSearch = () => {
  pagination.pageNum = 1 // 搜索重置页码
  getDeviceList()
}

// 重置搜索表单
const handleReset = () => {
  Object.keys(searchForm).forEach(key => {
    if (Array.isArray(searchForm[key])) {
      searchForm[key] = []
    } else {
      searchForm[key] = ''
    }
  })
  handleSearch()
}

// 分页-每页条数变化
const handleSizeChange = (val) => {
  pagination.pageSize = val
  getDeviceList()
}

// 分页-页码变化
const handleCurrentChange = (val) => {
  pagination.pageNum = val
  getDeviceList()
}

// 新增设备
const handleAdd = () => {
  router.push('/deviceManage/add')
}

// 查看设备详情
const handleDetail = (row) => {
  router.push(`/deviceManage/detail/${row.id}`)
}

// 编辑设备
const handleEdit = (row) => {
  router.push(`/deviceManage/edit/${row.id}`)
}

// 删除设备
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除设备【${row.sn}】吗？`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )

    // 实际项目中替换为删除接口请求
    // await deleteDeviceApi(row.id)
    ElMessage.success('删除成功')
    getDeviceList() // 重新获取列表
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败：' + error.message)
    }
  }
}

// 表格多选事件
const handleSelectionChange = (val) => {
  console.log('选中的设备：', val)
  // 可实现批量操作（如批量删除、批量启用/禁用）
}
</script>

<style scoped lang="scss">
.device-manage-container {
  padding: 20px;
  height: 100%;
  box-sizing: border-box;

  .search-bar {
    background: #fff;
    padding: 15px;
    border-radius: 4px;
    box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.04);
  }
  .pagination-container {
    margin-top: 20px;
    text-align: right;
    width:100%;
    display: flex;
    justify-content: center;
  }

}
</style>