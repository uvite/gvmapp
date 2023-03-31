import { defineStore } from 'pinia'
import tool from "@/utils/tool";

const useGenvStore = defineStore('bots', {

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
    setGenv(data) {
      tool.local.set("genv_", data)
    },
    getGenv() {
      return tool.local.get("genv_")
    },
  }
})

export default useGenvStore
