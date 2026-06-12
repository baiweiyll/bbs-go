// 百度统计 - 仅生产环境加载，包含 UrlChangeTracker 插件用于 SPA 路由追踪
export default defineNuxtPlugin(() => {
  // 生产环境加载百度统计
  if (import.meta.env.PROD) {
    // @ts-expect-error 禁用ts检查
    window._hmt = window._hmt || [];
    // @ts-expect-error 禁用ts检查
    window._hmt.push([
      '_requirePlugin',
      'UrlChangeTracker',
      {
        shouldTrackUrlChange: function (newPath: string, oldPath: string) {
          return newPath && oldPath;
        },
      },
    ]);

    const hm = document.createElement('script');
    hm.src = 'https://hm.baidu.com/hm.js?30c4c4ae8629ab5be34e5d3d3aca51d6';
    const s = document.getElementsByTagName('script')[0];
    s.parentNode!.insertBefore(hm, s);
  }
});
