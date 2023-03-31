<template>
  <div :style="{  height: '100%'}">
    <a-split :style="{
        height: '100%',
        width: '100%',
        minWidth: '500px',
        border: '1px solid var(--color-border)'
      }"
             v-model:size="size"
             min="80px"
    >
      <template #first>
        <a-typography-paragraph>
<!--          <a-space>-->
<!--            <a-button type="primary" @click="syncToCloud">同步云端</a-button>-->
<!--            <a-button @click="downToLocal">下载本地</a-button>-->

<!--          </a-space>-->
<!--          <a-button type="primary" @click="openAddForm">-->
<!--            <template #icon>-->
<!--              <icon-plus />-->
<!--            </template>-->
<!--          </a-button>-->
          <a-menu
              :style="{ width: '200px', height: '100%' }"
              :default-open-keys="['0']"
              :default-selected-keys="[]"
              show-collapse-button
              :selected-keys="[currentNote]"

              :menu-item-click="notebookSelect"

          >
            <a-menu-item   :index="nb" v-for="nb in notebooks"   @click="notebookSelect(nb)">{{nb}}</a-menu-item>

          </a-menu>

        </a-typography-paragraph>
      </template>
      <template #second>
        <div  :style="{  height: '100%'}">
          <a-split direction="horizontal" :style="{height: '100%'}"  v-model:size="size2">
            <template #first>
              <a-layout style="height: 100%;">
<!--                <a-layout-header> <a-space>-->

<!--                  <a-button type="primary" @click="newNote">-->
<!--                    <template #icon>-->
<!--                      <icon-plus />-->
<!--                    </template>-->
<!--                  </a-button>-->
<!--                  <a-button type="primary" @click="removeNote">-->
<!--                    <template #icon>-->
<!--                      <icon-minus />-->
<!--                    </template>-->
<!--                  </a-button>-->

<!--                </a-space></a-layout-header>-->
                <a-layout-content> <a-menu
                    :style="{ width: '200px', height: '100%' }"
                    :default-open-keys="[]"
                    :default-selected-keys="[]"
                    show-collapse-button
                    breakpoint="xl"

                >
                  <a-menu-item   :index="note" v-for="note in notes"   @click="noteSelect(note)">{{note}}</a-menu-item>



                </a-menu></a-layout-content>

              </a-layout>

            </template>
            <template #second>
              <div  :style="{  height: '100%'}">




                <vue-monaco-editor
                    v-model:value="code"
                    theme="vs-dark"
                    @change="handleChange"
                />



              </div>
            </template>
          </a-split>
        </div>
      </template>
    </a-split>


    <a-modal v-model:visible="addFormVisible" title="Modal Form" @cancel="addFormCancel" @before-ok="addFormOk">
      <a-form :model="form">
        <a-form-item field="name" label="Name">
          <a-input v-model="newNotebookName" />
        </a-form-item>

      </a-form>
    </a-modal>
  </div>
</template>

<script lang="ts" setup>

import {ref, reactive, onMounted, watch} from 'vue'
import {Message} from '@arco-design/web-vue';

import { emitter } from '@/utils/bus.js'
import {useStorage} from '@vueuse/core'

import {StorageName, generateHTML, useDarkGlobal} from './utils'

// import { defineComponent ,ref} from 'vue';
// import { Plus, Minus, Download, UploadFilled } from '@element-plus/icons-vue'
// import { MessageBox, Message } from 'element-plus'
const app = ref(window.go.gvmapp.App)
const notebooks = ref([])
const notes = ref([])
const currentNotebook = ref()
const currentNote = ref()
const mdTitle = ref()
const mdText = ref()
const addFormVisible = ref(false)
const newNotebookName = ref()
const showMdEditor = ref(false)
const notEditedMdtext = ref()
const newNoteMdFileName = ref()
const size=ref(0.15)
const size2=ref(0.15)

const code = ref('// some code...')


const getNotebookDateForJG = () => {

  app.value.GetJGTestDir().then((res) => {
    console.log(res);
    if (res && res.code == 200) {
      notebooks.value = res.data;
    } else {
      Message.error('获取云端数据失败：' + res.msg);
    }
  });
}
onMounted(() => {
  getNotebookDateForJG();
})


// 笔记本被选中
const notebookSelect = (key) => {


  var oldSelectNotebook = currentNotebook.value;
  console.log("选中key：" + key,currentNotebook.value);
  // console.log(keyPath);
  if (mdText.value && key != oldSelectNotebook && mdText.value != notEditedMdtext.value) {  // 编辑器有文本

    Message.info('已丢弃') ;
    mdText.value = '';
    showMdEditor.value = false;
    currentNote.value = '';


    // MessageBox.confirm(
    //     '编辑器文本未保存，是否丢弃？',
    //     '警告',
    //     {
    //       confirmButtonText: '丢弃',
    //       cancelButtonText: '取消',
    //       type: 'warning',
    //     }
    // ).then(() => {
    //   Message.info('已丢弃') ;
    //    mdText.value = '';
    //   showMdEditor.value = false;
    //   currentNote.value = '';
    // }).catch(() => {
    //   Message.info( '操作取消');
    //   currentNotebook.value = oldSelectNotebook;
    //   return;
    // });
  }
  currentNotebook.value = key;

  // 调用go方法获取笔记本下的全部笔记
  app.value.GetJGTestDirFile(currentNotebook.value).then((res) => {
    console.log(res);
    if (res && res.code == 200) {
      var data = res.data;
       notes.value = data.map(n => n.replace('.md', ''));
    } else {
      Message.error('获取数据失败：' + res.msg);
    }
  });
}

// 笔记被选中
const noteSelect = (key, keyPath) => {
  console.log("选中key：" + key);
  // console.log(keyPath);
  // key = key.replace('.md', '');
  currentNote.value = key;
  mdTitle.value=key;
  app.value.ReadNoteFile(currentNotebook.value, currentNote.value).then((res) => {
    // console.log(res);
    if (res && res.code == 200) {
      var data = res.data;
       mdText.value = data;
       console.log("[1]",data)

      code.value=data
       notEditedMdtext.value = data;    // 原始笔记内容

       showMdEditor.value = true;
    } else {
      Message.error('读取失败：' + res.msg);
    }
  });

}

// 打开弹出表单
const openAddForm = () => {
  newNotebookName.value = '';
  addFormVisible.value = true;
}

// 弹出表单取消
const addFormCancel = () => {
  addFormVisible.value = false;
  newNotebookName.value = '';
}

// 弹出表单确认,创建笔记本
const addFormOk = () => {
  // 创建笔记本
  var bookName = newNotebookName.value.trim();
  app.value.CreateNotebook(bookName).then((res) => {
    // console.log(res);
    if (res && res.code == 200) {
      Message.success('笔记本创建成功');
      notebooks.value.push(bookName);
      currentNotebook.value = bookName;
      newNotebookName.value = '';
      addFormVisible.value = false;
    } else {
      Message.error('笔记本创建失败：' + res.msg);
    }
  });
}

// 新建笔记
const newNote = () => {
  // console.log('newNote...');
  if (currentNotebook.value) {
    app.value.CreateNoteFile(currentNotebook.value).then((res) => {
      console.log(res);
      if (res && res.code == 200) {
        var fileName = res.data;
        Message.success('笔记创建成功：' + fileName);
        newNoteMdFileName.value = fileName;
        notes.value.push(fileName);
        currentNote.value = fileName;
        // console.log('currentNote==', currentNote.value);
      } else {
        Message.error('笔记创建失败：' + res.msg);
      }
    });
  }
}
const handleChange =()=>{
  console.log(currentNotebook.value , currentNote.value)
  console.log(mdTitle.value, code.value)
  saveMdText()
}
// 保存文章/笔记
const saveMdText = ( ) => {
  return
  // console.log(MdText);
  // console.log(html);
  if (!mdTitle.value) {    //没有标题
    return;
  }
  if (currentNotebook.value && currentNote.value) { // 不为空
    // 调用go方法保存笔记
    app.value.SaveNote(currentNotebook.value, newNoteMdFileName.value, mdTitle.value, code.value).then((res) => {
      if (res && res.code == 200) {
      //  Message.success("笔记保存成功");

        // 重新获取笔记列表
        app.value.GetJGTestDirFile(currentNotebook.value).then((res) => {
          if (res && res.code == 200) {
            var data = res.data;
            notes.value = data.map(n => n.replace('.md', ''));
          } else {
            Message.error('获取笔记列表失败：' + res.msg);
          }
        });
      } else {
        Message.error('笔记保存失败：' + res.msg);
      }
    });
  } else {  // 存在为空情况
    Message.error('');
  }
}

// 删除笔记
const removeNote = () => {
  console.log('删除笔记');
  if (currentNotebook.value == '' || currentNote.value == '') {
    Message.warning("请选择笔记");
    return;
  }

  app.value.RemoveNote(currentNotebook.value, currentNote.value).then((res) => {
    if (res && res.code == 200) {
      Message.info('删除成功');
      // 调用go方法重新获取笔记本下的全部笔记
      app.value.GetJGTestDirFile(currentNotebook.value).then((res) => {
        if (res && res.code == 200) {
          var data = res.data;
          notes.value = data.map(n => n.replace('.md', ''));
        } else {
          Message.error('获取数据失败：' + res.msg);
        }
      });
    } else {
      Message.error('删除失败：' + res.msg);
    }
  });
  // MessageBox.confirm(
  //     `确认删除 ${currentNotebook.value} / ${currentNote.value} 笔记吗？`,
  //     '警告',
  //     {
  //       confirmButtonText: '删除',
  //       cancelButtonText: '取消',
  //       type: 'warning',
  //     }
  // ).then(() => {
  //   // 点击了确定，开始删除
  //
  // }).catch(() => {
  //   Message.info( '取消操作');
  // });
}

// 同步到云端
const syncToCloud = () => {
  app.value.SyncToCloud().then((res) => {
    if (res && res.code == 200) {
      Message.info('同步成功');
    } else {
      Message.error('同步失败：' + res.msg);
    }
  });
}

// 同步到本地

const downToLocal = () => {
  app.value.DownToLocal().then((res) => {
    if (res && res.code == 200) {
      Message.info('同步成功');
    } else {
      Message.error('同步失败：' + res.msg);
    }
  });
}


</script>

<style>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
