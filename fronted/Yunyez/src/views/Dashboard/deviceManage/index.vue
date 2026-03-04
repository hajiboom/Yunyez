<template>
  <div class="deviceContent">
    <div class="deviceTitle">
      <div class="headTitle">
        <h3>设备管理</h3>
        <span>管理所有云也子AI机器人设备</span>
      </div>
      <button @click="handleAddDevice">+ 添加设备</button>
    </div>
    <div class="formPart">
      <el-input
        @keyup.enter="handleInputEnter"
        v-model="input"
        style="width: 100%"
        placeholder="请输入设备序列号"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>

      <!-- 动态绑定所有字段的表格 -->
      <el-table :data="deviceList" stripe style="width: 100%; margin-top: 16px">
        <!-- 遍历完整列配置数组 -->
        <el-table-column
          v-for="column in columns"
          :key="column.prop || column.label"
          :prop="column.prop"
          :label="column.label"
          :width="column.width"
          :align="column.align || 'center'"
        >
          <!-- 自定义列内容插槽 -->
          <template #default="scope">
            <!-- 时间列：格式化ISO时间 -->
            <template v-if="column.type === 'date'">
              {{ formatDateTime(scope.row[column.prop]) }}
            </template>
            <!-- 状态枚举列：中文转换 + 标签样式 -->
            <template v-else-if="column.type === 'status'">
              <el-tag :type="column.tagTypeMap[scope.row[column.prop]]">
                {{ column.enumMap[scope.row[column.prop]] || scope.row[column.prop] }}
              </el-tag>
            </template>
            <!-- 普通枚举列：仅中文转换 -->
            <template v-else-if="column.type === 'enum'">
              {{ column.enumMap[scope.row[column.prop]] || scope.row[column.prop] }}
            </template>
            <!-- 操作列：编辑/删除按钮 -->
            <template v-else-if="column.type === 'operation'">
              <el-button
                type="primary"
                size="small"
                @click="handleEditDevice(scope.row)"
              >
                编辑
              </el-button>
              <el-button
                type="danger"
                size="small"
                @click="handleDelDevice(scope.row.sn)"
              >
                删除
              </el-button>
              <el-button
                type="success"
                size="small"
                @click="handleVoiceConnection(scope.row.sn)"
              >
               语音
              </el-button>
            </template>
            <!-- 普通列：空值显示“-” -->
            <template v-else>
              {{ scope.row[column.prop] || '-' }}
            </template>
          </template>
        </el-table-column>
      </el-table>
    </div>
    <!-- 引入弹窗组件 -->
    <dialog-part
      v-model="dialogVisible"
      @confirmUpdate="handleUpdateDevice"
      :data="data"  
    />
    <dialogVoiceConnection
      v-model="dialogVoiceVisible"  
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'
import deviceStore from '@/store/deviceStore.js'
import { storeToRefs } from 'pinia'
import dialogPart from './dialogPart.vue'
import { formatDateTime } from '@/utils/formatTime.js'
import dialogVoiceConnection from './dialogVoiceConnection.vue'

// 获取pinia仓库 & 响应式设备列表
const UseDeviceStore = deviceStore()
// const { deviceList } = storeToRefs(UseDeviceStore)

// 弹窗显隐 + 编辑设备数据 + 搜索输入框
const dialogVisible = ref(false)
//语音弹窗
const dialogVoiceVisible = ref(false)
const currentEditDevice = ref({})
const input = ref('')
const data = ref({})//传递对话框数据
const columns = ref([
  // 唯一标识类（仅保留存在的sn）
  { label: '设备序列号', prop: 'sn', width: 220, align: 'left' },
  
  // 设备基础信息类（仅保留存在的字段）
  { 
    label: '设备类型', 
    prop: 'deviceType', 
    width: 120, 
    type: 'enum', 
    enumMap: { sensor: '传感器', camera: '摄像头', gateway: '网关', robot: 'AI机器人' } 
  },
  { label: '供应商', prop: 'vendorName', width: 150, align: 'left' },
  { label: '产品型号', prop: 'productModel', width: 150, align: 'left' },
  
  // 时间维度类（新增createTime作为注册时间）
  { label: '注册时间', prop: 'createTime', width: 200, type: 'date' },
  
  // 状态类（仅保留存在的设备状态）
  { 
    label: '设备状态', 
    prop: 'status', 
    width: 120, 
    type: 'status',
    enumMap: { inactivated: '未激活', activated: '已激活', disabled: '已禁用', scrapped: '已报废' },
    tagTypeMap: { inactivated: 'warning', activated: 'success', disabled: 'info', scrapped: 'danger' }
  },
  
  // 操作列（保留）
  { label: '操作', width: 280, type: 'operation' }
])

const deviceList = ref([
    {
      sn: 'AIROBOT2026001', // 设备序列号
      deviceType: 'robot', // 设备类型-枚举
      vendorName: '云也子科技', // 供应商
      productModel: 'YZ-AI-R003', // 产品型号
      createTime: '2026-01-15T09:23:45Z', // 注册时间-ISO格式
      status: 'activated' // 设备状态-枚举
    },
    {
      sn: 'SENSOR2026002',
      deviceType: 'sensor',
      vendorName: '华控传感',
      productModel: 'HK-TS100',
      createTime: '2026-01-16T10:15:30Z',
      status: 'inactivated'
    },
    {
      sn: 'CAMERA2026003',
      deviceType: 'camera',
      vendorName: '海康威视',
      productModel: 'DS-2CD3T46WD-I3',
      createTime: '2026-01-17T14:08:22Z',
      status: 'disabled'
    },
    {
      sn: 'GATEWAY2026004',
      deviceType: 'gateway',
      vendorName: '华为',
      productModel: 'AR509CG-Lc',
      createTime: '2026-01-18T08:50:10Z',
      status: 'activated'
    },
    {
      sn: 'AIROBOT2026005',
      deviceType: 'robot',
      vendorName: '云也子科技',
      productModel: 'YZ-AI-R002',
      createTime: '2026-01-19T16:30:00Z',
      status: 'scrapped'
    },
    {
      sn: 'SENSOR2026006',
      deviceType: 'sensor',
      vendorName: '', // 空值测试-显示"-"
      productModel: 'JY-TH20',
      createTime: '2026-01-20T11:20:55Z',
      status: 'activated'
    },
    {
      sn: 'CAMERA2026007',
      deviceType: 'camera',
      vendorName: '大华技术',
      productModel: null, // null值测试-显示"-"
      createTime: '2026-01-21T15:40:33Z',
      status: 'inactivated'
    },
    {
      sn: 'GATEWAY2026008',
      deviceType: 'gateway',
      vendorName: '中兴',
      productModel: 'ZXWL W8150',
      createTime: '2026-01-22T09:10:28Z',
      status: 'disabled'
    }
  ])
// ===================== 业务逻辑 =====================
// 封装列表请求（复用 + 携带搜索条件）
const fetchDeviceList = (params = { pageNum: 1, pageSize: 5 }) => {
  UseDeviceStore.fetchDeviceList({
    sn: input.value,
    ...params
  })
}

// 组件挂载加载列表
onMounted(() => {
  fetchDeviceList()
  
})

// 删除设备
async function handelDelDevice(sn) {
  try {
    await UseDeviceStore.delDevice({ sn })
    fetchDeviceList()
  } catch (error) {
    console.error('删除设备失败：', error)
  }
}

// 回车搜索
const handleInputEnter = () => {
  fetchDeviceList({ pageNum: 1, pageSize: 5 })
}

// 新增设备
const handleAddDevice = () => {
  currentEditDevice.value = {}
  dialogVisible.value = true
  data.value={
    title: '新增设备'
    ,
    mode: 'add'
  }
}

// 编辑设备
const handleEditDevice = (device) => {
  
  dialogVisible.value = true
  data.value={
    title: '编辑设备'
    ,
    currentEditDevice : { ...device },
    mode: 'edit'
  }
}

// 确认更新设备
const handleUpdateDevice = ()=> {
    //刷新表格数据
    fetchDeviceList()
}
// 编辑语音连接
const handleVoiceConnection = () => {
  dialogVoiceVisible.value = true
}
</script>

<style scoped lang="scss">
$blue-color: #1729f1;

.deviceTitle {
  display: flex;
  justify-content: space-between;
  align-items: center;

  .headTitle {
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: flex-start;

    span {
      font-size: 14px;
      color: #999;
      margin-top: 8px;
    }
  }

  button {
    padding: 8px 16px;
    background-color: $blue-color;
    border: none;
    color: #fff;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    border-radius: 5px;
    transition: background-color 0.2s;

    &:hover {
      background-color: #2a35ab;
    }
  }
}

.formPart {
  margin-top: 20px;
  background-color: #ffffff;
  padding: 20px;
  border-radius: 10px;
  overflow-x: auto; // 列多的时候横向滚动

  :deep(.el-input__wrapper) {
    padding: 5px 10px;
    margin-bottom: 16px;
  }

  // 表格标签样式优化
  :deep(.el-tag) {
    font-size: 12px;
    padding: 2px 8px;
  }
}
</style>