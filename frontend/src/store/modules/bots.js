import { defineStore } from 'pinia'
import tool from "@/utils/tool";

const useBotsStore = defineStore('bots', {

  state: () => ({
    auth: undefined,
    appId: undefined,
    appSecret: undefined,
    exchangeId: undefined,
    globalParams: undefined,
  }),

  getters: {
    setDoc(state) {
      return { ...state };
    },
  },

  actions: {
    setInfo(data) { this.$patch(data) },
    setBotToken(data) {
      tool.local.set(import.meta.env.VITE_BOT_TOKEN_PREFIX, data)
    },
    getBotToken() {
      return tool.local.get(import.meta.env.VITE_BOT_TOKEN_PREFIX)
    },
  }
})

export default useBotsStore
