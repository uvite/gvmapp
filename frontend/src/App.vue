
<script setup>
  import cn from '@arco-design/web-vue/es/locale/lang/zh-cn'
  import en from '@arco-design/web-vue/es/locale/lang/en-us'
  import { ref } from 'vue'
  import { useAppStore,useUserStore } from './store'
  const appStore = useAppStore()
  const userStore = useUserStore()
  const lang = ref(appStore.language === 'zh_CN' ? cn : en)
  userStore.setAppInfo()
</script>

<template>
  <a-config-provider :locale="lang">
    <router-view />
  </a-config-provider>
</template>
<style>
html,body,#app{
  height: 100vh;
}
#app {
  font-family: 'Avenir', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
}

</style>
