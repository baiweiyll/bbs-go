<template>
  <div>
    <!-- 页面内容为空，仅用于处理回调逻辑 -->
  </div>
</template>
  
<script setup lang="ts">
import { Base64 } from 'js-base64'
import { useRouter } from '#app'

// 使用 Nuxt 3 的路由
const router = useRouter()

// 获取 URL 参数
function getUrlParam(name: string): string | null {
  if (typeof window === 'undefined') return null

  const urlParams = new URLSearchParams(window.location.search)
  console.log('URL Parameters:', urlParams)
  return urlParams.get(name)
}

// 处理登录回调逻辑
const processLoginCallback = () => {
  // 从 URL 参数中获取数据
  const loginData = getUrlParam('data')

  if (!loginData) {
    console.error('未找到登录数据')
    router.push('/')
    return
  }

  try {
    // Base64 解码数据
    const decodedData = Base64.decode(loginData)

    // 解析用户信息
    const userInfo = JSON.parse(decodedData)?.user || null

    // 更新 Pinia store 状态
    const userStore = useUserStore()
    userStore.setUser(userInfo)
    
    console.log('Console - User state set:', userStore.user);

    // 使用 setTimeout 确保状态完全更新后再跳转
    setTimeout(() => {
      console.log('Console - About to navigate, user state:', userStore.user);
      router.push(`/user/${userInfo.id}`)
    }, 200)
  } catch (error) {
    console.error('处理登录回调失败:', error)
    router.push('/user/signin')
  }
}

// 在组件挂载时执行
onMounted(() => {
  processLoginCallback()
})
</script>