

<template>
  <a-layout-content class="h-full main-container">
    <columns-layout   />


    <ma-button-menu />
  </a-layout-content>
</template>
<script setup>
  import { useAppStore, useUserStore } from '@/store'

  import ColumnsLayout from './components/columns/index.vue'
  import ClassicLayout from './components/classic/index.vue'
  import BannerLayout from './components/banner/index.vue'
  import MaButtonMenu from './components/ma-buttonMenu.vue'

  const appStore = useAppStore()
  const userStore = useUserStore()
  userStore.setAppInfo()


</script>
<style scoped lang="less">
</style>
