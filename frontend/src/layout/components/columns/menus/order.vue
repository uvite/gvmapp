<template>
  <a-layout>
    <div class="menu-title flex items-center">
      <a-space>
        <a-button >当前价格</a-button>
      </a-space>
    </div>
    <a-layout>
      <a-layout-sider :resize-directions="['bottom']" style="width: 100%">
        <a-list :max-height="280"  style="height: 280px"   :scrollbar="scrollbar">
          <a-list-item v-for="(item,index) in asksData" :key="index">{{ item[0] }} </a-list-item>
        </a-list>
      </a-layout-sider>
    </a-layout>
    <a-layout-footer>  <a-card>
      <a-tabs type="rounded">
        <a-tab-pane key="open" title="开仓">
          <open-order/>
        </a-tab-pane>
        <a-tab-pane key="close" title="平仓">
          <close-order/>
        </a-tab-pane>
        <a-tab-pane key="limitstop" title="止赢止损">
          <pending-order/>
        </a-tab-pane>
      </a-tabs>
    </a-card>
    </a-layout-footer>
  </a-layout>



</template>

<script setup>
import CloseOrder from './closeorder.vue'
import OpenOrder from './openOrder.vue'
import PendingOrder from './pendingOrder.vue'
import {onMounted, onUnmounted, reactive, ref, watchEffect} from "vue";
import {useWebSocket} from "v3hooks";
import {emitter} from "@/utils/bus";

import {useGenvStore} from '@/store'
const genvStore = useGenvStore()

const options = reactive({
  symbol: "ETHUSDT",
  interval: "1m",
})
let initdata=genvStore.getGenv()
if(initdata){
  options.value=initdata

}

const scrollbar = ref(true);
const bids=ref()
const asksData=ref([])
const asks=reactive([])
const Depth=[]
const getAppInfo =   () => {

  if (Depth[options.symbol.toLowerCase()]){
    console.log("[111]",Depth[options.symbol.toLowerCase()],options.symbol.toLowerCase())
    return
  }
  console.log("[222]",Depth[options.symbol.toLowerCase()])
  Depth[options.symbol.toLowerCase()]=true
  const SOCKET_URL = `wss://stream.binance.com/ws/${options.symbol.toLowerCase()}@depth`;

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

      if (options.symbol==data.s ){
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
    options.symbol=data.symbol
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
