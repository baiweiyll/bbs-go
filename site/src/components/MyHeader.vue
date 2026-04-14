<template>
  <nav class="navbar has-shadow is-fixed-top" role="navigation" aria-label="main navigation">
    <div class="container">
      <div class="navbar-brand">
        <nuxt-link to="/" class="navbar-item">
          <img v-if="config.siteLogo" :alt="config.siteTitle" :src="config.siteLogo" />
          <img v-else :alt="config.siteTitle" src="~/assets/images/logo.svg" />
        </nuxt-link>
        <a :class="{ 'is-active': navbarActive }" class="navbar-burger burger" data-target="navbarBasic"
          @click="toggleNav">
          <span aria-hidden="true" />
          <span aria-hidden="true" />
          <span aria-hidden="true" />
        </a>
      </div>
      <div :class="{ 'is-active': navbarActive }" class="navbar-menu">
        <div class="navbar-start">
          <nuxt-link v-for="(nav, index) in config.siteNavs" :key="index" :to="nav.url" class="navbar-item">
            {{ nav.title }}
          </nuxt-link>
        </div>

        <div class="navbar-end">
          <div class="navbar-item">
            <search-input />
          </div>

          <div class="navbar-item">
            <create-topic-btn />
          </div>

          <msg-notice v-if="user" />

          <div v-if="user" class="navbar-item has-dropdown is-hoverable user-menus">
            <div class="navbar-link">
              <MyAvatar :user="user" :size="24" />
              <span :to="`/user/${user.id}`" class="user-menus-nickname ellipsis">{{ user.nickname }}</span>
            </div>
            <div class="navbar-dropdown">
              <nuxt-link class="navbar-item" :to="`/user/${user.id}`">
                <i class="iconfont icon-username" />
                <span>{{ $t("common.header.profile") }}</span>
              </nuxt-link>
              <nuxt-link class="navbar-item" to="/user/favorites">
                <i class="iconfont icon-favorite" />
                <span>{{ $t("common.header.favorites") }}</span>
              </nuxt-link>
              <nuxt-link class="navbar-item" to="/user/profile">
                <i class="iconfont icon-edit" />
                <span>{{ $t("common.header.editProfile") }}</span>
              </nuxt-link>
              <a class="navbar-item" @click="signout">
                <i class="iconfont icon-log-out" />
                <span>{{ $t("common.header.logout") }}</span>
              </a>
            </div>
          </div>
          <div v-else class="navbar-item">
            <div class="buttons">
              <!-- <nuxt-link class="button login-btn" to="/user/signin">
                {{ $t("common.header.login") }}
              </nuxt-link> -->
              <a class="button login-btn" @click="handleLoginClick">
                {{ $t("common.header.login") }}
              </a>
            </div>
          </div>
          <div class="navbar-item">
            <color-mode />
          </div>
        </div>
      </div>
    </div>
  </nav>
</template>

<script setup>
const userStore = useUserStore();
const configStore = useConfigStore();

const { user } = storeToRefs(userStore);

// 监听用户状态变化
watch(user, (newUser) => {
  console.log('MyHeader - User state changed:', newUser);
  console.log('MyHeader - User ID:', newUser?.id);
  console.log('MyHeader - User nickname:', newUser?.nickname);
}, { immediate: true });

// 组件挂载时检查当前用户状态
onMounted(() => {
  console.log('MyHeader mounted - Current user:', user.value);
  console.log('MyHeader mounted - UserStore user:', userStore.user);
  
  // 检查 localStorage
  if (typeof localStorage !== 'undefined') {
    const storedUserInfo = localStorage.getItem('userInfo');
    console.log('MyHeader - Stored user info:', storedUserInfo);
  }
});
const { config } = storeToRefs(configStore);
const { t } = useI18n();

const navbarActive = ref(false);

function toggleNav() {
  navbarActive.value = !navbarActive.value;
}

async function signout() {
  if (confirm(t("common.header.confirmLogout"))) {
    await userStore.signout();
    useLinkTo("/");
  }
}

function handleLoginClick() {
  // 跳转到 OIDC 登录接口，通过 Nuxt 代理转发到后端
  window.location.href = "/bbsoidc/login/signin";
}

</script>

<style lang="scss" scoped>
.navbar {
  /*opacity: 0.99;*/
  /*border-bottom: 1px solid #e7edf3;*/
  background-color: var(--bg-color);

  .navbar-item {
    font-weight: 700;
  }

  .publish {
    color: var(--text-color);
    background-color: #3174dc;
    width: 100px;

    &:hover {
      color: var(--text-color);
      background-color: #4d91fa;
    }
  }

  .login-btn {
    height: 32px;
    border-color: #000; // TODO

    &:hover {
      color: var(--text-color3);
      border-color: var(--text-color3);
    }
  }
}

.user-menus {
  .navbar-link {
    display: flex;
    align-items: center;

    .user-menus-nickname {
      margin-left: 5px;
      padding: 0 4px;
      font-size: 14px;
      color: var(--text-color);

      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
  }

  .navbar-dropdown {
    border: 1px solid var(--border-color);

    a {
      display: flex;
      align-items: center;

      // padding: 8px 16px;
      img {
        width: 20px;
        height: 20px;
      }

      span {
        margin-left: 10px;
        width: 56px;
        height: 20px;
        font-size: 14px;
        font-weight: 400;
        line-height: 20px;
      }
    }
  }
}
</style>
