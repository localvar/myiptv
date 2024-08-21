<script setup lang="ts">
import { ref, onUnmounted, onUpdated, defineExpose } from 'vue';
import mpegts from 'mpegts.js';

const emit = defineEmits(['error']);

const props = defineProps({
	src: {
		type: String,
		required: true,
	},
});

const videoRef = ref<HTMLMediaElement>();
const player = ref<mpegts.Player | null>(null);

const onError = (e: Error) => {
	emit('error', e);
	closePlayer();
};

const closePlayer = () => {
	if (player.value) {
		player.value.off(mpegts.Events.ERROR, onError);
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

	p.on(mpegts.Events.ERROR, onError);
	p.attachMediaElement(videoRef.value!);
	p.load();
	p.play();

	player.value = p;
});

onUnmounted(closePlayer);

defineExpose({ stop: closePlayer });
</script>

<template>
	<video ref="videoRef" controls></video>
</template>