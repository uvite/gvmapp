<template>
  <a-form :model="form" :style="{ width: '600px' }">
    <a-form-item field="symbol" tooltip="Please enter username" label="Username">
      <a-input
          v-model="form.symbol"
          placeholder="please enter your username..."
      />
    </a-form-item>
    <a-form-item field="post" label="Post">
      <a-input v-model="form.interval" placeholder="please enter your post..." />
    </a-form-item>
    <a-form-item>
      <a-button @click="handleSubmit">Submit</a-button>
    </a-form-item>
  </a-form>
  {{ form }}
</template>

<script setup>
import {reactive, ref} from 'vue';
import {Message} from "@arco-design/web-vue";
const app = ref(window.go.gvmapp.App)
const form = reactive({
  symbol: 'BTCUSDT',
  interval: '15m',

});
const handleSubmit = (data) => {
  console.log(data);
  app.value.AddSymbolInterval(form.symbol,form.interval).then((res) => {
    console.log(res);
    if (res && res.code == 200) {
     console.log(res)
    } else {
      Message.error('获取数据失败：' + res.msg);
    }
  });
};

</script>
