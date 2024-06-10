<script setup lang="ts">
import { ref, onUnmounted, onUpdated } from 'vue';
import mpegts from 'mpegts.js';

const props = defineProps({
	src: {
		type: String,
		required: true,
	},
});

const id = ref("mpegts-player-" + (new Date().getTime().toString()));
const player = ref<mpegts.Player | null>(null);

onUpdated(() => {
	let p = player.value;
	if (p) {
		p.detachMediaElement();
		p.destroy();
	}
	p = mpegts.createPlayer({
		type: 'm2ts',
		isLive: true,
		url: props.src,
	});

	const elem = document.getElementById(id.value)! as HTMLMediaElement;
	p.attachMediaElement(elem);
	p.load();
	p.play();

	player.value = p;
});

onUnmounted(() => {
	if (player.value) {
		player.value.detachMediaElement();
		player.value.destroy();
	}
});
</script>

<template>
	<video :id="id" controls></video>
</template>