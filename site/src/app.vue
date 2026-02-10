<template>
  <NuxtLayout>
    <NuxtLoadingIndicator color="#ea6f5acc" />
    <NuxtPage />
  </NuxtLayout>
</template>

<script setup>
// 应用启动时初始化用户状态
onMounted(() => {
  if (process.client) {
    // 确保在客户端环境中初始化
    const { $initUserState } = useNuxtApp()
    if ($initUserState) {
      $initUserState()
    }
  }
})
</script>
