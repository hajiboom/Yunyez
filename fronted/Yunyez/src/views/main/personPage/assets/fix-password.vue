<template>
  <el-dialog
    v-model="dialogVisible"
    title="修改密码"
    width="480px"
    :close-on-click-modal="false"
    :close-on-press-escape="true"
    @close="closeModal"
  >
    <el-form ref="formRef" :model="pwdForm" label-position="top" :rules="rules">
      <el-form-item label="当前密码" prop="oldPassword">
        <el-input
          type="password"
          v-model="pwdForm.oldPassword"
          placeholder="请输入当前密码"
          show-password
        />
      </el-form-item>
      <el-form-item label="新密码" prop="newPassword">
        <el-input
          type="password"
          v-model="pwdForm.newPassword"
          placeholder="请输入新密码"
          show-password
        />
      </el-form-item>
      <el-form-item label="确认新密码" prop="confirmPassword">
        <el-input
          type="password"
          v-model="pwdForm.confirmPassword"
          placeholder="请再次输入新密码"
          show-password
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="closeModal">取消</el-button>
      <el-button type="primary" @click="submitPassword">确认修改</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { reactive, computed,ref} from 'vue'
import {ERROR_CODES} from '@/utils/codes'
import {useLoginStore} from '@/store/login'
import {encryptRsa} from '@/utils/encrypt'



const props = defineProps({
  modalVisible: {
    type: Boolean,
    default: false
  }
})

const formRef = ref(null)
const emit = defineEmits(['closeModal'])
const loginStore = useLoginStore()

// 将 prop 转为 dialog 可用的 v-model 值（只读）
const dialogVisible = computed({
  get: () => props.modalVisible,
  set: (val) => {
    if (!val) emit('closeModal')
  }
})

// 密码表单数据
const pwdForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

// 确认密码是否与新密码一致
const validateConfirmPassword = (rule, value, callback) => {
  if (value === '') {
    callback(new Error('请再次输入新密码'))
  } else if (value !== pwdForm.newPassword) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const rules = reactive({
  oldPassword: [ { required: true, message: '请输入当前密码', trigger: 'blur' },
    { min: 6, message: '密码长度不少于 6 个字符', trigger: 'blur' }],
  newPassword: [ { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度不少于 6 个字符', trigger: 'blur' }],
  confirmPassword: [ { required: true, message: '请输入确认新密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' } ]
})
// 关闭弹窗
const closeModal = () => {
  emit('closeModal')
   pwdForm.oldPassword = ''
      pwdForm.newPassword = ''
      pwdForm.confirmPassword = ''
}

// 提交修改密码
const submitPassword = async () => {
  if (!formRef.value) return
  
  // 调用表单校验
  try {
   const valid =  await formRef.value.validate()   // 校验通过则继续，不通过会抛出异常
    if (!valid) return
    
    // 校验通过，执行提交逻辑
    //加密数据
    const transportData = {
      oldPassword:  encryptRsa(pwdForm.oldPassword),
      newPassword:  encryptRsa(pwdForm.newPassword),
      confirmPassword:  encryptRsa(pwdForm.confirmPassword)
    }
    const res = await loginStore.fixPassword(transportData)
    if (res.code === ERROR_CODES.SUCCESS) {
      ElMessage.success('密码修改成功')
      //清空表单数据
      pwdForm.oldPassword = ''
      pwdForm.newPassword = ''
      pwdForm.confirmPassword = ''
    } else {
      ElMessage.error(res.msg || '密码修改失败')
       //清空表单数据
      pwdForm.oldPassword = ''
      pwdForm.newPassword = ''
      pwdForm.confirmPassword = ''
    }
    closeModal()
  } catch (error) {
    // 校验失败，Element Plus 会自动显示错误信息，无需额外处理
    console.log('表单校验失败', error)
  }
}
</script>

<style scoped lang="scss">
// 可保留一些微调样式，无需覆盖太多，Element Plus 自带样式已足够
:deep(.el-dialog__body) {
  padding-top: 0;
}
:deep(.el-form-item) {
  margin-bottom: 20px;
}
</style>