<template>
  <a-form class="w-full md:w-full mt-3" :model="order" @submit="modifyInfo">
    <a-form-item label="触发价格" label-col-flex="80px">
      <a-input v-model="order.price" allow-clear/>
      不填为市价
    </a-form-item>
    <a-form-item label="止赢止损" label-col-flex="80px">
      <a-input v-model="order.limitstop" allow-clear/>
    </a-form-item>
    <a-form-item label="数量" label-col-flex="80px">
      <a-input v-model="order.size" allow-clear/>
    </a-form-item>

    <a-form-item label-col-flex="80px">
      <a-space direction="vertical">
        <a-space>
          <a-button type="primary" status="success" @click="cancelOrder">全部撤单</a-button>

        </a-space>

        <a-space>

          <a-button type="primary" status="success" @click="algoOrder('limit','Long')">止赢平多</a-button>
          <a-button type="primary" status="danger" @click="algoOrder('stop','Long')">止损平多</a-button>


        </a-space>
        <a-space>

          <a-button type="primary" status="success" @click="algoOrder('limit','Short')">止赢平空</a-button>
          <a-button type="primary" status="danger" @click="algoOrder('stop','Short')">止损平空</a-button>


        </a-space>
      </a-space>
    </a-form-item>
  </a-form>
</template>

<script setup>
import {reactive, ref} from 'vue'
import {useBotsStore} from '@/store'
import Base from '@/api/online/base'

const botStore = useBotsStore()

const order = ref({})

const cancelOrder = async () => {
  const data = {}
  data.file = 'cancel'
  const response = await Base.gvmRun(data)
  console.log(response)
}

const algoOrder = async (sltp, side) => {


  const data = {}
  data.side = side
  data[sltp] = (order.value.limitstop)
  data.size = (order.value.size)
  data.price = (order.value.price)
  data.file = sltp

  const response = await Base.gvmRun(data)
  console.log(response)
}


</script>
