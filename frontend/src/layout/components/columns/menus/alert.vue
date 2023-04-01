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
                    <div>{{ item.metadata.symbol }}</div>
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
        </template>
      </a-split>
    </div>


    <a-modal v-model:visible="visible" title="创建新警报" @cancel="handleCancel" @before-ok="handleBeforeOk"  width="auto" height="auto">
      <new-task ref="editRef" :symint="symint"/>
    </a-modal>

  </div>
</template>
<script setup>

import {onMounted,onActivated, ref, reactive} from 'vue';

import NewTask from './newTask.vue'
import {emitter} from "@/utils/bus";
import {useGenvStore} from '@/store'
// import {OnDOMContentLoaded} from "../wailsjs/go/gvmapp/App"
import rpc from '@/rpc'
const genvStore = useGenvStore()
const options = ref({
  symbol: "ETHUSDT",
  interval: "1m"
})
let symint = genvStore.getGenv()
if (symint) {
  options.value = symint
}
const app = ref(window.go.gvmapp.App)
const editRef = ref()
const show = ref(true)
const sizeList = ref('small');
const size = ref(0.5)
const data = ref([])

const runBot = (record) => {
  app.value.RunAlert(record.id).then(res => {
    console.log(res)

  })
}
const closeBot = (record) => {
  app.value.CloseAlert(record.id).then(res => {
    console.log(res)

  })
}
const deleteR = (record) => {
  console.log(record)
  app.value.DelAlertItem(record.id).then(res => {
    getData()

  })
}


const getData = () => {
  app.value.GetAlertList().then(res => {

    if (res.code == 200) {
      data.value = res.data.list

    } else {
      message.error(res.msg)
    }
  })
}

const visible = ref(false);


const createAlert = () => {

  visible.value = true;
};
const handleBeforeOk = (done) => {

  options.value.metadata = editRef.value.form
  options.value.content = editRef.value.code

  done()
  console.log(options.value)

  let res = app.value.AddAlertItem(options.value)
  console.log("[3333]",res)
  getData()

};
const handleCancel = () => {
  visible.value = false;
}

onActivated(() => {
  load()
  nextTick(() => {
    rpc.on('shortcut.view.refresh', () => {
      if (route.name === 'dashboard') load()
    })
    rpc.on('shortcut.view.hard-refresh', () => {
      if (route.name === 'start') {
        store.documents = []
        load()
      }
    })
  })
})

async function load() {
  rpc.setPageTitle('dashboard')

  // projects.value = []
  //
  // store.documents.forEach(async (file) => {
  //   const contents = await rpc.FileSystemService.ReadFileContents(file)
  //   const collection = types.Collection.createFrom(JSON.parse(contents))
  //   projects.value = [...projects.value, collection]
  // })
}
onMounted(() => {

  emitter.on('symbolChange', (data) => {
    options.value = data
    symint.value = data

  })
  emitter.on('appRun', (data) => {

    getData()

  })
  // EventsOn("alertList",function(data){
  //   alert(JSON.stringify(data))
  //   console.log(data)
  // })
 // rpc.setPageTitle("Review & Export")

  //getData()

})

</script>
