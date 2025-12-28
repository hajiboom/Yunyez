<template>
  <div class="deviceContent">
      <div class="deviceTitle">
        <div class="headTitle">
        <h3>设备管理</h3>
        <span>管理所有云也子AI机器人设备</span>
        </div>
          <button >+ 添加设备</button>
      </div>
      <div class="formPart">
          <el-input  v-model="input" style="width: 100%" placeholder="请输入设备序列号" >
             <template #prefix>
          <!-- 直接使用导入的Search图标，可自定义size -->
          <el-icon ><Search /></el-icon>
        </template>
            </el-input>
           <el-table :data="deviceList" stripe style="width: 100%;">
    <el-table-column prop=" sn" label="设备ID" width="180" />
    <el-table-column prop=" deviceType" label="设备类型" width="180" />
    <el-table-column prop="vendorName" label="供应商" width="180"/>
    <el-table-column prop="productModel" label="产品型号" width="180"/>
    <el-table-column prop=" status" label="状态" width="180"/>
    <el-table-column prop=" createTime" label="注册时间" width="180"/>
 <!-- 2. 操作列：通过作用域插槽获取当前行数据 -->
  <el-table-column label="操作" width="180">
    <!-- #default="scope" 是Element Plus的作用域插槽 -->
    <template #default="scope">
      <el-button type="primary" size="mini">编辑</el-button>
      <el-button type="danger" size="mini" @click="handelDelDevice(scope.row.sn)">删除</el-button>
    </template>
  </el-table-column>
  </el-table>
      </div>
  </div>
</template>
<script setup>
  import { ref,onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'
import deviceStore from '@/store/deviceStore.js'
import { storeToRefs } from 'pinia'

//获取列表数据
const UseDeviceStore = deviceStore()
const { deviceList } = storeToRefs(UseDeviceStore)

// 组件挂载时调用获取列表数据
onMounted(() => {
  //传入参数
  UseDeviceStore.fetchDeviceList({
    pageNum: 1,
    pageSize: 5,
  })
})
const input = ref('')

//删除设备
async function handelDelDevice(sn) {
  await UseDeviceStore.delDevice({sn})
  //删除成功后刷新列表
  UseDeviceStore.fetchDeviceList({
    pageNum: 1,
    pageSize: 5,
  })
}
</script>
<style scoped lang="scss">
  $blue-color:#1729f1;
.deviceTitle{
  display: flex;
  justify-content: space-between;
  align-items: center;
  .headTitle{
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: flex-start;
    span{
      font-size: 14px;
      color: #999;
      margin-top: 8px;
    }
  }
  button{
      padding: 8px 16px;
      background-color: $blue-color;
      border: none;
      color: #fff;
      font-size: 14px;
      font-weight: 500;
      cursor: pointer;
      border-radius: 5px;
      border: none;
  }
  
}
.formPart{
      margin-top: 20px;
      background-color: #ffffff;
      padding: 1%;
      border-radius: 10px;
  :deep(.el-input__wrapper){
    padding: 5px 10px;
  }
  }
</style>