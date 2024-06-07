import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import { RouteRecordRaw, createWebHistory, createRouter } from 'vue-router'

const routes: RouteRecordRaw[] = [
  { path: '/', component: () => import('./views/HomeView.vue') },
  { path: '/watch', component: () => import('./views/WatchView.vue') },
  { path: '/channel', component: () => import('./views/ChannelView.vue') },
  { path: '/status', component: () => import('./views/StatusView.vue') },
  { path: '/config', component: () => import('./views/ConfigView.vue') },
  { path: '/about', component: () => import('./views/AboutView.vue') },
]

const router = createRouter({ history: createWebHistory(), routes })

createApp(App).use(router).mount('#app')
