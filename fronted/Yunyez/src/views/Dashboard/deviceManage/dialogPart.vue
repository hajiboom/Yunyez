<template>
  <el-dialog v-model="dialogVisible" :title="props.data.title" width="500px" draggable="true">
    <!-- 设备更新表单 -->
    <el-form v-if="data.mode === 'edit'" :model="form" label-width="100px" style="padding: 0 20px">

      <el-form-item label="设备序列号">
        <el-input v-model="form.sn" placeholder="设备序列号" />
      </el-form-item>

      <el-form-item label="固件版本号">
        <el-input v-model="form.firmwareVersion" placeholder="请输入固件版本" />
      </el-form-item>

      <el-form-item label="质保时间">
        <el-input v-model="form.ExpireDate" placeholder="请输入质保时间" />
      </el-form-item>

      <el-form-item label="注册时间">
        <el-date-picker v-model="form.createTime" type="datetime" format="YYYY-MM-DD HH:mm:ss"
          value-format="YYYY-MM-DD HH:mm:ss" placeholder="请选择注册时间" style="width: 100%;" />
      </el-form-item>

      <!-- 状态：下拉选择框（匹配业务常用状态） -->
      <el-form-item label="设备状态">
        <el-select v-model="form.status" placeholder="请选择设备状态">
          <el-option label="已激活" value="activated" />
          <el-option label="未激活" value="inactivated" />
          <el-option label="已禁用" value="disabled" />
          <el-option label="已报废" value="scrapped" />
        </el-select>
      </el-form-item>
      <el-form-item label="设备备注">
        <el-input v-model="form.Remark" placeholder="设备备注" />
      </el-form-item>
    </el-form>
    <el-form v-if="data.mode === 'add'" :model="form" label-width="100px" style="padding: 0 20px">

      <el-form-item label="设备序列号">
        <el-input v-model="form.sn" placeholder="设备序列号" />
      </el-form-item>

      <el-form-item label="设备类型">
        <el-input v-model="form.deviceType" placeholder="请输入设备类型" />
      </el-form-item>

      <el-form-item label="供应商">
        <el-input v-model="form.vendorName" placeholder="请输入供应商" />
      </el-form-item>

      <el-form-item label="产品型号">
        <el-input v-model="form.productModel" placeholder="请输入产品型号" />
      </el-form-item>
      <el-form-item label="注册时间">
        <el-date-picker v-model="form.createTime" type="datetime" format="YYYY-MM-DD HH:mm:ss"
          value-format="YYYY-MM-DD HH:mm:ss" placeholder="请选择注册时间" style="width: 100%;" />
      </el-form-item>
      <!-- 状态：下拉选择框（匹配业务常用状态） -->
      <el-form-item label="设备状态">
        <el-select v-model="form.status" placeholder="请选择设备状态">
          <el-option label="已激活" value="activated" />
          <el-option label="未激活" value="inactivated" />
          <el-option label="已禁用" value="disabled" />
          <el-option label="已报废" value="scrapped" />
        </el-select>
      </el-form-item>
    </el-form>
    <!-- 底部按钮 -->
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleCancel">取消</el-button>
        <el-button type="primary" @click="handleConfirm">确认提交</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, watch, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { formatDateTime } from '@/utils/formatTime.js'
import deviceStore from '@/store/deviceStore.js'

const useDeviceStore = deviceStore()
// 1. 定义Props：接收父组件传递的「显隐状态」和「当前设备数据」
const props = defineProps({
  // 对话框显隐（v-model绑定）
  modelValue: {
    type: Boolean,
    default: false
  },
  // 父组件传递的操作类型（新增/编辑）
  data: {
    type: Object,
    default: () => ({})
  }
})

// 2. 定义Emits：向父组件发送事件（关闭、提交更新）
const emit = defineEmits(['update:modelValue', 'confirmUpdate'])

// 3. 响应式表单数据（匹配表格字段）
const form = reactive({
  sn: '',
})

// 4. 对话框显隐状态（适配v-model）
const dialogVisible = ref(props.modelValue)
// 监听props的modelValue变化，同步对话框显隐
watch(() => props.modelValue, (val) => {
  dialogVisible.value = val
  const handleTimeData = ref({})
  //编辑状态下，处理时间格式，回显数据
  if (props.data.mode == "edit") {
    //处理时间格式
    Object.assign(handleTimeData.value, props.data.currentEditDevice
    )
    console.log(props.data);

    handleTimeData.value.createTime = formatDateTime(props.data.currentEditDevice
      .createTime)
    handleTimeData.value.ExpireDate = formatDateTime(props.data.currentEditDevice
      .expireDate)

    // 当对话框打开且有设备数据时，回显表单
    if (val && props.data.currentEditDevice
    ) {
      Object.assign(form, handleTimeData.value)
    }
  }
})
// 监听对话框显隐，同步给父组件
watch(dialogVisible, (val) => {
  emit('update:modelValue', val)
})



// 6. 取消按钮逻辑
const handleCancel = () => {
  dialogVisible.value = false
  resetForm()
}

// 7. 确认更新按钮逻辑
const handleConfirm = async () => {
  // 基础校验：至少保证序列号存在（必填）
  if (!form.sn) {
    ElMessage.warning('设备序列号不能为空！')
    return
  }
  if (props.data.mode == "add") {

    //还未实现接口
    ElMessage.success('暂时未实现新增设备功能！')
    // 关闭对话框并清空表单
    dialogVisible.value = false
    resetForm()
  }
  if (props.data.mode == "edit") {
    await useDeviceStore.updateDevice(form)
    // 向父组件发送「确认更新」事件
    emit('confirmUpdate')
    // 关闭对话框并清空表单
    dialogVisible.value = false
    resetForm()
  }
}
// 辅助方法：重置表单
const resetForm = () => {
  form.sn = ''
  form.deviceType = ''
  form.vendorName = ''
  form.productModel = ''
  form.status = ''
  form.createTime = ''
}
</script>

<style scoped>
.dialog-footer {
  text-align: right;
}

.el-form {
  margin-top: 10px;
}
</style>
