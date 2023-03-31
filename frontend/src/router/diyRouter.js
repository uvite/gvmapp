import homePageRoutes from './homePageRoutes'
//系统路由
const routes = [
    {
        name: '策略',
        path: '/strategies',

        redirect: 'strategies',
        "meta": {
            "type": "M",
            "redirect": "",
            "icon": "IconCodeSquare",
            "activeName": "",
            "keepAlive": false,
            "defaultMenu": false,
            "title": "策略",
            "closeTab": false
        },
        children: [
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
            }]

    }

]

export default routes
