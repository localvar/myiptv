<template>
	<a-row :gutter="10">
		<a-col :span="14">
			<a-tabs type="editable-card" @edit="onEditChannelGroup">
				<a-tab-pane v-for="(grp, i) in groups" :tab="grp.name" :key="i" closable>
					<a-tabs tab-position="left">
						<a-tab-pane v-for="(ch, j) in grp.channels" :tab="ch.name" :key="j">
							<a-row :gutter="10">
								<a-col :span="12">
									<a-form-item label="频道名称" :labelCol="{span: 6}">
										<a-input v-model:value="ch.name" />
									</a-form-item>
									<a-form-item label="显示名称" :labelCol="{span: 6}">
										<a-input v-model:value="ch.displayName" />
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
								<a-input v-model:value="ch.logo" />
							</a-form-item>

							<source-list :sources="ch.sources" :newSource="newSource" @verify="onVerifySource" />
						</a-tab-pane>
					</a-tabs>
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

import { Channel, ChannelGroup, listChannelGroups } from '../api/iptv';
import MpegTsPlayer from '../components/MpegTsPlayer.vue';
import SourceList from '../components/SourceList.vue';


const groups = ref<ChannelGroup[]>([]);
const newGroupName = ref('');
const source = ref('');
const newSource = ref('');
const {modal} = App.useApp();

// const selectedNode = ref<TreeNode>();

listChannelGroups().then((data) => {
	groups.value = data;
});

onMounted(()=>{
	/*
	// TODO: this works but not always correct
	const tab = document.querySelector('#tab-channa-group .a-tabs__nav');
	Sortable.create(tab, {
		filter: '.new-channa-group-pane',
		onEnd: (evt: any) => {
			const group = groups.value[evt.oldIndex];
			groups.value.splice(evt.oldIndex, 1);
			groups.value.splice(evt.newIndex, 0, group);
		},
	});
	*/
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
		content: `确定要删除该频道组“${name}”吗？频道组下面的所有频道也会被删除！`,
		onOk: () => {
			const idx = groups.value.findIndex((group) => group.name === name);
			if (idx >= 0) {
				groups.value.splice(idx, 1);
			}
		},
	});
}

const onVerifySource = (src: string) => {
	if (!src) {
		return;
	}
	source.value = `/iptv/relay/${src}`;
}

</script>

<style scoped>
.ant-form-item {
	margin-bottom: 10;
}
</style>