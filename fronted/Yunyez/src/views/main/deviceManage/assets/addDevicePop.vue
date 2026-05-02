<template>
  <el-dialog v-model="dialogVisible" title="添加设备" width="800" draggable>
    <el-form :inline="true" :model="formInline" class="demo-form-inline" label-width="100">
    <el-form-item label="设备序列号">
      <el-input v-model="formInline.user" placeholder="请输入设备序列号" clearable />
    </el-form-item>
    <el-form-item label="设备类型" >
      <el-select
        v-model="formInline.region"
        placeholder="请选择设备类型"
        clearable
      >
        <el-option label="Zone one" value="shanghai" />
        <el-option label="Zone two" value="beijing" />
      </el-select>
    </el-form-item>
   <el-form-item label="供应商">
      <el-select
        v-model="formInline.region"
        placeholder="请选择供应商"
        clearable
      >
        <el-option label="Zone one" value="shanghai" />
        <el-option label="Zone two" value="beijing" />
      </el-select>
    </el-form-item>
    <el-form-item label="产品型号">
      <el-select
        v-model="formInline.region"
        placeholder="请选择产品型号"
        clearable
      >
        <el-option label="Zone one" value="shanghai" />
        <el-option label="Zone two" value="beijing" />
      </el-select>
    </el-form-item>
    <el-form-item label="状态">
      <el-select
        v-model="formInline.region"
        placeholder="请选择状态"
        clearable
      >
        <el-option label="Zone one" value="shanghai" />
        <el-option label="Zone two" value="beijing" />
      </el-select>
    </el-form-item>
  </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="dialogVisible = false">
          确认添加
        </el-button>
      </div>
    </template>
  </el-dialog>


</template>

<script setup>
import { ref, computed,reactive} from 'vue'


const props = defineProps({
  moduleVisible: {
    type: Boolean,
    default: false
  }
})
const emit = defineEmits(['update:moduleVisible'])
const dialogVisible = computed({
  get: () => props.moduleVisible,
  set: (val) => {
    if (!val) emit('update:moduleVisible', false)
  }
})

const formInline = reactive({
  user: '',
  region: '',
  date: '',
})




</script>

<style scoped lang="scss">
:global(.custom-dragging-style.is-dragging) {
  border: 2px dashed var(--el-color-primary);
  opacity: 0.65;
}
.demo-form-inline .el-input {
  --el-input-width: 220px;
}

.demo-form-inline .el-select {
  --el-select-width: 220px;
}
</style>
