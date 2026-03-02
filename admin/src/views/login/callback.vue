<template>
  <div class="callback-container">
    <a-spin :loading="true" tip="正在处理登录回调...">
      <div class="callback-content">
        <p>正在跳转,请稍候...</p>
      </div>
    </a-spin>
  </div>
</template>

<script setup lang="ts">
  import { useRouter } from 'vue-router';
  import { onMounted } from 'vue';

  // 使用 Vue Router
  const router = useRouter();

  // 获取环境变量中的服务器地址
  const serverURL = import.meta.env.VITE_API_BASE_URL || '';

  // 获取 URL 参数
  function getUrlParam(name: string): string | null {
    if (typeof window === 'undefined') return null;

    const urlParams = new URLSearchParams(window.location.search);
    console.log('URL Parameters:', urlParams);
    return urlParams.get(name);
  }

  // 处理登录回调逻辑
  const processLoginCallback = () => {
    // 从 URL 参数中获取数据
    const code = getUrlParam('code');
    const state = getUrlParam('state');

    console.log('Code:', code, 'State:', state);

    if (!code) {
      console.error('未找到登录数据');
      router.push('/');
    } else {
      // 直接跳转到登录接口,让浏览器处理302重定向
      window.location.href = `${serverURL}/bbsoidc/login/callback?code=${code}&state=${state}`;
    }
  };

  // 在组件挂载时执行
  onMounted(() => {
    processLoginCallback();
  });
</script>

<style lang="less" scoped>
  .callback-container {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100vh;
    background: var(--color-bg-1);

    .callback-content {
      text-align: center;
      padding: 20px;

      p {
        margin-top: 16px;
        color: var(--color-text-2);
        font-size: 14px;
      }
    }
  }
</style>
