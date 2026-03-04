<template>
  <el-dialog v-model="dialogVoiceVisible" :title="语音连接" width="800px" draggable="true">
    
  </el-dialog>
</template>

<script setup>
import { ref, watch, reactive } from 'vue'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: () => false
  }
})
const dialogVoiceVisible = ref(props.modelValue)
const emit  = defineEmits(['update:modelValue'])
watch(()=>dialogVoiceVisible.value, (newVal) => {
  emit('update:modelValue', newVal)
})
watch(()=>props.modelValue, (newVal) => {
  dialogVoiceVisible.value = newVal
  emit('update:modelValue', newVal)
})
</script>

<style scoped>
.dialog-footer {
  text-align: right;
}

</style>
