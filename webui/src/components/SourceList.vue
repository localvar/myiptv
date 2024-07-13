<template>
	<a-list ref="listRef" :data-source="sources" header="节目源" bordered size="small">
		<template #renderItem="{ item, index }">
			<a-list-item :key="item">
				<a-list-item-meta>
					<template #avatar>
						<MenuOutlined style="cursor: grab" />
					</template>
					<template #description>
						<a-typography-text v-model:content="sources[index]" editable />
					</template>
				</a-list-item-meta>

				<template #actions>
					<a-space-compact size="small">
						<a-button type="link" @click="onVerifySource(sources[index])">测试</a-button>
						<a-button type="link" danger @click="onDeleteSource(index)">删除</a-button>
					</a-space-compact>
				</template>
			</a-list-item>
		</template>
		<template v-if="newSource !== undefined" #footer>
			<a-flex>
				<a-input v-model:value="newSource" style="width: 490px"/>
				<a-space-compact size="small">
					<a-button type="link" @click="onVerifySource(newSource)">测试</a-button>
					<a-button type="link" @click="onAddSource">添加</a-button>
				</a-space-compact>
			</a-flex>
		</template>
	</a-list>
</template>

<script setup lang="ts">
import { ref, defineModel, onMounted } from 'vue';
import { MenuOutlined } from '@ant-design/icons-vue';
import { App } from 'ant-design-vue'
import Sortable from 'sortablejs';

const listRef = ref();
const emit = defineEmits<{ (e: 'verify', source: string): void; }>();
const sources = defineModel<string[]>('sources', { required: true });
const newSource = defineModel<string>('newSource');
const { message } = App.useApp();

const onVerifySource = (source: string) => {
	emit('verify', source);
};

const onDeleteSource = (index: number) => {
	sources.value.splice(index, 1);
};

const onAddSource = () => {
	const source = newSource.value!.trim();
	if (!source) {
		return;
	}
	if (sources.value.includes(source) ) {
		message.error('节目源已存在');
		return;
	}
	sources.value.push(source);
};

onMounted(()=>{
	const el = listRef.value.$el.getElementsByClassName('ant-list-items')[0];
	new Sortable( el, {
		handle: '.ant-list-item-meta-avatar',
		onEnd: (evt: Sortable.SortableEvent) => {
			if (evt.newIndex !== evt.oldIndex) {
				const item = sources.value.splice(evt.oldIndex!, 1)[0];
				sources.value.splice(evt.newIndex!, 0, item);
			}
		},
	});
});

</script>