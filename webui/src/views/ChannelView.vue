<template>
	<a-row :gutter="10">
		<a-col :span="14">
			<a-tabs type="editable-card" @edit="onEditChannelGroup">
				<a-tab-pane v-for="grp in groups" :tab="grp.name" :key="grp.name" closable>
					<a-tabs tab-position="left">
						<a-tab-pane v-for="ch in grp.channels" :tab="ch.name" :key="ch.name">
							<a-form>
								<a-row :gutter="10">
									<a-col :span="12">
										<a-form-item label="频道名称">
											<a-input v-model:value="ch.name" />
										</a-form-item>
										<a-form-item label="显示名称">
											<a-input v-model:value="ch.displayName" />
										</a-form-item>
										<a-form-item label="是否隐藏">
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
								<a-row>
									<a-col :span="24">
										<a-form-item label="台标地址">
											<a-input v-model:value="ch.logo" />
										</a-form-item>
									</a-col>
								</a-row>
								<a-divider>节目源</a-divider>
								<a-row :gutter="10" v-for="(_, i) in ch.sources">
									<a-col :span="16">
										<a-form-item>
											<a-input v-model:value="ch.sources[i]" />
										</a-form-item>
									</a-col>
									<a-col :span="8">
										<a-button :icon="h(PlaySquareOutlined)" @click="onVerifySource(ch.sources[i])" />
										<a-button :icon="h(DeleteOutlined)" @click="onDeleteSource(ch, i)" />
									</a-col>
								</a-row>
								<a-row :gutter="10">
									<a-col :span="16">
										<a-form-item>
											<a-input v-model:value="newSource" />
										</a-form-item>
									</a-col>
									<a-col :span="8">
										<a-button-group>
											<a-button :icon="h(PlaySquareOutlined)" @click="onVerifySource(newSource)" />
											<a-button :icon="h(PlusSquareOutlined)" @click="onAddSource(ch)" />
										</a-button-group>
									</a-col>
								</a-row>
							</a-form>
						</a-tab-pane>
					</a-tabs>
				</a-tab-pane>
			</a-tabs>
		</a-col>
		<a-col :span="10">
			<MpegTsPlayer :src="source" width="100%" />
		</a-col>
	</a-row>
</template>

<script setup lang="ts">
import { h, ref, onMounted } from 'vue';
import { DeleteOutlined, PlaySquareOutlined, PlusSquareOutlined } from '@ant-design/icons-vue';
import {App} from 'ant-design-vue';

import { Channel, ChannelGroup, listChannelGroups } from '../api/iptv';
import MpegTsPlayer from '../components/MpegTsPlayer.vue';
// import Sortable from 'sortablejs';

const groups = ref<ChannelGroup[]>([]);
const newGroupName = ref('');
const source = ref('');
const newSource = ref('');
const {modal} = App.useApp();

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

const onDeleteSource = (ch: Channel, idx: number) => {
	// TODO: confirm
	ch.sources.splice(idx, 1);
}

const onMoveSource = (ch: Channel, idx: number, to: string) => {
	switch (to) {
		case 'top':
			if(idx >0) {
				const src = ch.sources[idx];
				ch.sources.splice(idx, 1);
				ch.sources.unshift(src);
			}
			break;

		case 'up':
			if (idx > 0) {
				const src = ch.sources[idx];
				const dst = ch.sources[idx - 1];
				ch.sources[idx - 1] = src;
				ch.sources[idx] = dst;
			}
			break;

		case 'down':
			if (idx < ch.sources.length - 1) {
				const src = ch.sources[idx];
				const dst = ch.sources[idx + 1];
				ch.sources[idx + 1] = src;
				ch.sources[idx] = dst;
			}
			break;

		case 'bottom':
			if (idx < ch.sources.length - 1) {
				const src = ch.sources[idx];
				ch.sources.splice(idx, 1);
				ch.sources.push(src);
			}
			break;
	}
}

const onAddSource = (ch: Channel) => {
	if (!newSource.value) {
		return;
	}
	for (const src of ch.sources) {
		if (src === newSource.value) {
			// TODO: alert
			return
		}
	}
	ch.sources.push(newSource.value);
	newSource.value = '';
}

</script>
