<template>
	<a-table :columns="columns" :dataSource="conns" size="middle" bordered>
		<template #bodyCell="{ text, record, column }">
			<template v-if="column.key.endsWith('reatedAt')">
				{{ dayjs(text).format('YYYY-MM-DD HH:mm:ss') }}
			</template>
			<template v-else-if="column.key === 'action'">
				<a-button type="link" @click="dropConnection(record.addr)">断开</a-button>
			</template>
			<template v-else-if="column.key === 'clientAction'">
				<a-button type="link" @click="dropClient(record.addr, record.clientAddr)">断开</a-button>
			</template>
			<template v-else>
				{{ text }}
			</template>
		</template>
	</a-table>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import dayjs from 'dayjs';
import { listRelayConnections, dropRelayConnection, dropRelayClient } from '../api/iptv';

type Connection = {
	addr: string;
	span: number;
	createdAt: string;
	clientAddr: string;
	clientCreatedAt: string;
};

const conns = ref<Connection[]>([]);

const columns = ref([
	{
		title: '组播地址',
		dataIndex: 'addr',
		key: 'addr',
		align: 'center',
		customCell: (_: any, index: number) => ({
			rowSpan: conns.value[index].span,
		}),
	},
	{
		title: '创建时间',
		dataIndex: 'createdAt',
		key: 'createdAt',
		align: 'center',
		customCell: (_: any, index: number) => ({
			rowSpan: conns.value[index].span,
		}),
	},
	{
		title: '客户端',
		key: 'clients',
		align: 'center',
		children: [
			{
				title: '地址',
				key: 'clientAddr',
				dataIndex: 'clientAddr',
				align: 'center',
			},
			{
				title: '创建时间',
				key: 'clientCreatedAt',
				dataIndex: 'clientCreatedAt',
				align: 'center',
			},
			{
				title: '操作',
				key: 'clientAction',
				align: 'center',
			},
		],
	},
	{
		title: '操作',
		key: 'action',
		align: 'center',
		customCell: (_: any, index: number) => ({
			rowSpan: conns.value[index].span,
		}),
	},
]);

const refresh = () => {
	listRelayConnections().then((data) => {
		const list: Connection[] = [];
		for (const conn of data) {
			let span = conn.clients.length;
			for (const client of conn.clients) {
				list.push({
					addr: conn.addr,
					span: span,
					createdAt: conn.createdAt,
					clientAddr: client.addr,
					clientCreatedAt: client.createdAt,
				} as Connection);
				span = 0;
			}
		}
		conns.value = list;
	});
}

const dropConnection= (addr: string) => {
	dropRelayConnection(addr).then(refresh);
}

const dropClient = (addr: string, clientAddr: string) => {
	dropRelayClient(addr, clientAddr).then(refresh);
}

refresh();

</script>
