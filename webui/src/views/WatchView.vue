<template>
	<a-layout>
		<a-layout-content>
			<MpegTsPlayer ref="videoPlayerRef" @error="onPlayerError" :src="source.toLowerCase().startsWith('http') ? source : `/iptv/relay/${source}`" width="100%" />
			<a-dropdown :trigger="['click']"">
				<a class="ant-dropdown-link" @click.prevent>
					<a-typography-text strong>频道：</a-typography-text>
					<a-typography-text type="secondary">{{chName}}</a-typography-text>
					<a-typography-text strong> 节目源：</a-typography-text>
					<a-typography-text type="secondary">{{source}}</a-typography-text>
					<DownOutlined />
				</a>
				<template #overlay>
					<a-menu>
						<a-sub-menu v-for="(grp, i) in groups" :key="i" :title="grp.name">
							<template v-for="(ch, j) in grp.channels">
								<a-sub-menu v-if="ch.sources.length > 1" :key="`${i}-${j}`" :title="ch.name">
									<a-menu-item v-for="(src, k) in ch.sources" :key="`${i}-${j}-${k}`" @click="watchTV(ch, k)">
										源：{{src}}
									</a-menu-item>
								</a-sub-menu>
								<a-menu-item v-else :key="`${i}-${j}-0`" @click="watchTV(ch, 0)">
									{{ch.name}}
								</a-menu-item>
							</template>
						</a-sub-menu>
					</a-menu>
				</template>
			</a-dropdown>
		</a-layout-content>
		<a-layout-sider id="programme-guide" v-if="progs.length > 0" width="260px" breakpoint="lg" collapsed-width="0" :style="{backgroundColor: '#f0f2f5', height: epgHeight}">
			<a-timeline style="overflow-y:scroll; height: 100%" >
				<a-timeline-item v-for="prog in progs" v-bind="calcProgProps(prog, now)">
					<a-typography-text strong ellipsis :content="prog.title" /> <br/>
					<a-typography-text>{{prog.start.toLocaleTimeString() + ' - ' + prog.end.toLocaleTimeString()}}</a-typography-text>
				</a-timeline-item>
			</a-timeline>
		</a-layout-sider>
	</a-layout>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue';
import { DownOutlined } from '@ant-design/icons-vue';
import MpegTsPlayer from '../components/MpegTsPlayer.vue';
import { Channel, ChannelGroup, listChannelGroups, Programme, getEpg } from '../api/iptv';

const groups = ref<ChannelGroup[]>([]);
const progs = ref<Programme[]>([]);

const chName = ref('');
const source = ref('');

const currentTimelineItemRef = ref()
const now = ref(new Date());
const timer = ref<number | null>(null);

const videoPlayerRef = ref();
const epgHeight = ref('574px');

onMounted(() => {
	window.onresize = () => {
		if (videoPlayerRef.value) {
			epgHeight.value = videoPlayerRef.value.$el.clientHeight + 'px';
		}
	}
});

onUnmounted(() => {
	window.onresize = null;
	timer.value && window.clearTimeout(timer.value);
});

const showCurrentProg = () => {
	now.value = new Date();
	
	nextTick(() => {
		const value = currentTimelineItemRef.value;
		if (value && value.length > 0 && value[value.length - 1].$el) {
			value[value.length - 1].$el.scrollIntoView();
		}
	});

	for(let prog of progs.value) {
		if (prog.start > now.value) {
			const to = prog.start.getTime() - now.value.getTime();
			timer.value = setTimeout(showCurrentProg, to);
			return;
		}
	}
}

const watchTV = function(ch: Channel, k: number) {
	chName.value = ch.name;
	source.value = ch.sources[k];

	getEpg(ch.name).then((data) => {
		progs.value = data;
		timer.value && window.clearTimeout(timer.value);
		showCurrentProg();
	});
}

listChannelGroups().then(data => {
	for (let grp of data) {
		grp.channels = grp.channels.filter(ch => ch.sources.length > 0);
	}
	data = data.filter(grp => grp.channels.length > 0);
	groups.value = data;
	if (data.length > 0) {
		watchTV(groups.value[0].channels[0], 0);
	}
});

const calcProgProps = (prog: Programme, now: Date) => {
	const props = {
		color: '#32CD32',
		ref: '',
	}

	if(prog.end < now) {
		props.color = '#D3D3D3';
	} else if (prog.start > now) {
		props.color = '#8470FF';
	} else {
		props.ref = 'currentTimelineItemRef';
	}

	return props;
}

const onPlayerError = (e: Error) => {
	console.log("Player error: ", e);
}

</script>

<style scoped>
</style>