// src/store/device.js
import { defineStore } from 'pinia'
import { getDeviceList } from '@/api/device/deviceReq.js'
import { delDevice } from '@/api/device/deviceDel.js'



export const useDeviceStore = defineStore('device', {
  state: () => ({
    devicePage: {},
    deviceList: []
  }),
  actions: {
    //获取设备列表
   async fetchDeviceList(params) {
      const list = await getDeviceList(params)
      this.deviceList = list.data.list;
      this.devicePage = list.data.page;
    },
    //删除设备
    async delDevice(data) {
      await delDevice(data)
    }
  }
})

export default useDeviceStore