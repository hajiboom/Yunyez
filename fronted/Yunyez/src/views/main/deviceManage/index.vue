<template>
  <div class="deviceContent">
    <div class="deviceTitle">
      <div class="headTitle">
        <h3>设备管理</h3>
        <span>管理所有云也子AI机器人设备</span>
      </div>
      <button @click="moduleVisible = true" v-if="isCreate">+ 添加设备</button>
    </div>
    <div class="formPart">
      <el-input
        v-model="input"
        style="width: 100%"
        placeholder="请输入设备序列号"
        v-if="isQuery"
      >
        <template #prefix>
          <!-- 直接使用导入的Search图标，可自定义size -->
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      <el-table :data="deviceList" stripe style="width: 100%">
        <el-table-column prop="sn" label="设备序列号" width="380" />
        <el-table-column prop="deviceType" label="设备类型" width="180" />
        <el-table-column prop="vendorName" label="供应商" width="180" />
        <el-table-column prop="productModel" label="产品型号" width="180" />
        <el-table-column label="状态" width="180">
          <template #default="scope">
            <el-tag :type="scope.row.status === 'activated' ? 'success' : 'danger'">
              {{ scope.row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="注册时间" width="180">
          <template #default="scope">
            {{ formatDate(scope.row.createTime) }}
          </template>
        </el-table-column>
        <!-- 2. 操作列：通过作用域插槽获取当前行数据 -->
        <el-table-column label="操作" width="180">
          <!-- #default="scope" 是Element Plus的作用域插槽 -->
          <template #default="scope">
            <el-button type="primary" size="small" v-if="isUpdate">编辑</el-button>
            <el-button
              type="danger"
              size="small"
              @click="handelDelDevice(scope.row.sn)"
              v-if="isDelete"
              >删除</el-button
            >
          </template>
        </el-table-column>
      </el-table>
    </div>
    <addDevicePop
      :moduleVisible="moduleVisible"
      @update:moduleVisible="moduleVisible = false"
    />
    <div class="pagination" style="width: 100%;display: flex;justify-content: center;">
      <el-pagination
  :current-page="currentPage"
  :page-size="pageSize"
  :total="total"
  :pager-count="7"
  layout="prev, pager, next, sizes"
  :page-sizes="[8, 16, 32, 64]"
  @current-change="handleCurrentChange"
  @size-change="handleSizeChange"
/>
    </div>
   
  </div>
</template>
<script setup>
import { ref, onMounted } from "vue";
import { Search } from "@element-plus/icons-vue";
import deviceStore from "@/store/deviceStore.js";
import { storeToRefs } from "pinia";
import dayjs from "dayjs";
import addDevicePop from "./assets/addDevicePop.vue";
import usePermissions from "@/hooks/usePermissions.js";

// 0.判断是否有增删改查的权限：'user:create'
const isCreate = usePermissions('device:create')
const isDelete = usePermissions('device:delete')
const isUpdate = usePermissions('device:update')
const isQuery = usePermissions('device:query')

const moduleVisible = ref(false);
//获取列表数据
const UseDeviceStore = deviceStore();
const { deviceList } = storeToRefs(UseDeviceStore);
// 分页相关变量
const currentPage = ref(1);
const pageSize = ref(8);
const total = ref(0);

const formatDate = (time) => {
  if (!time) return "-";
  return dayjs(time).format("YYYY-MM-DD HH:mm:ss");
};
const input = ref("");
// 获取设备列表（带分页和搜索）
const fetchData = async () => {
  const params = {
    pageNum: currentPage.value,
    pageSize: pageSize.value,
    sn: input.value || undefined,   // 如果搜索框有内容则传入
  };
  await UseDeviceStore.fetchDeviceList(params);
  // 假设 store 中保存了 total 字段，从 storeToRefs 中解构
  // 如果 store 没有 total，需要修改 store 保存 total
  total.value = UseDeviceStore.total || 0; 
};


// 监听页码变化
const handleCurrentChange = (page) => {
  currentPage.value = page;
  fetchData();
};

// 监听每页条数变化
const handleSizeChange = (size) => {
  pageSize.value = size;
  currentPage.value = 1; // 重置到第一页
  fetchData();
};

// 搜索处理（可以添加防抖）
const handleSearch = () => {
  currentPage.value = 1;
  fetchData();
};

// 组件挂载
onMounted(() => {
  fetchData();
});


//删除设备
async function handelDelDevice(sn) {
  await UseDeviceStore.delDevice({ sn });
  //删除成功后刷新列表
  UseDeviceStore.fetchDeviceList({
    pageNum: 1,
    pageSize: 5,
  });
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
    border: none;
  }
}
.formPart {
  margin-top: 20px;
  background-color: #ffffff;
  padding: 1%;
  border-radius: 10px;
  :deep(.el-input__wrapper) {
    padding: 5px 10px;
  }
}
</style>