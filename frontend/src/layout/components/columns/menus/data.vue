<template>
  <div>
    <div class="menu-title flex items-center">
      <a-space>
        <a-button type="primary" status="success" @click="() => editRef.open()">数据</a-button>
      </a-space>
    </div>



    <div>
      <a-split direction="vertical" :style="{height: '700px'}" v-model:size="size">
        <template #first>
 asdfasd


        </template>
        <template #second>
          afasdfasdf
        </template>
      </a-split>
    </div>
    <edit-task ref="editRef" />
  </div>
</template>
<script setup>

import {onMounted, ref, reactive} from 'vue';

const app = ref(window.go.gvmapp.App)



</script>
