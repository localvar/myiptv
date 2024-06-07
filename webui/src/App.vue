<script setup lang="ts">
import { ref } from 'vue'
import zhCN from 'ant-design-vue/es/locale/zh_CN';
import dayjs from 'dayjs';
import 'dayjs/locale/zh-cn';
import MainMenu from './components/MainMenu.vue'

dayjs.locale('zh-cn');
const locale = ref(zhCN)
</script>

<template>
  <a-config-provider :locale="locale">
    <a-app>
      <a-layout>
        <a-layout-header>
            <MainMenu />
        </a-layout-header>
        <a-layout-content>
          <router-view />
        </a-layout-content>
        <a-layout-footer>Footer</a-layout-footer>
      </a-layout>
    </a-app>
  </a-config-provider>
</template>

<style scoped>
</style>
