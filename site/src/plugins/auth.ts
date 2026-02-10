export default defineNuxtPlugin((nuxtApp) => {
  const userStore = useUserStore()
  
  // 应用启动时从 localStorage 恢复用户状态
  const initUserState = () => {
    if (process.client && typeof localStorage !== 'undefined') {
      try {
        const storedUserInfo = localStorage.getItem('userInfo')
        console.log('Stored user info:', storedUserInfo)
        if (storedUserInfo) {
          const userInfo = JSON.parse(storedUserInfo)
          console.log('Restoring user state:', userInfo)
          userStore.setUser(userInfo)
        }
      } catch (error) {
        console.error('恢复用户状态失败:', error)
        // 清除无效的存储数据
        localStorage.removeItem('userInfo')
      }
    }
  }

  // 在客户端初始化时恢复用户状态
  if (process.client) {
    // 使用 nextTick 确保在组件渲染前恢复状态
    nextTick(() => {
      initUserState()
    })
  }

  // 提供全局方法来手动初始化用户状态
  return {
    provide: {
      initUserState
    }
  }
})