<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue';
import MpegTsPlayer from '../components/MpegTsPlayer.vue';
import { Channel, ChannelGroup, listChannelGroups, Programme, getEpg } from '../api/iptv';

const groups = ref<ChannelGroup[]>([]);
const progs = ref<Programme[]>([]);

const chName = ref('');
const source = ref('');

const currentTimelineItemRef = ref()
const now = ref(new Date());
const timer = ref<number | null>(null);

const videoContainerRef = ref();
const epgHeight = ref(580);

onMounted(() => {
	window.onresize = () => {
		const v = videoContainerRef.value;
		v && (epgHeight.value = v.$el.clientHeight);
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

listChannelGroups().then((data) => {
	groups.value = data;
	for (const group of data) {
		for (const ch of group.channels) {
			if (ch.sources.length > 0) {
				watchTV(ch, 0);
				return;
			}
		}
	}
});

const calcProgProps = (prog: Programme, index: number, now: Date) => {
	const ts = prog.start.toLocaleTimeString() + ' - ' + prog.end.toLocaleTimeString();

	const props = {
		key: index.toString(),
		color: '#32CD32',
		timestamp: ts,
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

</script>

<template>
	<el-row :gutter="1">
		<el-col :span="18">
			<el-container ref="videoContainerRef">
				<el-main style="padding: 0">
					<MpegTsPlayer :src="source.toLowerCase().startsWith('http') ? source : `/iptv/relay/${source}`" width="100%" />
				</el-main>
				<el-footer>
					<el-menu mode="horizontal" :ellipsis="false">
						<el-sub-menu index="choose-channel">
							<template #title>
								<el-text type="primary">频道：</el-text>
								<el-text>{{chName}}&nbsp;&nbsp;</el-text>
								<el-text type="primary"> 节目源：</el-text>
								<el-text>{{source}}</el-text>
							</template>
							<el-sub-menu v-for="(group, i) in groups" :index="i.toString()">
								<template #title>{{group.name}}</template>
								<template v-for="(ch, j) in group.channels">
									<template v-if="ch.hide || ch.sources.length === 0" />
									<el-menu-item v-else-if="ch.sources.length === 1" @click="watchTV(ch, 0)" :index="`${i}-${j}`">
										{{ch.displayName || ch.name}}
									</el-menu-item>
									<template v-else>
										<el-sub-menu :index="`${i}-${j}`">
											<template #title>{{ch.displayName || ch.name}}</template>
											<el-menu-item v-for="(src, k) in ch.sources" @click="watchTV(ch, k)" :index="`${i}-${j}-${k}`">
												{{'源 ' + src}}
											</el-menu-item>
										</el-sub-menu>
									</template>
								</template>
							</el-sub-menu>
						</el-sub-menu>
					</el-menu>
				</el-footer>
			</el-container>
		</el-col>

		<el-col :span="6">
			<el-scrollbar :height="epgHeight">
				<el-timeline>
					<el-timeline-item v-for="(prog, index) in progs" v-bind="calcProgProps(prog, index, now)">
						<p>{{prog.title}}</p>
					</el-timeline-item>
				</el-timeline>
			</el-scrollbar>
		</el-col>
	</el-row>
</template>
