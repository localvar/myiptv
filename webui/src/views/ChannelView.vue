<template>
	<a-layout>
		<a-layout-sider style="background-color: lightgrey">
			<a-flex justify="space-between" style="margin: 5px 5px">
				<a-space-compact block>
					<a-button block type="primary" :icon="h(SaveOutlined)" @click="onSave">保存</a-button>
					<a-upload accept="text/plain,text/csv" :showUploadList="false" :beforeUpload="onImport">
						<a-tooltip title="导入频道列表">
							<a-button :icon="h(UploadOutlined)" />
						</a-tooltip>
					</a-upload>
					<a-tooltip title="导出频道列表">
						<a-button :icon="h(CloudDownloadOutlined)" @click="onExport" />
					</a-tooltip>
				</a-space-compact>
			</a-flex>

			<a-space-compact style="margin: 5px 5px 0px">
				<a-input placeholder="频道组名称" v-model:value="newGroupName" />
				<a-tooltip title="添加频道组">
					<a-button :icon="h(PlusOutlined)" @click="onAddGroup" />
				</a-tooltip>
				<a-tooltip title="重命名频道组">
					<a-button :icon="h(EditOutlined)" @click="onRenameGroup" />
				</a-tooltip>
				<a-tooltip title="删除频道组">
					<a-button :icon="h(DeleteOutlined)" @click="onRemoveGroup" />
				</a-tooltip>
			</a-space-compact>

			<a-radio-group style="margin: 0px 5px 10px" ref="radioGroupRef" button-style="solid" v-model:value="selectedGroup">
				<a-radio-button style="width: 100%; height: 40px; line-height: 40px" v-for="grp in groups" :value="grp" :key="grp.name">{{grp.name}}</a-radio-button>
			</a-radio-group>
		</a-layout-sider>

		<a-layout-content>
			<a-flex style="margin: 5px" justify="space-between">
				<a-space>
					<a-button :icon="h(VideoCameraAddOutlined)" @click="onCreateChannel">添加频道</a-button>
				</a-space>

				<a-space>
					<a-button-group>
						<a-button :icon="h(CheckSquareOutlined)" @click="onSelectAllChannel">全选</a-button>
						<a-button @click="onInvertSelectChannel">反选</a-button>
						<a-button @click="onUnselectAllChannel">全不选</a-button>
					</a-button-group>

					<a-button-group>
						<a-dropdown>
							<template #overlay>
								<a-menu @click="onCopyChannelTo">
								<a-menu-item v-for = "grp in groups" :key="grp.name">{{grp.name}}</a-menu-item>
								</a-menu>
							</template>
							<a-button :icon="h(CopyOutlined)">复制到</a-button>
						</a-dropdown>

						<a-dropdown>
							<template #overlay>
								<a-menu @click="onMoveChannelTo">
								<a-menu-item v-for = "grp in groups" :key="grp.name">{{grp.name}}</a-menu-item>
								</a-menu>
							</template>
							<a-button :icon="h(ScissorOutlined)">移动到</a-button>
						</a-dropdown>

						<a-button :icon="h(DeleteOutlined)" danger @click="onRemoveChannel">删除</a-button>
					</a-button-group>
				</a-space>
			</a-flex>

			<a-row style="margin: 0px 5px" ref="channelListRef" :gutter="[5,5]">
				<a-col :sm="12" :md="8" :lg="6" :xxl="4" v-for="ch in selectedGroup?.channels" :key="ch.name">
					<a-card size="small">
						<template #title>
							<a-space>
								<a-checkbox v-model:checked="ch.selected" />
								<a-typography>{{ch.displayName || ch.name}}</a-typography>
							</a-space>
						</template>

						<template #extra>
								<a-tooltip title="编辑频道">
									<a-button type="text" :icon="h(SettingOutlined)" @click="onEditChannel(ch)" />
								</a-tooltip>
						</template>

						<template #cover>
							<div style="background-color: lightgrey">
								<a-image :preview="false" v-if="ch.logo" :src="ch.logo" />
							</div>
						</template>
					</a-card>
				</a-col>
			</a-row>
		</a-layout-content>
	</a-layout>

	<ChannelDialog v-model:open="showChannelDialog" v-model:channel="currentChannel" @created="onChannelCreated" @updated="onChannelUpdated" @cancel="showChannelDialog = false" />
</template>

<script setup lang="ts">

import { ref, h, onMounted } from 'vue';
import { App } from 'ant-design-vue';
import type { MenuProps, UploadProps } from 'ant-design-vue';
import { PlusOutlined, EditOutlined, DeleteOutlined, UploadOutlined, CloudDownloadOutlined, VideoCameraAddOutlined, SettingOutlined, SaveOutlined, CopyOutlined, ScissorOutlined, CheckSquareOutlined } from '@ant-design/icons-vue';
import Sortable from 'sortablejs';

import { ChannelGroup, listChannelGroups, Channel, updateChannelGroups } from '../api/iptv';
import ChannelDialog from '../components/ChannelDialog.vue';

const { message, modal } = App.useApp();
const radioGroupRef = ref();
const channelListRef = ref();

const groups = ref<ChannelGroup[]>([]);
const selectedGroup = ref<ChannelGroup | null>(null);
const newGroupName = ref('');

const currentChannel = ref<Channel|null>(null);
const showChannelDialog = ref(false);
// const importedFiles = ref([]);

onMounted(()=>{
	new Sortable( radioGroupRef.value.$el, {
		group: 'channelGroupList',
		onEnd: (evt: Sortable.SortableEvent) => {
			if (evt.newIndex !== evt.oldIndex) {
				const item = groups.value.splice(evt.oldIndex!, 1)[0];
				groups.value.splice(evt.newIndex!, 0, item);
			}
		},
	});
	new Sortable( channelListRef.value.$el, {
		group: 'channelList',
		onEnd: (evt: Sortable.SortableEvent) => {
			if (evt.newIndex !== evt.oldIndex) {
				const item = selectedGroup.value!.channels.splice(evt.oldIndex!, 1)[0];
				selectedGroup.value!.channels.splice(evt.newIndex!, 0, item);
			}
		},
	});
});

listChannelGroups().then((data) => {
	groups.value = data || [];
	selectedGroup.value = (data.length > 0) ? data[0] : null;
});

const onCreateChannel = () => {
	currentChannel.value = null;
	showChannelDialog.value = true;
}

const onEditChannel = (ch: Channel) => {
	currentChannel.value = ch;
	showChannelDialog.value = true;
}

const onChannelCreated = (ch: Channel) => {
	const chExist = selectedGroup.value!.channels.find(c => {
		if (ch.name.toLowerCase() === c.name.toLowerCase()) {
			return true;
		}
		if (!ch.displayName) {
			return false;
		}
		return ch.displayName.toLowerCase() === c.displayName?.toLowerCase();
	});
	if (chExist) {
		message.error('频道已存在');
		return;
	}
	selectedGroup.value!.channels.push({
		...ch,
		sources: [...ch.sources],
	});
	showChannelDialog.value = false;
}

const onChannelUpdated = (ch: Channel) => {
	let chExist = selectedGroup.value!.channels.find(c => ch.name.toLowerCase() === c.name.toLowerCase());
	if (!chExist && ch.displayName) {
		chExist = selectedGroup.value!.channels.find(c => ch.displayName.toLowerCase() === c.displayName.toLowerCase());
	}

	if (chExist && chExist !== currentChannel.value) {
		message.error('频道已存在');
		return;
	}

	currentChannel.value!.name = ch.name;
	currentChannel.value!.displayName = ch.displayName;
	currentChannel.value!.logo = ch.logo;
	currentChannel.value!.hide = ch.hide;
	currentChannel.value!.sources = [...ch.sources];

	showChannelDialog.value = false;
}

const onSelectAllChannel = () => {
	selectedGroup.value!.channels.forEach((ch) => ch.selected = true);
}

const onInvertSelectChannel = () => {
	selectedGroup.value!.channels.forEach((ch) => ch.selected = !ch.selected);
}
const onUnselectAllChannel = () => {
	selectedGroup.value!.channels.forEach((ch) => ch.selected = false);
}

const putChannelToGroup = (chs: Channel[], grp: ChannelGroup) => {
	let hasDup = false;

	for(const ch of chs) {
		let name = ch.name, displayName = ch.displayName;

		while (grp.channels.find(c => c.name.toLowerCase() === name.toLowerCase())) {
			hasDup = true;
			name = '新_' + name;
		}

		if (displayName) {
			while (grp.channels.find(c => c.displayName?.toLowerCase() === displayName.toLowerCase())) {
				hasDup = true;
				displayName = '新_' + displayName;
			}
		}

		grp.channels.push({...ch, name, displayName, selected: false});
	}

	if (hasDup) {
		message.warning('重复的频道已被重命名。');
	}
}

const onCopyChannelTo: MenuProps['onClick'] = (e) => {
	if (selectedGroup.value!.name === e.key) {
		return;
	}

	const toGroup = groups.value.find((grp) => grp.name === e.key);
	if (!toGroup) {
		return;
	}

	const channels = selectedGroup.value!.channels;
	if (!channels.find(ch => ch.selected)) {
		return;
	}

	modal.confirm({
		title: '复制频道',
		content: `确定要复制选中的频道到频道组“${toGroup.name}”吗？`,
		onOk: () => {
			const selectedChannels = channels.filter(ch => ch.selected);
			putChannelToGroup(selectedChannels, toGroup);
		},
	});
}

const onMoveChannelTo: MenuProps['onClick'] = (e) => {
	if (selectedGroup.value!.name === e.key) {
		return;
	}

	const toGroup = groups.value.find((grp) => grp.name === e.key);
	if (!toGroup) {
		return;
	}

	const channels = selectedGroup.value!.channels;
	if (!channels.find(ch => ch.selected)) {
		return;
	}

	modal.confirm({
		title: '移动频道',
		content: `确定要移动选中的频道到频道组“${toGroup.name}”吗？`,
		onOk: () => {
			const selectedChannels = channels.filter(ch => ch.selected);
			selectedGroup.value!.channels = channels.filter(ch => !ch.selected);
			putChannelToGroup(selectedChannels, toGroup);
		},
	});
}

const onRemoveChannel = () => {
	const channels = selectedGroup.value!.channels;

	if (!channels.find(ch => ch.selected)) {
		return;
	}

	modal.confirm({
		title: '删除频道',
		content: '确定要删除选中的频道吗？',
		onOk: () => {
			selectedGroup.value!.channels = channels.filter(ch => !ch.selected);
		},
	});
}

const onAddGroup = () => {
	if (!newGroupName.value) {
		return;
	}
	if (groups.value.find((group) => group.name === newGroupName.value)) {
		message.error('频道组已存在');
		return;
	}
	groups.value.push({
		name: newGroupName.value,
		channels: [],
	})
	selectedGroup.value = groups.value[groups.value.length - 1];
	newGroupName.value = '';
}

const onRenameGroup = () => {
	if (!selectedGroup.value) {
		return;
	}
	if (!newGroupName.value) {
		return;
	}
	if (selectedGroup.value.name === newGroupName.value) {
		return;
	}
	if (groups.value.find((group) => group.name === newGroupName.value)) {
		message.error('频道组已存在');
		return;
	}
	selectedGroup.value.name = newGroupName.value;
	newGroupName.value = '';
}

const onRemoveGroup = () => {
	if (!selectedGroup.value) {
		return;
	}
	const name = selectedGroup.value!.name;
	modal.confirm({
		title: '删除频道组',
		content: `确定要删除频道组“${name}”吗？频道组下面的所有频道也会被删除！`,
		onOk: () => {
			const idx = groups.value.findIndex((group) => group.name === name);
			groups.value.splice(idx, 1);
			selectedGroup.value = groups.value.length > 0 ? groups.value[0] : null;
		},
	});
}

const onSave = () => {
	updateChannelGroups(groups.value).then(() => {
		message.success('保存成功');
	}).catch((err) => {
		message.error('保存失败：' + err);
	});
}

const doImport = (data: string) => {
	const lines = data.split('\n')
		.map((line) => line.trim())
		.filter((line) => line && !line.startsWith('#'));

	let hasErr = false;

	for (const line of lines) {
		const fields = line.split(',');
		if (fields.length < 6) {
			hasErr = true;
			continue;
		}

		const grpName = fields[0].trim();
		const name = fields[1].trim();
		const displayName = fields[2].trim();
		const hide = ['是', 'y', 'yes', 't', 'true'].includes(fields[3].trim().toLowerCase());
		const logo = fields[4].trim();
		const src = fields[5].trim();

		if (!grpName || !name || !src) {
			hasErr = true;
			continue;
		}

		let grp = groups.value.find((grp) => grp.name === grpName);
		if (!grp) {
			grp = {name: grpName, channels: []};
			groups.value.push(grp);
		}

		let ch = grp.channels.find((ch) => ch.name.toLowerCase() === name.toLowerCase());
		if (!ch) {
			ch = {name, displayName, hide, logo, sources: [src], selected: false};
			grp.channels.push(ch);
		} else {
			ch.displayName = displayName;
			ch.hide = hide;
			ch.logo = logo;
			if (!ch.sources.includes(src)) {
				ch.sources.push(src);
			}
		}
	}

	if (!selectedGroup.value && groups.value.length > 0) {
		selectedGroup.value = groups.value[0];
	}

	if (hasErr) {
		message.warning('导入成功，但文件格式有错误，请仔细检查导入结果。');
	} else {
		message.success('导入成功。');
	}
}

const onImport: UploadProps['beforeUpload'] = (file) => {
	return new Promise((_, reject) => {
		const reader = new FileReader();
		reader.onload = (e) => {
			doImport(e.target?.result as string);
			reject(); // reject because we don't want to upload the file
		};
		reader.readAsText(file);
	});
}

const onExport = () => {
	let data = '#频道组,频道名称,显示名称,是否隐藏,台标,节目源\n';
	groups.value.forEach((grp) => {
		grp.channels.forEach((ch) => {
			ch.sources.forEach((src) => {
				data += `${grp.name},${ch.name},${ch.displayName || ''},${ch.hide ? '是' : '否'},${ch.logo || ''},${src}\n`;
			});
		});
	});

	const a = document.createElement('a');
	a.download = 'channels.csv';
	a.href = URL.createObjectURL(new Blob([data], {type: 'text/csv'}));
	a.style.display = 'none';
	document.body.appendChild(a);
	a.click();
	a.remove();
}

</script>

<style scoped>
</style>
