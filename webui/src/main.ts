import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import { RouteRecordRaw, createWebHistory, createRouter } from 'vue-router'

const routes: RouteRecordRaw[] = [
  { path: '/', redirect: '/watch' },
  { path: '/watch', component: () => import('./views/WatchView.vue') },
  { path: '/channel', component: () => import('./views/ChannelView.vue') },
  { path: '/status', component: () => import('./views/StatusView.vue') },
  { path: '/config', component: () => import('./views/ConfigView.vue') },
  { path: '/about', component: () => import('./views/AboutView.vue') },
  { path: '/:pathMatch(.*)', component: () => import('./views/NotFound.vue') },
]

const router = createRouter({ history: createWebHistory(), routes })

createApp(App).use(router).mount('#app')
