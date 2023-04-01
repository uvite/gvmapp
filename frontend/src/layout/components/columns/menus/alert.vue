<template>
  <div>
    <div class="menu-title flex items-center">
      <a-space>
        <a-button type="primary" status="success" @click="createAlert">创建警报</a-button>

      </a-space>
    </div>

    <div>
      <a-split direction="vertical" :style="{height: '700px'}" v-model:size="size">
        <template #first>

          <a-list
              :style="{ width: `100%` }"
              :virtualListProps="{
                height: 560,
              }"
              :data="data"
          >
            <template #item="{ item, index }">

              <a-list-item :key="index" style="border-bottom: rgba(14,1,1,0.14) solid 1px">
                <a-row class="grid-demo" style="margin-bottom: 16px;">
                  <a-col flex="50px">
                    <div>{{ item.symbol }}</div>
                  </a-col>
                  <a-col flex="auto">
                    <a-button-group>
                      <a-button type="primary" status="success" size="small" @click="runBot(item)">运行</a-button>
                      <a-button type="primary" status="danger" size="small" @click="closeBot(item)">停止</a-button>
                      <a-button size="small" @click="deleteR(item)"> 删除</a-button>
                    </a-button-group>
                  </a-col>
                </a-row>


              </a-list-item>
            </template>
          </a-list>


        </template>
        <template #second>
          {{ symint }}
          <a-button type="primary" status="success" @click="test">测试</a-button>

        </template>
      </a-split>
    </div>


    <a-modal v-model:visible="visible" title="创建新警报" @cancel="handleCancel" @before-ok="handleBeforeOk" width="auto"
             height="auto">
      <new-task ref="editRef" :symint="symint"/>
    </a-modal>

  </div>
</template>
<script setup>

import {onMounted, onActivated, ref, reactive} from 'vue';
import NewTask from './newTask.vue'
import {emitter} from "@/utils/bus";
import {useGenvStore} from '@/store'
import rpc from '@/rpc'
import {Message} from "@arco-design/web-vue";
import * as PoolService from "@/wailsjs/go/services/PoolService";

const genvStore = useGenvStore()
const options = ref({
  symbol: "ETHUSDT",
  interval: "1m"
})
let symint = genvStore.getGenv()
if (symint) {
  options.value = symint
}

const editRef = ref()
const show = ref(true)
const sizeList = ref('small');
const size = ref(0.5)
const data = ref([])




const runBot = async (record) => {
  console.log("record.id", record.id)
  const res = await rpc.LauncherService.RunTask(record)
  console.log(res)
  // app.value.RunAlert(record.id).then(res => {
  //   console.log(res)
  //
  // })
}
const closeBot = async (record) => {
  const res = await rpc.LauncherService.CloseTask(record)
  console.log(res)
}

const deleteR = async (record) => {
  await rpc.AlertService.DelAlertItem(record.id).then(res => {
    console.log(res)
    rpc.emit("service.alert.getall")
  })
}


const visible = ref(false);
const createAlert = () => {
  visible.value = true;
};

const handleBeforeOk = async (done) => {

  options.value.metadata = editRef.value.form
  options.value.content = editRef.value.code
  const res = await rpc.AlertService.CreateAlert(options.value)

  done()
  rpc.emit("service.alert.getall")

};
const handleCancel = () => {
  visible.value = false;
}

async function load() {

  rpc.setPageTitle('K线助手')
  try {
    const res = await rpc.AlertService.GetAlertList()
    console.log("alert all", JSON.stringify(res))
    if (res.code == 200) {
      data.value = res.data.list
    }

  } catch (e) {
    console.log("eee", e)
  }
  emitter.on('symbolChange', (data) => {
    options.value = data
    symint.value = data

  })


}
const speak=async (text)=> {
  //await rpc.PoolService.Speak(text)
  const msg = new SpeechSynthesisUtterance();
  msg.text = text;
  msg.volume = 1.0; // speech volume (default: 1.0)
  msg.pitch = 1.0; // speech pitch (default: 1.0)
  msg.rate = 1.0; // speech rate (default: 1.0)
  msg.lang = 'zh-CN'; // speech language (default: 'en-US')
  // msg.voiceURI = 'Google UK English Female'; // voice URI (default: platform-dependent)
  // msg.onboundary = function (event) {
  //   console.log('Speech reached a boundary:', event.name);
  // };
  // msg.onpause = function (event) {
  //   console.log('Speech paused:', event.utterance.text.substring(event.charIndex));
  // };
  window.speechSynthesis.speak(msg);
}
onMounted(() => {
  load()
  rpc.on('service.alert.all', (res) => {
    console.log(res)
    if (res) {
      data.value = res
    } else {
      data.value = []
    }
  })
  rpc.on('service.alert.create', (res) => {
    console.log("service.alert.create", res)
  })
  rpc.on('service.message.alert', (res) => {
    Message.success(res)
    speak(res)
    console.log("service.alert.create", res)
  })
})

</script>
