import * as app from '@/wailsjs/go/gvmapp/App'
import * as StartService from '@/wailsjs/go/services/StartService'
import * as AlertService from '@/wailsjs/go/services/AlertService'
import * as ExchangeService from '@/wailsjs/go/services/ExchangeService'
import * as LauncherService from '@/wailsjs/go/services/LauncherService'
import * as PoolService from '@/wailsjs/go/services/PoolService'

import {WindowSetTitle, EventsOnMultiple, EventsEmit} from '@/wailsjs/runtime'

const rpc = {app, StartService, AlertService, ExchangeService, LauncherService,PoolService,
    on, emit,setPageTitle}

// const rpc = { app, CollectionService, FileSystemService, SettingsService, on, setPageTitle }

function on(event: string, callback: (...data: any) => void) {
    EventsOnMultiple(event, callback, -1)
}

function emit(event: string, ...data: any) {

    EventsEmit(event, data)
}

async function setPageTitle(title: string) {
    const prefix = await app.Title()

    WindowSetTitle(`${prefix} — ${title}`)
}

export default rpc
