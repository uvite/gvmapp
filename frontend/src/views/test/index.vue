<template>
  <a-layout style="height: 100%; background-color: white">

    <a-layout>
      <a-layout-sider>
        <a-button type="primary" @click="downToLocal" :disabled="sysnc">更新策略模板</a-button>
        <a-list>

          <a-list-item v-for="(item, index) in notes" @click="noteSelect(item)">{{ item }}</a-list-item>

        </a-list>


      </a-layout-sider>
      <a-layout-content>
        <a-tabs v-model:active-key="activeTab">
          <a-tab-pane title="简单模式" key="base_config">
            <a-divider orientation="left">基础信息</a-divider>
            {{ columns }}
            <a-row :gutter="24">

              <a-col :xs="24" :md="24" :xl="24">


                <ma-form v-if="showMa"
                    ref="maFormRef"
                    :columns="columns"
                    v-model="form"
                    :options="{ ...options, showButtons: false }"

                />
              </a-col>
            </a-row>
          </a-tab-pane>
          <a-tab-pane title="代码模式" key="code_config">


          </a-tab-pane>
        </a-tabs>


      </a-layout-content>
    </a-layout>

  </a-layout>
</template>
<script setup lang="ts">
import {Message} from "@arco-design/web-vue";
import {onMounted, reactive, ref} from "vue";

const app = ref(window.go.gvmapp.App)
const activeTab = ref('base_config')

const notebooks = ref([])
const notes = ref([])
const currentNotebook = ref()
const currentNote = ref()
const columns = ref([])
const form = ref({})
const options=ref({
  init:false
})

const showMa=ref(false)
const Load = () => {

  app.value.GetDirs().then((res) => {

    if (res && res.code == 200) {
      notebooks.value = res.data.dirs;
      //alert(JSON.stringify(res.data))
      currentNotebook.value = res.data.dirs[0];
      notes.value = res.data.files.map((n, i) => {
        return n.replace('.js', '')

      });
      notes.value.shift()
      // alert(JSON.stringify(notes.value))

    } else {
      Message.error('获取云端数据失败：' + res.msg);
    }
  });

}

const noteSelect = (key, keyPath) => {
  console.log("选中key：" + key);
  options.value.init=false
  columns.value = []
  form.value = {}
  currentNote.value = key + ".js";
  showMa.value=false
  app.value.ParseNoteFile(currentNotebook.value, currentNote.value).then((res) => {
    if (res && res.code == 200) {
      var data = res.data;


      // Object.keys(columns).forEach(key => (columns[key] = '' ));
      // Object.keys(form).forEach(key => (form[key] = '' ));
      // currentView.value="MaForm"
      let ui = JSON.parse(data.ui)
      console.log(ui)

      //columns.value = ui
      for (const key in ui) {

        if (ui[key].value) {
          console.log(ui[key].dataIndex, ui[key].value)
          form.value[ui[key].dataIndex] = ui[key].value
        }
      }
      columns.value = ui
      showMa.value=true
      // Object.assign(columns, ui)
      // mdText.value = data;
      //console.log("[1]",data)
      //
      // code.value=data
      // notEditedMdtext.value = data;    // 原始笔记内容
      //
      // showMdEditor.value = true;
    } else {
      Message.error('读取失败：' + res.msg);
    }
  });

}


onMounted(() => {
  Load()
})


</script>


