<template>
	<a-modal :title="props.channel ? '编辑频道 - ' + props.channel.name : '新频道'" width="560px" @ok="onOk" :afterClose="stopVideoPlayer">
		<a-tabs @change="stopVideoPlayer">
			<a-tab-pane key="general" tab="基本信息">
				<a-space align="start">
					<a-space-compact direction="vertical">
						<a-form-item label="频道名称">
							<a-input v-model:value="ch.name" />
						</a-form-item>
						<a-form-item label="显示名称">
							<a-input v-model:value="ch.displayName" />
						</a-form-item>
						<a-form-item label="是否隐藏">
							<a-switch v-model:checked="ch.hide" />
						</a-form-item>
					</a-space-compact>
					<div style="text-align: center; width: 250px; height: 150px; background-color: lightgray">
						<a-image v-if="ch.logo" style="margin: 0 auto; height: 150px; object-fit: scale-down;" :src="ch.logo" alt="无法加载台标" />
					</div>
				</a-space>
				<a-form-item label="台标地址">
					<a-input v-model:value="ch.logo" />
				</a-form-item>
			</a-tab-pane>
			<a-tab-pane key="sources" tab="节目源">
				<MpegTsPlayer ref="videoPlayerRef" :src="source" width="98%" />
				<source-list v-model:sources="ch.sources" @verify="onVerifySource" />
			</a-tab-pane>>
		</a-tabs>
	</a-modal>
</template>

<script setup lang="ts">
import { ref, onUpdated } from 'vue';
import { App } from 'ant-design-vue';
import { Channel } from '../api/iptv';
import MpegTsPlayer from './MpegTsPlayer.vue';

const { message } = App.useApp();

const newChannel = () => {
	return {
		name: '',
		displayName: '',
		logo: '',
		hide: false,
		sources: [],
	} as Channel;
};

const emit = defineEmits<{
	(e: 'created', ch: Channel): void;
	(e: 'updated', ch: Channel): void;
}>();

const props = defineProps<{channel: Channel | null}>();
const ch = ref<Channel>(newChannel());
const source = ref<string>('');
const videoPlayerRef = ref();

onUpdated(() => {
	if (props.channel) {
		ch.value = { ...props.channel, sources: [...props.channel.sources] };
	} else {
		ch.value = newChannel();
	}
});

const onVerifySource = (src: string) => {
	src && (source.value = `/iptv/relay/${src}`);
};

const stopVideoPlayer = () => {
	source.value = '';
	videoPlayerRef.value?.stop();
};

const onOk = () => {
	if (!ch.value.name) {
		message.error('频道名称不能为空');
		return;
	}

	if (props.channel) {
		emit('updated', ch.value);
	} else {
		emit('created', ch.value);
	}
};

</script>

<style scoped>
</style>