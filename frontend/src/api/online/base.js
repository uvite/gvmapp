import {request} from '@/utils/request.js'
import {useBotsStore} from '@/store'
const botStore = useBotsStore()
export default {
    gvmRun(data={} ) {
        return request({
            url: `gmvBase/gvm/run`,
            method: 'post',
            data
        })
    },
    gvmRunTest(data={} ) {
        return request({
            url: `gmvBase/gvm/testcode`,
            method: 'post',
            data
        })
    },
    gvmUuid(data={} ) {
        return request({
            url: `gmvBase/gvm/uuid`,
            method: 'post',
            data
        })
    },

    getBalance(params = {}) {
        return request({
            url: `gvmBalance/getGvmBalance`,
            method: 'get',
            params
        })
    },
    getTradeList( ) {
        const params={
            exchange_id:botStore.exchangeId,
            pageSize: 10,
            orderBy: 'id',
            orderType: 'desc'
        }
        return request({
            url: `gvmTrades/getGvmTradesList`,
            method: 'get',
            params
        })
    },



}
