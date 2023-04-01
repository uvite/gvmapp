

<template>
  <a-layout-content class="h-full main-container">
    <columns-layout   />


    <ma-button-menu />
  </a-layout-content>
</template>


<script setup lang="ts">
import { nextTick, onActivated, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import ColumnsLayout from './components/columns/index.vue'
import MaButtonMenu from './components/ma-buttonMenu.vue'
import rpc from '@/rpc'
import * as StartService from "@/wailsjs/go/services/StartService";
import { useAppStore, useUserStore } from '@/store'

const appStore = useAppStore()
const userStore = useUserStore()
userStore.setAppInfo()

const WIDTH = 1500
const HEIGHT = 1500
const SIZE = 100

const route = useRoute()

onMounted(() => {
  load()
  nextTick(() => {
    rpc.on('shortcut.view.refresh', () => {
      if (route.name === 'artwork.build') load()
    })
  })
  rpc.setPageTitle("Build Settings")
})

async function load() {
  const sync=await rpc.StartService.DownToLocal()
  if (sync?.code==200){
      appStore.setStrategie(sync.data)
  }
  rpc.on('debug', async (data) => {

    console.log(data)
  })
  console.log(sync)
}
</script>

<style scoped lang="less">
</style>
