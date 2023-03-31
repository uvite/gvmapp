<template>

  <div>

          <Indicator/>


  </div>


</template>

<script setup>



import {ref, reactive, watchEffect, onMounted, onUnmounted} from 'vue'

import {useDocStore} from '@/store'

import { emitter } from '@/utils/bus.js'

import Indicator from "./components/Indicator.vue";
import {useWebSocket} from "v3hooks";
const splitSize=ref({})
splitSize.value=0.7

const docStore = useDocStore()


const columns =reactive( [
  {
    title: 'Price',
    dataIndex: 'Price',
  },
  {
    title: 'Amount',
    dataIndex: 'Amount',
  },
  {
    title: 'Total',
    dataIndex: 'Total',
  },

]);

const options = ref({
  symbol: "ETHUSDT",
  interval: "1m",
})

const scrollbar = ref(true);
const bids=ref()
const asksData=ref([])
const asks=reactive([])
const Depth=[]
const getAppInfo =   () => {

  if (Depth[options.value.symbol.toLowerCase()]){
    console.log("[111]",Depth[options.value.symbol.toLowerCase()],options.value.symbol.toLowerCase())
    return
  }
  console.log("[222]",Depth[options.value.symbol.toLowerCase()])
  Depth[options.value.symbol.toLowerCase()]=true
  const SOCKET_URL = `wss://stream.binance.com/ws/${options.value.symbol.toLowerCase()}@depth`;

  const {
    readyState,
    latestMessage,
    disconnect,
    connect,
    sendMessage,
  } = useWebSocket(SOCKET_URL)
  //
  // const handleSendMessage = () => {
  //   //sendMessage('hello v3hooks')
  // }
  watchEffect(() => {
    if (latestMessage.value != undefined) {

      const data = JSON.parse(latestMessage.value.data);
      // console.log("[333]",data)

      if (options.value.symbol==data.s ){
        let [asksCreated, bidsCreated] = [
          data.a.filter(item => item[1] != 0),
          data.b.filter(item => item[1] != 0)
        ];
        //this.bids.sort((a, b) => b[0] - a[0])

        asks.splice(asks.length - asksCreated.length, asksCreated.length);

        // bids.value.splice(bids.value.length - bidsCreated.length, bidsCreated.length);
        asks.value = [...asksCreated, ...asks, ];
        asks.value=asks.value.slice(0,5)
        asks.value.sort((a, b) => a[0] - b[0])

        asks.value.reverse()
        // console.log("[asks]",asks)
        asksData.value=asks.value
        // bids.value = [...bidsCreated, ...this.bids, ];
      }

    }


  })
}


const initPage = () => {
  // 全局监听 关闭当前页面函数
  emitter.on('symbolChange', (data) => {

    options.value=data
    getAppInfo()
  })
  getAppInfo()
}
onMounted(() => {
  initPage()
})
onUnmounted(() => {
  emitter.off('symbolChange')

})

</script>

