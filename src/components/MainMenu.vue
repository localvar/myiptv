<script setup lang="ts">
</script>

<template>
  <el-scrollbar>
    <el-menu :router="true">
      <el-menu-item index="/watch">观看</el-menu-item>
      <el-menu-item index="/channel">频道</el-menu-item>
      <el-menu-item index="/config">设置</el-menu-item>
      <el-menu-item index="/status">状态</el-menu-item>
      <el-menu-item index="/about">关于</el-menu-item>
    </el-menu>
  </el-scrollbar>
</template>

<style scoped>
</style>
