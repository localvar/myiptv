<script setup lang="ts">
import { ref, onUnmounted, onUpdated } from 'vue';
import mpegts from 'mpegts.js';

const props = defineProps({
	src: {
		type: String,
		required: true,
	},
});

const videoRef = ref<HTMLMediaElement>();
const player = ref<mpegts.Player | null>(null);

const closePlayer = () => {
	if (player.value) {
		player.value.detachMediaElement();
		player.value.destroy();
		player.value = null;
	}
};

onUpdated(() => {
	closePlayer();

	const p = mpegts.createPlayer({
		type: 'm2ts',
		isLive: true,
		url: props.src,
	});

	p.attachMediaElement(videoRef.value!);
	p.load();
	p.play();

	player.value = p;
});

onUnmounted(closePlayer);
</script>

<template>
	<video ref="videoRef" controls></video>
</template>