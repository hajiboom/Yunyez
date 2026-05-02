import {useLoginStore} from '@/store/login'

function usePermissions(permissionID) {
  const loginStore = useLoginStore()
  //获取到登录后返回的权限数组
  const { permissions } = loginStore
  //获取到的数据是字符串数组如：system:users:create
  //检查permissions数组中是否有元素包含permissionID，存在则返回true，否则返回false。
  return !!permissions.find((item) => item.includes(permissionID))
}
export default usePermissions