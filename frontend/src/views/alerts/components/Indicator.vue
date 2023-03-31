<template>
  <div class="layout-kline" style="height: 880px;">
    <a-layout style="height: 600px;">
      <a-layout-header>
        <a-row class="grid-demo" style="margin-bottom: 16px;height:40px;background-color: white">
          <a-col flex="100px">
            <div>
              <a-select :style="{width:'180px'}" placeholder="选择交易对" allow-search v-model="options.symbol" @change="resetSymbol">
                <a-option>ETHUSDT</a-option>
                <a-option>BTCUSDT</a-option>
                <a-option>BNBUSDT</a-option>
              </a-select>
            </div>
          </a-col>
          <a-col flex="auto">
            <div>
              <a-space>
                <a-radio-group v-model:model-value="options.interval" type="button">
                  <a-radio v-for="interval in intervals" :value="interval"
                           @click="resetInterval(interval)">{{ interval }}
                  </a-radio>

                </a-radio-group>

              </a-space>
            </div>
          </a-col>
        </a-row>


      </a-layout-header>
      <a-layout-content>
        <div id="indicator-k-line" class="k-line-chart"/>
        <div class="k-line-chart-menu-container">
          <span style="padding-right: 10px">主图指标</span>
          <button
              v-for="type in mainIndicators"
              :key="type"
              v-on:click="setMainIndicator(type)"
          >
            {{ type }}
          </button>
          <button v-on:click="setMainIndicator('EMOJI')">自定义</button>
          <span style="padding-right: 10px; padding-left: 12px">副图指标</span>
          <button
              v-for="type in subIndicators"
              :key="type"
              v-on:click="setSubIndicator(type)"
          >
            {{ type }}
          </button>
          <button v-on:click="setSubIndicator('EMOJI')">自定义</button>
        </div>
      </a-layout-content>

    </a-layout>
  </div>
</template>


<script setup>
import {ref, reactive, watchEffect, inject, onMounted, watch, onUnmounted} from 'vue'

import {dispose, init, registerIndicator} from "klinecharts";
import generatedDataList from "../api/kline";


import { emitter } from '@/utils/bus.js'
import useSocket from '@/utils/socket.js'

const kchart = ref()
const mainIndicators = reactive(["MA", "EMA", "SAR", "SMA", "BOLL"])
const subIndicators = reactive(["VOL", "MACD", "KDJ", "RSI", "CCI", "DMI"])
const panes = ref({})
const intervals = reactive([
  "1m", "3m", "5m", "10m", "15m", "30m",
  "1h", "2h", "3h", "4h", "8h", "12h",
  "1d", "2d", "3d", "1w"
])
const options = reactive({
  symbol: "ETHUSDT",
  interval: "1m",
})

const socketAdaptor = (data) => {
  var candle = data.k;
  var time, open, high, low, close, volume;

  time = candle.t;
  open = candle.o;
  high = candle.h;
  low = candle.l;
  close = candle.c;
  volume = candle.v;

  return {
    timestamp: time,
    open: parseFloat(open),
    high: parseFloat(high),
    low: parseFloat(low),
    close: parseFloat(close),
    value: parseFloat(volume),

  };
};

const fruits = [
  "🍏",

];

registerIndicator({
  name: "EMOJI",
  figures: [{key: "emoji"}],
  calc: (kLineDataList) => {
    return kLineDataList.map((kLineData) => ({
      emoji: kLineData.close,
      text: fruits[Math.floor(Math.random() * 17)],
    }));
  },
  draw: ({ctx, barSpace, visibleRange, indicator, xAxis, yAxis}) => {
    const {from, to} = visibleRange;
    ctx.font = `${barSpace.gapBar}px Helvetica Neue`;
    ctx.textAlign = "center";
    const result = indicator.result;
    for (let i = from; i < to; i++) {
      const data = result[i];
      const x = xAxis.convertToPixel(i);
      const y = yAxis.convertToPixel(data.emoji);
      ctx.fillText(data.text, x, y);
    }
    return false;
  },
});
const sockets= ref(new Map())
const resetInterval = (interval) => {
  options.interval = interval
  // sockets.value.map((key,item)=>{
  //
  //   item()
  // })

  for (let sk of sockets.value.keys()) {

    sockets.value.get(sk)()

  }
  getKlineData()
}
const resetSymbol = () => {
  emitter.emit('symbolChange', options.symbol)

  getKlineData()
}



const getKlineData = async function () {
  const key=`${options.symbol.toLowerCase()}@kline_${options.interval}`
  const SOCKET_URL = `wss://stream.binance.com/ws/${key}`;

  if (sockets.value.has(key)){
    return
  }


  let res = await generatedDataList(options.symbol, options.interval)
  let kline = []
  res.forEach((prices, index) => {
    const kLineData = {
      open: parseFloat(prices[1]),
      high: parseFloat(prices[2]),
      low: parseFloat(prices[3]),
      close: parseFloat(prices[4]),
      volume: parseFloat(prices[5]),
      timestamp: parseInt(prices[0]),
    };
    kline.push(kLineData)
  })
  kchart.value.applyNewData(kline);


  const {
    readyState,
    latestMessage,
    disconnect,
    connect,
    sendMessage,
  } = useSocket(SOCKET_URL)
  sockets.value.set(key, function (){
    console.error(key)
    disconnect()

  });

  watchEffect(() => {
    if (latestMessage.value != undefined) {

      const dt = JSON.parse(latestMessage.value.data);
     // console.log("[333]",options.interval==dt.k.i &&options.symbol==dt.s)

      if (options.interval==dt.k.i &&options.symbol==dt.s){
     //   console.log("[4444]",options,dt)
        const candle = socketAdaptor(dt);
        //console.log(latestMessage.value, candle)
        kchart.value.updateData(candle);
      }

    }


  })

}
const setMainIndicator = function (name) {
  kchart.value.createIndicator(name, false, {id: "candle_pane"});
}
const setSubIndicator = function (name) {
  if (this.panes[name]) {
    kchart.value.removeIndicator(this.panes[name]);
    delete this.panes[name]

  } else {
    this.paneId = kchart.value.createIndicator(name);
    this.panes[name] = this.paneId
    kchart.value.createIndicator(name, false, {id: this.paneId});
  }

}

onMounted(() => {
  kchart.value = init("indicator-k-line");
  // const SOCKET_URL = `wss://stream.binance.com/ws/${options.symbol.toLowerCase()}@kline_${options.interval}`;
  //
  // Wsm.seturl(SOCKET_URL)
  // Wsm.init()
  getKlineData()
})

onUnmounted(() => {
  dispose("indicator-k-line");
})

</script>
<style>


.k-line-chart-container {
  display: flex;
  flex-direction: column;
  margin-right: 50px;
  border-radius: 2px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
  background-color: #ffffff;
  width: 100%;
  height: 800px;
  padding: 16px 6px 16px 16px;
  padding-right: 50px;

}

.k-line-chart-title {
  margin: 0;
  color: #252525;
  padding-bottom: 10px;
}

.k-line-chart {
  display: flex;
  flex: 1;
  height: 300px;
  background-color: #ffffff;
}

.k-line-chart-menu-container {
  display: flex;
  flex-direction: row;
  align-items: center;
  margin-top: 10px;
  font-size: 12px;
  color: #606060;
}

.k-line-chart-menu-container button {
  cursor: pointer;
  background-color: #1677ff;
  border-radius: 2px;
  margin-right: 8px;
  height: 24px;
  line-height: 26px;
  padding: 0 6px;
  font-size: 12px;
  color: #fff;
  border: none;
  outline: none;
}
</style>
<style scoped>
.layout-kline :deep(.arco-layout-header),
.layout-kline :deep(.arco-layout-footer),
.layout-kline :deep(.arco-layout-sider-children),
.layout-kline :deep(.arco-layout-content) {
  display: flex;
  flex-direction: column;
  justify-content: center;
  color: var(--color-white);
  font-size: 16px;
  font-stretch: condensed;
  text-align: center;
}


.layout-kline :deep(.arco-layout-header) {
  height: 64px;
  background-color: #fcfcfc;
}

.layout-kline :deep(.arco-layout-sider) {
  width: 206px;
  background-color: var(--color-primary-light-3);
}

.layout-kline :deep(.arco-layout-content) {
  background-color: white;
}
</style>
