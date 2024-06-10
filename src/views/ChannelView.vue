<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { VideoPlay, Delete, CirclePlus, ArrowUp, ArrowDown, Top, Bottom } from '@element-plus/icons-vue'
import { Channel, ChannelGroup, listChannelGroups } from '../api/iptv';
import MpegTsPlayer from '../components/MpegTsPlayer.vue';
// import Sortable from 'sortablejs';

const groups = ref<ChannelGroup[]>([]);
const newGroupName = ref('');
const source = ref('');
const newSource = ref('');

listChannelGroups().then((data) => {
	groups.value = data;
});

onMounted(()=>{
	/*
	// TODO: this works but not always correct
	const tab = document.querySelector('#tab-channel-group .el-tabs__nav');
	Sortable.create(tab, {
		filter: '.new-channel-group-pane',
		onEnd: (evt: any) => {
			const group = groups.value[evt.oldIndex];
			groups.value.splice(evt.oldIndex, 1);
			groups.value.splice(evt.newIndex, 0, group);
		},
	});
	*/
})

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
	const idx = groups.value.findIndex((group) => group.name === name);
	if (idx >= 0) {
		groups.value.splice(idx, 1);
	}
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

<template>
	<el-row :gutter="10">
		<el-col :span="14">
			<el-tabs type="border-card" @tab-remove="onRemoveGroup" id="tab-channel-group">
				<el-tab-pane v-for="grp in groups" :label="grp.name" :key="grp.name" closable>
					<el-tabs tab-position="left">
						<el-tab-pane v-for="ch in grp.channels" :label="ch.name" :key="ch.name">
							<el-card>
								<el-form>
									<el-row :gutter="10">
										<el-col :span="12">
											<el-form-item label="频道名称">
												<el-input v-model="ch.name" />
											</el-form-item>
											<el-form-item label="显示名称">
												<el-input v-model="ch.displayName" />
											</el-form-item>
											<el-form-item label="是否隐藏">
												<el-switch v-model="ch.hide" />
											</el-form-item>
										</el-col>
										<el-col :span="12">
											<el-card style="width: 98%; height: 85%; background-color: lightgray;" shadow="never">
												<img v-if="ch.logo" style="width: 100%; height: 100%" :src="ch.logo"
													fit="scale-down" alt="无法加载台标" />
											</el-card>
										</el-col>
									</el-row>
									<el-row>
										<el-col :span="24">
											<el-form-item label="台标地址">
												<el-input v-model="ch.logo" />
											</el-form-item>
										</el-col>
									</el-row>
									<el-divider>节目源</el-divider>
									<el-row :gutter="10" v-for="(_, i) in ch.sources">
										<el-col :span="16">
											<el-form-item>
												<el-input v-model="ch.sources[i]" />
											</el-form-item>
										</el-col>
										<el-col :span="8">
											<el-button type="success" :icon="VideoPlay" link @click="onVerifySource(ch.sources[i])" />
											<el-button-group>
												<el-button type="primary" :icon="Top" link @click="onMoveSource(ch, i, 'top')" />
												<el-button type="primary" :icon="ArrowUp" link @click="onMoveSource(ch, i, 'up')" />
												<el-button type="primary" :icon="ArrowDown" link @click="onMoveSource(ch, i, 'down')" />
												<el-button type="primary" :icon="Bottom" link @click="onMoveSource(ch, i, 'bottom')" />
											</el-button-group>
											<el-button type="danger" :icon="Delete" link @click="onDeleteSource(ch, i)" />
										</el-col>
									</el-row>
									<el-row :gutter="10">
										<el-col :span="16">
											<el-form-item>
												<el-input v-model="newSource" />
											</el-form-item>
										</el-col>
										<el-col :span="8">
											<el-button-group>
												<el-button type="success" :icon="VideoPlay" link @click="onVerifySource(newSource)" />
												<el-button type="primary" :icon="CirclePlus" link @click="onAddSource(ch)" />
											</el-button-group>
										</el-col>
									</el-row>
								</el-form>
							</el-card>
						</el-tab-pane>
					</el-tabs>
				</el-tab-pane>
				<el-tab-pane label="+" class="new-channel-group-pane">
					<el-card>
						<el-form>
							<el-row :gutter="10">
								<el-col :span="8">
									<el-input v-model="newGroupName" placeholder="新频道组的名称" />
								</el-col>
								<el-col :span="4">
									<el-button type="primary" @click="onAddGroup">创建</el-button>
								</el-col>
							</el-row>
						</el-form>
					</el-card>
				</el-tab-pane>
			</el-tabs>
		</el-col>
		<el-col :span="10">
			<MpegTsPlayer :src="source" width="100%" />
		</el-col>
	</el-row>
</template>
