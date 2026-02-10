<template>
  <div class="console-container">
    <a-spin :loading="true" tip="正在处理登录...">
      <div class="console-content">
        <p>正在跳转,请稍候...</p>
      </div>
    </a-spin>
  </div>
</template>

<script setup lang="ts">
  import { useRouter } from 'vue-router';
  import { onMounted } from 'vue';
  import { Base64 } from 'js-base64';
  import { useUserStore } from '@/store';
  import { DEFAULT_ROUTE_NAME } from '@/router/constants';

  // 使用 Vue Router
  const router = useRouter();
  const userStore = useUserStore();

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
    const loginData = getUrlParam('data');

    if (!loginData) {
      console.error('未找到登录数据');
      router.push('/');
      return;
    }

    try {
      // Base64 解码数据
      const decodedData = Base64.decode(loginData);

      // 解析用户信息
      const userInfo = JSON.parse(decodedData);

      // 保存到 localStorage
      if (typeof localStorage !== 'undefined') {
        localStorage.setItem('userInfo', JSON.stringify(userInfo));

        // 如果使用了 Pinia,可以在这里更新状态
        // 例如:更新用户store
        if (userStore) {
          // userStore.setInfo(userInfo);
        }
      }

      // 跳转到首页
      router.push({
        name: DEFAULT_ROUTE_NAME,
      });
    } catch (error) {
      console.error('处理登录回调失败:', error);
      router.push('/login');
    }
  };

  // 在组件挂载时执行
  onMounted(() => {
    processLoginCallback();
  });
</script>

<style lang="less" scoped>
  .console-container {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100vh;
    background: var(--color-bg-1);

    .console-content {
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
