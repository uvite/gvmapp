<template>
  <a-form class="w-full md:w-full mt-3" :model="order" @submit="modifyInfo" :size="form.size">
    <a-form-item label="价格" label-col-flex="80px">
      <a-input  v-model="order.price" allow-clear/>为空则为市价开单
    </a-form-item>
    <a-form-item label="数量" label-col-flex="80px">
      <a-input v-model="order.size" allow-clear/>
    </a-form-item>

    <a-form-item label-col-flex="80px">
      <a-space direction="vertical">
        <a-space>
          <a-button type="primary" status="success" @click="openLongPostion"> 开多</a-button>
          <a-button type="primary" status="danger" @click="openShortPostion">开空</a-button>

        </a-space>
      </a-space>
    </a-form-item>
  </a-form>
</template>

<script setup>
import {reactive, ref} from 'vue'
import { useBotsStore } from '@/store'
import Base from '@/api/online/base'

const botStore = useBotsStore()
const form = reactive({
  size: 'small'
})
const order = ref({})

const openLongPostion = async () => {
  const data={}
  data.side = "Long"
  data.size=(order.value.size)
  data.price = (order.value.price)
  data.file = 'entry'
  console.log("After converting "+  data.price + ", type: " + typeof  data.price);
  const response = await Base.gvmRun(data)
  console.log(response)
}

const openShortPostion = async () => {
  const data={}
  data.side = "Short"
  data.size=(order.value.size)
  data.price = (order.value.price)
  data.file = 'entry'
  const response = await Base.gvmRun(data)
  console.log(response)
}

</script>
