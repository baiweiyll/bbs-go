<template>
  <div>
    <!-- 页面内容为空，仅用于处理回调逻辑 -->
  </div>
</template>
  
<script setup lang="ts">
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
  const code = getUrlParam('code')
  const state = getUrlParam('state')

  console.log('Code:', code, 'State:', state)

  if (!code) {
    console.error('未找到登录数据')
    router.push('/')
    return
  } else {
    // 跳转到 OIDC 回调接口，通过 Nuxt 代理转发到后端
    window.location.href = "/oidc/login/callback?code=" + code + "&state=" + state;
  }

  // try {
  //   // Base64 解码数据
  //   const decodedData = Base64.decode(loginData)

  //   // 解析用户信息
  //   const userInfo = JSON.parse(decodedData)

  //   // 保存到 localStorage
  //   if (typeof localStorage !== 'undefined') {
  //     localStorage.setItem('userInfo', JSON.stringify(userInfo))

  //     // 如果使用了 Pinia，可以在这里更新状态
  //     // 例如：const authStore = useAuthStore()
  //     // authStore.setUser(userInfo)
  //   }

  //   // 如果使用了 Nuxt 的 useState 进行状态管理
  //   const userState = useState('user', () => userInfo)

  //   // 跳转到首页
  //   router.push('/')

  // } catch (error) {
  //   console.error('处理登录回调失败:', error)
  //   router.push('/user/signin')
  // }
}

// 在组件挂载时执行
onMounted(() => {
  processLoginCallback()
})
</script>