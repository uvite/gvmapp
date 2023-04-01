<template>

  <div class="layout-menu shadow flex flex-col" v-show="showMenu">


    <a-layout-sider
        :style="
        `width: ${appStore.menuCollapse ? '250px' :   '300px'};
        height: ${appStore.menuCollapse ? '100%' : 'calc(100% )'};`"
        :resize-directions="['left']"
    >


      <gvm-board v-show="showComment=='board'" ref="MaMenuRef"
                 :class="appStore.menuCollapse ? 'ml-0.5' : ''"></gvm-board>
      <gvm-alert v-show="showComment=='alert'"></gvm-alert>
      <gvm-message v-show="showComment=='board2'"></gvm-message>
      <gvm-deal v-show="showComment=='deal'"></gvm-deal>
      <gvm-data v-show="showComment=='data'"></gvm-data>
    </a-layout-sider>


  </div>

  <div class="sider flex flex-col items-center bg-gray-800 dark:border-blackgray-5">
    <a-avatar class="mt-2" :size="40"><img src="/logo.svg" class="bg-white"/></a-avatar>
    <ul class="mt-1 parent-menu-container">
      <template
          v-for="(bigMenu, index) in AppPage"
          :key="index"
      >
        <li
            :class="`${classStyle}`"
            @click="loadMenu(bigMenu, index,true)"
        >
          <component v-if="bigMenu.meta.icon" :is="bigMenu.meta.icon" class="text-xl mt-1"/>
          <span
              class="mt-0.5"
              :style="appStore.language === 'en' ? 'font-size: 10px' : ''"
          >{{
              appStore.i18n ? ($t(`menus.${bigMenu.name}`).indexOf('.') > 0 ? bigMenu.meta.title : $t(`menus.${bigMenu.name}`)) : bigMenu.meta.title
            }}</span>
        </li>
      </template>
    </ul>

  </div>

</template>

<script setup>
import {ref, onMounted, watch, defineAsyncComponent} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import MaMenu from '../ma-menu.vue'
import {useAppStore, useUserStore} from '@/store'
import AppPage from '@/router/appRoutes'

import GvmMessage from './menus/message.vue'
import GvmAlert from './menus/alert.vue'
import GvmBoard from './menus/message.vue'
import GvmDeal from './menus/order.vue'
import GvmData from './menus/data.vue'


const route = useRoute()
const router = useRouter()

const MaMenuRef = ref(null)
const appStore = useAppStore()
const userStore = useUserStore()
const showMenu = ref(true)
const selectMenu = ref(0)
const title = ref('')


const classStyle = ref('flex flex-col parent-menu items-center rounded mt-1 text-gray-200 hover:bg-gray-700 dark:hover:text-gray-50 dark:hover:bg-blackgray-1')
//
// const GvmMessage = defineAsyncComponent(() =>
//     import('./menus/message.vue')
// )
// const GvmAlert = defineAsyncComponent(() =>
//     import('./menus/alert.vue')
// )
// const GvmBoard = defineAsyncComponent(() =>
//     import('./menus/board.vue')
// )
// const GvmDeal = defineAsyncComponent(() =>
//     import('./menus/order.vue')
// )
// const GvmData = defineAsyncComponent(() =>
//     import('./menus/data.vue')
// )
const showComment = ref("alert")


onMounted(() => {
  initMenu()
})

watch(() => route, v => {
  initMenu()
}, {deep: true})

const initMenu = () => {
  let current
  if (route.matched[1]?.meta?.breadcrumb) {
    current = route.matched[1].meta.breadcrumb[0].name
  } else {
    current = 'home'
  }
  if (userStore.routers && userStore.routers.length > 0) {
    userStore.routers.map((item, index) => {
      if (item.name == current) loadMenu(item, index, false)
    })
  }
}

const loadMenu = (bigMenu, index, flag) => {

  //showMenu.value = true
  //dynamicComponent.value=bigMenu.name
  console.log(bigMenu.name)
  switch (bigMenu.name) {
    case "data":
      showComment.value = "data"

    break;
      case "alert":
      showComment.value = "alert"
      //router.push("/dashboard")
      break;
    case "deal":
      showComment.value = "deal"
      break;
    case "dashboard":
      if (bigMenu.children.length > 0) {
        MaMenuRef.value.loadChildMenu(bigMenu)

      }
      showComment.value = "board"
      break;
    default:

  }
  if (selectMenu.value == index && flag) {
    showMenu.value = showMenu.value ? false : true

  }



  selectMenu.value = index
  document.querySelectorAll('.parent-menu').forEach((item, id) => {
    index !== id ? item.classList.remove('active') : item.classList.add('active')
  })
}
</script>

<style>
.parent-menu-container {
  width: 62px;
}

.parent-menu {

  padding: 5px;
  height: 57px;
  cursor: pointer;
  font-size: 13px;
  fill: #fff;
  transition: all .2s;
}

.parent-menu.active {
  background: rgb(var(--primary-6));
  color: #fff;
}

:deep(.arco-menu-vertical .arco-menu-inner) {
  padding: 4px;
}

:deep(.arco-menu-vertical .arco-menu-item) {
  padding: 0px 9px;
  line-height: 36px;
}
</style>
