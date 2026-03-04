// src/store/device.js
import { defineStore } from 'pinia'
import { getDeviceList } from '@/api/device/deviceReq.js'
import { delDevice } from '@/api/device/deviceDel.js'
import { updateDevice } from '@/api/device/deviceUpdate.js'
import { addDevice } from '@/api/device/deviceAdd.js'



export const useDeviceStore = defineStore('device', {
  state: () => ({
    devicePage: {},
    deviceList: [],
  }),
  actions: {
    //添加设备
    async addDevice(data) {
      await addDevice(data)
    },
    //获取设备列表
   async fetchDeviceList(params) {
      const list = await getDeviceList(params)
      this.deviceList = list.Data.list;
      this.devicePage = list.Data.page;
    },
    //获取设备信息
    async fetchDevice(params) {      
      const list = await getDeviceList(params);
      // 单个对象 → 转成数组，匹配deviceList的数组类型
    this.deviceList = list.Data ? [list.Data] : []; 
    if (!list.Data) {
      ElMessage.warning('未查询到该设备信息');
    }
    },
    //删除设备
    async delDevice(data) {
      await delDevice(data)
    },
    //更新设备
    async updateDevice(data) {
      const res = await updateDevice({
      sn: 'SN123456789', // 必传：设备序列号
      firmwareVersion: 'V2.1.0', // 可选：固件版本号
      Status: 'activated', // 可选：设备状态（枚举值选一个：disabled/scrapped/activated/inactivated）
      ExpireDate: '2025-12-31', // 可选：质保时间
      ActivationTime: '2024-01-03', // 可选：激活时间
      Remark: '设备运行正常' // 可选：备注
    })
    },
  }
})

export default useDeviceStore