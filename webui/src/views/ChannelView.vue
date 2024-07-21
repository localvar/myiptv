<template>
	<a-row :gutter="10">
		<a-col :span="14">
			<a-tabs type="editable-card" @edit="onEditChannelGroup" hide-add>
				<a-tab-pane v-for="(grp, i) in groups" :tab="grp.name" :key="i" closable>
					<channel-tab :channels="grp.channels" @verify-source="onVerifySource" />
				</a-tab-pane>
			</a-tabs>
		</a-col>
		<a-col :span="10">
			<a-affix :offset-top="120">
				<MpegTsPlayer :src="source" width="98%" />
				<a-flex justify="space-evenly">
					<a-button>导入频道列表</a-button>
					<a-button>导出频道列表</a-button>
				</a-flex>
			</a-affix>
		</a-col>
	</a-row>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { App } from 'ant-design-vue';

import { /*Channel,*/ ChannelGroup, listChannelGroups } from '../api/iptv';
import MpegTsPlayer from '../components/MpegTsPlayer.vue';
import ChannelTab from '../components/ChannelTab.vue';

const groups = ref<ChannelGroup[]>([]);
const newGroupName = ref('');
const source = ref('');
const {modal} = App.useApp();

listChannelGroups().then((data) => {
	groups.value = data;
});

onMounted(()=>{
})

const onEditChannelGroup = (targetKey: string, action: string) => {
	if (action === 'add') {
		onAddGroup();
	} else if (action === 'remove') {
		onRemoveGroup(targetKey);
	}
}

const onAddGroup = () => {
	if (!newGroupName.value) {
		return;
	}
	for (const group of groups.value) {
		if (group.name === newGroupName.value) {
			// TODO: alert
			return;
		}
	}
	groups.value.push({
		name: newGroupName.value,
		channels: [],
	});
	newGroupName.value = '';
}

const onRemoveGroup = (name: string) => {
	modal.confirm({
		title: '删除频道组',
		content: `确定要删除频道组“${name}”吗？频道组下面的所有频道也会被删除！`,
		onOk: () => {
			const idx = groups.value.findIndex((group) => group.name === name);
			if (idx >= 0) {
				groups.value.splice(idx, 1);
			}
		},
	});
}

const onVerifySource = (src: string) => {
	src && (source.value = `/iptv/relay/${src}`);
}

</script>

<style scoped>
.ant-form-item {
	margin-bottom: 10;
}
</style>