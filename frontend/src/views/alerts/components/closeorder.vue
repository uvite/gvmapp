<template>
  <a-form class="w-full md:w-full mt-3" :model="order" @submit="modifyInfo">
    <a-form-item label="价格" label-col-flex="80px">
      <a-input disabled :default-value="order.price" allow-clear/>
    </a-form-item>
    <a-form-item label="数量" label-col-flex="80px">
      <a-input v-model="order.size" allow-clear/>
    </a-form-item>

    <a-form-item label-col-flex="80px">
      <a-space direction="vertical">
        <a-space>
          <a-button type="primary" status="success" @click="closeLongPostion">平多</a-button>
          <a-button type="primary" status="danger" @click="closeShortPostion">平空</a-button>

        </a-space>
      </a-space>
    </a-form-item>
  </a-form>
</template>

<script setup>
import {reactive, ref} from 'vue'
import Base from '@/api/online/base'
import { useBotsStore } from '@/store'
const botStore = useBotsStore()

const order = ref({})
const closeLongPostion=async()=>{

  order.value.side = "Long"
  order.value.exchange_code = botStore.exchangeId
  order.value.file = 'exit'
  order.value.size=(order.value.size)
  const response = await Base.gvmRun(order.value)
  console.log(response)
}

const closeShortPostion=async()=>{
  order.value.side = "Short"
  order.value.exchange_code = botStore.exchangeId
  order.value.file = 'exit'
  order.value.size=(order.value.size)
  const response = await Base.gvmRun(order.value)
  console.log(response)
}

</script>
