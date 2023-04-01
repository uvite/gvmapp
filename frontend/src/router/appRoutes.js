const AppPage = [
   {
        name: 'alert',
        path: '',
        meta: {
            title: '警报',
            icon: 'icon-notification',
            type: 'M',
        },
        component: () => import('@/views/dashboard/index.vue'),
        children:[]
    }
]
const test= [

    {
        name: 'data',
        path: '',
        meta: {
            title: 'data',
            icon: 'icon-notification',
            type: 'M',
        },

        children:[]
    },
    {
        name: 'deal',
        path: '/deal',
        meta: {
            title: '实盘',
            icon: 'icon-message',
            type: 'M',
        },
        component: () => import('@/views/userCenter/message.vue'),
        children:[]

    },
    {
        name: 'dashboard',
        path: '/dashboard',
        meta: {
            title: '仪表盘',
            icon: 'icon-dashboard',
            type: 'M',
            affix: true
        },
        component: () => import('@/views/dashboard/index.vue'),
        children: [
            {
                name: 'message',
                path: '/message',
                meta: {
                    title: '报警历史',
                    icon: 'icon-message',
                    type: 'M',
                },
                component: () => import('@/views/userCenter/message.vue'),
                children: []

            },
            {
                name: 'userCenter',
                path: '/usercenter',
                meta: {
                    title: '个人信息',
                    icon: 'icon-user',
                    type: 'M',
                },
                component: () => import('@/views/userCenter/index.vue'),

            },
            {
                name: 'code',
                path: '/strategies',
                meta: {
                    title: '策略源码',
                    icon: 'icon-dashboard',
                    type: 'M',
                    affix: true
                },
                component: () => import('@/views/strategies/code/index.vue'),
            }, {
                name: 'test',
                path: '/test',
                meta: {
                    title: 'test',
                    icon: 'icon-dashboard',
                    type: 'M',
                    affix: true
                },
                component: () => import('@/views/test/index.vue'),
            }]
    }

]



export const homePage = {
    name: 'home',
    path: '/home',
    meta: {title: '首页', icon: 'icon-home', hidden: false, type: 'M'}
}

export default AppPage
