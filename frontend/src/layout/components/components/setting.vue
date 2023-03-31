

<template>
  <a-drawer
    class="backend-setting"
    v-model:visible="visible"
    :on-before-ok="save"
    width="350px"
    :ok-text="$t('sys.saveToBackend')"
    @cancel="close"
    unmountOnClose
  >
    <template #title>配置</template>
    <a-form :model="form" :auto-label-width="true">


      <a-form-item :label="交易所"  >
        <a-select v-model="form.layout" @change="handleLayout">
          <a-option value="binace">Binance</a-option>
          <a-option value="okex">Okex</a-option>

        </a-select>
      </a-form-item>


    </a-form>
  </a-drawer>


</template>

<script setup>
import { ref, reactive, watch } from 'vue'
import { useAppStore, useUserStore } from '@/store'
import { Message } from '@arco-design/web-vue'
import user from '@/api/system/user'
import Skin from './skin.vue'
import skins from '@/config/skins'
import { useI18n } from 'vue-i18n'
import { ColorPicker } from 'vue-color-kit'
import 'vue-color-kit/dist/vue-color-kit.css'

const userStore = useUserStore()
const appStore  = useAppStore()
const { t } = useI18n()

const skin = ref(null)
const visible = ref(false)
const okLoading = ref(false)
const currentSkin = ref('')
const form = reactive({
  mode: appStore.mode === 'dark',
  tag: appStore.tag,
  menuCollapse: appStore.menuCollapse,
  menuWidth: appStore.menuWidth,
  layout: appStore.layout,
  language: appStore.language,
  animation: appStore.animation,
  i18n: appStore.i18n,
})

const defaultColorList = reactive([
  '#165DFF', '#F53F3F', '#F77234', '#F7BA1E', '#00B42A', '#14C9C9', '#3491FA',
  '#722ED1', '#F5319D', '#D91AD9', '#34C759', '#43a047', '#7cb342', '#c0ca33',
  '#86909c', '#6d4c41',
])
const changeColor = (color) => {
  appStore.changeColor(color.hex)
}

skins.map(item => {
  if (item.name === appStore.skin) currentSkin.value = t('skin.' + item.name)
})

watch(() => appStore.skin, v => {
  skins.map(item => {
    if (item.name === v) currentSkin.value = t('skin.' + item.name)
  })
})

const open = () => visible.value = true
const close = () => visible.value = false

const handleLayout = (val) => appStore.changeLayout(val)
const handleI18n = (val) => appStore.toggleI18n(val)
const handleLanguage = (val) => appStore.changeLanguage(val)
const handleAnimation = (val) => appStore.changeAnimation(val)
const handleSettingMode = (val) => appStore.toggleMode(val ? 'dark' : 'light')
const handleSettingTag = (val) => appStore.toggleTag(val)
const handleMenuCollapse = (val) => appStore.toggleMenu(val)
const handleMenuWidth = (val) => appStore.changeMenuWidth(val)

watch(() => appStore.menuCollapse, val => form.menuCollapse = val)

const save = async (done) => {
  const data = {
    mode: appStore.mode,
    tag: appStore.tag,
    menuCollapse: appStore.menuCollapse,
    menuWidth: appStore.menuWidth,
    layout: appStore.layout,
    skin: appStore.skin,
    i18n: appStore.i18n,
    language: appStore.language,
    animation: appStore.animation,
    color: appStore.color
  }

  user.updateInfo({ id: userStore.user.id, backend_setting: data }).then(res => {
    res.success && Message.success(res.message)
  })
  done(true)
}

defineExpose({ open })
</script>


