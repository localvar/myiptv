<template>
	<a-list ref="listRef" :data-source="sources" bordered size="small">
		<template #renderItem="{ item, index }">
			<a-list-item :key="item">
				<a-list-item-meta>
					<template #avatar>
						<MenuOutlined style="cursor: grab" />
					</template>
					<template #description>
						<a-typography-text v-model:content="clonedSources[index]" :editable="{onChange: onSourceChange, onEnd: () => onSourceUpdated(index)}"/>
					</template>
				</a-list-item-meta>
				<template #actions>
					<a-space-compact size="small">
						<a-tooltip title="测试">
							<a-button type="link" :icon="h(CaretRightOutlined)" @click="onVerifySource(sources[index])" />
						</a-tooltip>
						<a-tooltip title="删除">
							<a-button type="link" :icon="h(DeleteOutlined)" danger @click="onDeleteSource(index)" />
						</a-tooltip>
					</a-space-compact>
				</template>
			</a-list-item>
		</template>
		<template #footer>
			<a-flex>
				<a-input v-model:value="newSource" style="width: 422px"/>
				<a-space-compact size="small">
					<a-tooltip title="测试">
						<a-button type="link" :icon="h(CaretRightOutlined)" @click="onVerifySource(newSource)" />
					</a-tooltip>
					<a-tooltip title="添加">
						<a-button type="link" :icon="h(PlusOutlined)" @click="onAddSource" />
					</a-tooltip>
				</a-space-compact>
			</a-flex>
		</template>
	</a-list>
</template>

<script setup lang="ts">
import { ref, h, defineModel, computed, watch, nextTick, onMounted } from 'vue';
import { CaretRightOutlined, DeleteOutlined, MenuOutlined, PlusOutlined } from '@ant-design/icons-vue';
import { App } from 'ant-design-vue'
import Sortable from 'sortablejs';

const { message } = App.useApp();
const emit = defineEmits<{ (e: 'verify', source: string): void; }>();

const listRef = ref();
const sources = defineModel<string[]>('sources', { required: true });
const newSource = ref<string>('');
const sortable = ref<Sortable | null>();

const onVerifySource = (source: string) => {
	emit('verify', source);
};

const onDeleteSource = (index: number) => {
	sources.value.splice(index, 1);
};

// clonedSources is used to enable source edit, because the key of the list item
// is also the value of a source.
const clonedSources = computed(() => [...sources.value]);
const editedSource = ref<string>('');

const onSourceChange = (value: any) => {
	editedSource.value = value;
};

const onSourceUpdated = (index: number) => {
	if (editedSource.value === sources.value[index]) {
		return;
	}
	if (sources.value.includes(editedSource.value)) {
		message.error('节目源已存在');
		return;
	}
	sources.value[index] = editedSource.value;
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
	newSource.value = '';
};

const createSortable = (srcs: string[]) => {
	if (srcs.length === 0) {
		sortable.value?.destroy();
		sortable.value = null;
		return;
	}

	nextTick(() => {
		if (sortable.value) {
			return;
		}
		const el = listRef.value.$el.getElementsByClassName('ant-list-items')[0];
		if (!el) {
			return;
		}
		sortable.value = new Sortable( el, {
			handle: '.ant-list-item-meta-avatar',
			onEnd: (evt: Sortable.SortableEvent) => {
				if (evt.newIndex !== evt.oldIndex) {
					const item = sources.value.splice(evt.oldIndex!, 1)[0];
					sources.value.splice(evt.newIndex!, 0, item);
				}
			},
		});
	});
};

onMounted(() => createSortable(sources.value));
watch(() => sources.value, createSortable, { deep: true });
</script>