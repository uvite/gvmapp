
<template>
  <div class="flex justify-between">
    asdfasf
  </div>
</template>

<script setup>
  import { ref, onMounted } from 'vue'
  import { useUserStore } from '@/store'
  import dayjs from 'dayjs'

  const userStore = useUserStore()
  const visible = ref(false)

  const time = ref(null)
  const day = ref(null)

  onMounted(() => {
    showTime()
    setInterval(() => showTime(), 1000)
  })

  const showTime = () => {
    time.value = dayjs().format('HH:mm:ss')
    day.value  = dayjs().format('YYYY年MM月DD日')
  }

  const donate = () => visible.value = true

</script>

<style scoped>
.datetime {
  background: rgb(var(--primary-6));
  color: #fff;
  width: 160px;
  text-align: center;
  border-radius: 3px;
  padding: 5px 10px;
}
</style>
