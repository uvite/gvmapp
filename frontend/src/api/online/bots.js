import {request} from '@/utils/request.js'
import {useBotsStore} from '@/store'

const pre = "gvmBots"
const module = "GvmBots"
const botStore = useBotsStore()
function getRandom(n, m) {
    var num = Math.floor(Math.random() * (m - n + 1) + n)
    return num
}

function singBot(obj) {
    let data = {
        app_id: botStore.appId,
        app_secret: botStore.appSecret,
        exchange_id: botStore.exchange_id,
    }
    let dd=Object.assign(data, obj);
    console.log("[231]",data,dd)
    return data
}

export default {
    /**
     * @returns
     */
    vmRun(data = {}) {
        data = singBot(data)
        console.log(data)
        return request({
            url: `gmvBase/vm/file`,
            method: 'post',
            data
        })
    },

    /**
     * @returns
     */
    getPageList(params = {}) {
        return request({
            url: `${pre}/get${module}List`,
            method: 'get',
            params
        })
    },

    getAppSecret() {
        var id = getRandom(1000, 9999)
        return request({
            url: `${pre}/get${module}AppSecret`,
            method: 'get'
        })
    },

    /**
     * @returns
     */
    getStrategies(params = {}) {
        return request({
            url: `${pre}/getStrategiesList`,
            method: 'get',
            params
        })
    },

    /**
     * 获取列表
     * @returns
     */
    getList(params = {}) {
        return request({
            url: `${pre}/get${module}List`,
            method: 'get',
            params
        })
    },

    /**
     * 从回收站获取
     * @returns
     */
    getRecyclePageList(params = {}) {
        return request({
            url: '',
            method: 'get',
            params
        })
    },

    /**
     * 添加
     * @returns
     */
    save(params = {}) {
        return request({
            url: `${pre}/create${module}`,
            method: 'post',
            data: params
        })
    },

    /**
     * 移到回收站
     * @returns
     */
    deletes(data) {
        return request({
            url: `${pre}/delete${module}ByIds`,
            method: 'delete',
            data
        })
    },

    /**
     * 恢复数据
     * @returns
     */
    recoverys(data) {
        return request({
            url: 'systemPost/recovery',
            method: 'put',
            data
        })
    },

    /**
     * 真实删除
     * @returns
     */
    realDeletes(data) {
        return request({
            url: 'systemPost/realDelete',
            method: 'delete',
            data
        })
    },

    /**
     * 更新数据
     * @returns
     */
    update(id, data = {}) {
        return request({
            url: `${pre}/update${module}`,
            method: 'put',
            data
        })
    },

    /**
     * 数字运算操作
     * @returns
     */
    numberOperation(data = {}) {
        return request({
            url: `${pre}/changeSort`,
            method: 'put',
            data
        })
    },

    /**
     * 更改状态
     * @returns
     */
    changeStatus(data = {}) {
        return request({
            url: `${pre}/changeStatus`,
            method: 'put',
            data
        })
    },

    login(params = {}) {
        return request({
            url: `${pre}/login`,
            method: 'get',
            params
        })
    },

    backTest(params = {}) {
        return request({
            url: `${pre}/backtest${module}`,
            method: 'post',
            data: params
        })
    },

}
