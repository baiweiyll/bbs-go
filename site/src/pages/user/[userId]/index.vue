<template>
  <section class="main">
    <div class="container">
      <!-- 只有当 user 数据存在时才渲染 -->
      <template v-if="user">
        <user-profile :user="user" />

        <div class="container main-container right-main size-320">
          <user-center-sidebar :user="user" />
          <div class="right-container">
            <div class="tabs-warp">
              <div class="tabs">
                <ul>
                  <li :class="{ 'is-active': activeTab === 'topics' }">
                    <nuxt-link :to="'/user/' + user.id">
                      <span class="icon is-small">
                        <i class="iconfont icon-topic" aria-hidden="true" />
                      </span>
                      <span>{{ $t("pages.user.topics") }}</span>
                    </nuxt-link>
                  </li>
                  <li :class="{ 'is-active': activeTab === 'articles' }">
                    <nuxt-link :to="'/user/' + user.id + '/articles'">
                      <span class="icon is-small">
                        <i class="iconfont icon-article" aria-hidden="true" />
                      </span>
                      <span>{{ $t("pages.user.articles") }}</span>
                    </nuxt-link>
                  </li>
                </ul>
              </div>

              <load-more-async
                v-slot="{ results }"
                url="/bbsapi/topic/user/topics"
                :params="{ userId: user.id }"
              >
                <topic-list :topics="results" :show-avatar="false" />
              </load-more-async>
            </div>
          </div>
        </div>
      </template>
      
      <!-- 加载状态 -->
      <div v-else class="loading-container">
        <div class="loading">{{ $t("common.loading") }}</div>
      </div>
    </div>
  </section>
</template>

<script setup>
const route = useRoute();
const userStore = useUserStore();

// 优先使用 store 中的当前用户数据（如果是查看自己的页面）
const userId = parseInt(route.params.userId);
const currentUser = userStore.user;

let user;

if (currentUser && currentUser.id === userId) {
  // 如果是查看自己的页面，直接使用 store 数据
  user = currentUser;
} else {
  // 否则从 API 获取用户数据
  user = await useHttpGet(`/bbsapi/user/${route.params.userId}`);
}

const activeTab = ref("topics");
const { t } = useI18n();

// 只有当 user 存在时才设置标题
if (user?.nickname) {
  useHead({
    title: useSiteTitle(t("pages.user.profile"), user.nickname),
  });
}
</script>

<style lang="scss" scoped>
.tabs-warp {
  background-color: var(--bg-color);
  padding: 0 10px 10px;
  border-radius: var(--border-radius);

  .tabs {
    margin-bottom: 5px;
  }

  .more {
    text-align: right;
  }
}

.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 200px;
  
  .loading {
    font-size: 16px;
    color: var(--text-color3);
  }
}
</style>
