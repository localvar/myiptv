<template>
  <a-menu v-model:selectedKeys="selectedKeys" mode="horizontal" theme="dark" @click="onClick">
    <a-menu-item key="watch">看电视</a-menu-item>
    <a-menu-item key="channel">频道管理</a-menu-item>
    <a-menu-item key="config">设置</a-menu-item>
    <a-menu-item key="status">状态</a-menu-item>
    <a-menu-item key="about">关于</a-menu-item>
  </a-menu>
</template>

<script setup lang="ts">
import {ref} from 'vue';
import {useRouter} from 'vue-router';

const router = useRouter();
const selectedKeys = ref<string[]>(['watch']);

const onClick = ({key}: {key: string}) => {
  selectedKeys.value = [key];
  router.push('/' + key);
}

</script>

<style scoped>
</style>
