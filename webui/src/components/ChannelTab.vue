<template>
	<a-tabs ref="tabsRef" tab-position="left" type="editable-card" @edit="onEditChannel">
		<a-tab-pane v-for="(ch, j) in channels" :tab="ch.name" :key="j">
			<a-row :gutter="10">
				<a-col :span="12">
					<a-form-item label="频道名称" :labelCol="{span: 6}">
						<a-typography-text v-model:content="clonedNames[j]" :editable="{onChange: onNameChange, onEnd: () => onNameUpdated(ch)}" />
					</a-form-item>
					<a-form-item label="显示名称" :labelCol="{span: 6}">
						<a-typography-text v-model:content="clonedDisplayNames[j]" :editable="{onChange: onDisplayNameChange, onEnd: () => onDisplayNameUpdated(ch)}" />
					</a-form-item>
					<a-form-item label="是否隐藏" :labelCol="{span: 6}">
						<a-switch v-model:checked="ch.hide" />
					</a-form-item>
				</a-col>
				<a-col :span="12">
					<a-card style="width: 98%; height: 85%; background-color: lightgray;" shadow="never">
						<img v-if="ch.logo" style="width: 100%; height: 100%" :src="ch.logo"
							fit="scale-down" alt="无法加载台标" />
					</a-card>
				</a-col>
			</a-row>

			<a-form-item label="台标地址" :labelCol="{span: 3}">
				<a-typography-text v-model:content="ch.logo" editable />
			</a-form-item>

			<source-list :sources="ch.sources" @verify="onVerifySource" />
		</a-tab-pane>
	</a-tabs>
</template>

<script setup lang="ts">
import { ref, onMounted, defineModel, computed } from 'vue';
import { App } from 'ant-design-vue';

import Sortable from 'sortablejs';
import SourceList from '../components/SourceList.vue';
import { Channel } from '../api/iptv';

const {modal, message} = App.useApp();
const channels = defineModel<Channel[]>('channels', { required: true });
const tabsRef = ref();
const emit = defineEmits<{ (e: 'verify-source', source: string): void; }>();

const clonedNames = computed(() => channels.value.map((ch) => ch.name));
const editedName = ref<string>('');
const onNameChange = (value: string) => {
	editedName.value = value;
};
const onNameUpdated = (ch: Channel) => {
	if (!editedName.value || editedName.value === ch.name) {
		return;
	}
	if (channels.value.find((ch) => ch.name === editedName.value)) {
		message.error({content: '频道名称重复！'});
		return;
	}
	ch.name = editedName.value;
};

const clonedDisplayNames = computed(() => channels.value.map((ch) => ch.displayName));
const editedDisplayName = ref<string>('');
const onDisplayNameChange = (value: string) => {
	editedDisplayName.value = value;
};
const onDisplayNameUpdated = (ch: Channel) => {
	if (editedDisplayName.value === ch.displayName) {
		return;
	}
	if (!editedDisplayName.value) {
		ch.displayName = '';
		return;
	}
	if (channels.value.find((ch) => ch.displayName === editedDisplayName.value)) {
		message.error({content: '显示名称重复！'});
		return;
	}
	ch.displayName = editedDisplayName.value;
};


const onVerifySource = (source: string) => {
	emit('verify-source', source);
};

const onEditChannel = (targetKey: string, action: string) => {
	if (action !== 'remove') {
		return
	}

	modal.confirm({
		title: '删除频道',
		content: `确定要删除频道“${targetKey}”吗？频道下面的所有节目源也会被删除！`,
		onOk: () => {
			const idx = channels.value.findIndex((ch) => ch.name === targetKey);
			if (idx >= 0) {
				channels.value.splice(idx, 1);
			}
		},
	});
};

onMounted(() => {
	const el = tabsRef.value.$el.getElementsByClassName('ant-tabs-nav-list')[0];
	new Sortable(el, {
		animation: 150,
		onEnd: (evt) => {
			if (evt.newIndex !== evt.oldIndex) {
				const item = channels.value.splice(evt.oldIndex!, 1)[0];
				channels.value.splice(evt.newIndex!, 0, item);
			}
		},
	});
});

</script>