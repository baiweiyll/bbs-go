import { defineStore } from "pinia";

export const useUserStore = defineStore("user", {
  state: () => ({
    user: null,
  }),
  getters: {
    isLogin() {
      return !!this.user;
    },
  },
  actions: {
    // 设置用户信息
    setUser(userInfo) {
      console.log('UserStore - Setting user:', userInfo);
      this.user = userInfo;
      console.log('UserStore - User set to:', this.user);
      
      // 同步保存到 localStorage
      if (typeof localStorage !== 'undefined' && userInfo) {
        localStorage.setItem('userInfo', JSON.stringify(userInfo));
        console.log('UserStore - Saved to localStorage:', userInfo.id);
      } else if (typeof localStorage !== 'undefined' && !userInfo) {
        localStorage.removeItem('userInfo');
        console.log('UserStore - Removed from localStorage');
      }
    },
    
    async fetchCurrent() {
      const { data } = await useMyFetch("/api/user/current");
      this.user = data.value;
      return this.user;
    },
    async signin(body) {
      const { user, token, redirect } = await useHttpPost(
        "/api/login/signin",
        useJsonToForm(body)
      );
      this.user = user;
      return {
        user,
        token,
        redirect,
      };
    },
    async signout() {
      // await useHttpGet("/oidc/login/signout");
      this.user = null;
      // 清除 localStorage 中的用户信息
      if (typeof localStorage !== 'undefined') {
        localStorage.removeItem('userInfo');
      }
       window.location.href = "/oidc/login/signout";
    },
    async signup(form) {
      const { user, token, redirect } = await useHttpPost(
        "/api/login/signup",
        useJsonToForm(form)
      );
      this.user = user;
      return {
        user,
        token,
        redirect,
      };
    },
  },
});
