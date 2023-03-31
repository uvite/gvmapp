<template>

  <div>
    <a-split :style="{
        height: '1200px',
        width: '100%',
        minWidth: '500px',
        border: '1px solid var(--color-border)'
      }"
             v-model:size="splitSize"
             min="80px"
    >
      <template #first>
        <a-typography-paragraph>
          <Indicator/>

        </a-typography-paragraph>
      </template>
      <template #second>



        <a-typography-paragraph>


          <a-layout style="height: 400px;">

            <a-layout-content style="background-color: white">
              <a-list :max-height="240"  style="height: 350px"   :scrollbar="scrollbar">
                <a-list-item v-for="(item,index) in asksData" :key="index">{{ item[0] }}---{{ item[1] }}---{{ item[0] * item[1] }}</a-list-item>
              </a-list>
              <!--              <a-list>-->
              <!--                <template #header>-->
              <!--                 -->
              <!--                </template>-->

              <!--                <a-list-item  v-for="(item,index) in asksData" :key="index" >{{ item[0] }}&#45;&#45;{{ item[1] }}&#45;&#45;{{ item[0] * item[1] }}</a-list-item>-->

              <!--              </a-list>-->
            </a-layout-content>
            <a-layout-footer>

              <a-card>
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

        </a-typography-paragraph>
      </template>
    </a-split>
  </div>


</template>

<script setup>


import CloseOrder from './components/closeorder.vue'
import OpenOrder from './components/openOrder.vue'
import PendingOrder from './components/pendingOrder.vue'

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
const options = reactive({
  symbol: "ETHUSDT",
  interval: "1m",
})
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
  emitter.on('symbolChange', (symbol) => {
    options.symbol=symbol
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

