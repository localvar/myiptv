<script setup lang="ts">
import { ref } from 'vue'
import zhCN from 'element-plus/es/locale/lang/zh-cn'
// import en from 'element-plus/es/locale/lang/en'
import MainMenu from './components/MainMenu.vue'

const locale = ref(zhCN)
</script>

<template>
  <el-config-provider :locale="locale">
    <el-container>
      <el-header>
          <MainMenu />
      </el-header>
        <el-main>
          <RouterView />
        </el-main>
      <el-footer>Footer</el-footer>
    </el-container>
  </el-config-provider>
</template>

<style scoped>
</style>
