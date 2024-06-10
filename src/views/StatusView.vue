<script setup lang="ts">
import { ref } from 'vue';
import {RelayConnection, listRelayConnections, dropRelayConnection, dropRelayClient} from '../api/iptv';

const conns = ref<RelayConnection[]>([]);

listRelayConnections().then((data) => {
	conns.value = data;
});

const dropConnection= (addr: string) => {
	dropRelayConnection(addr).then(() => {
		conns.value = conns.value.filter((conn) => conn.addr !== addr);
	});
}

const dropClient = (addr: string, clientAddr: string) => {
	dropRelayClient(addr, clientAddr).then(() => {
		conns.value = conns.value.map((conn) => {
			if (conn.addr === addr) {
				conn.clients = conn.clients.filter((client) => client.addr !== clientAddr);
			}
			return conn;
		});
	});
}

</script>

<template>
	<table border="1" width="100%">
		<thead>
			<tr>
				<th rowspan="2">Address</th>
				<th rowspan="2">Created At</th>
				<th colspan="3">Clients</th>
				<th rowspan="2">Actions</th>
			</tr>
			<tr>
				<th>Address</th>
				<th>Created At</th>
				<th>Actions</th>
			</tr>
		</thead>
		<tbody>
			<template v-for="conn in conns">
				<tr>
					<td :rowspan="conn.clients.length">
						{{ conn.addr }}
					</td>
					<td :rowspan="conn.clients.length">
						{{ conn.createdAt }}
					</td>
					<td>
						{{ conn.clients[0].addr}}
					</td>
					<td>
						{{ conn.clients[0].createdAt}}
					</td>
					<td>
						<button @click="dropClient(conn.addr, conn.clients[0].addr)">Delete</button>
					</td>
					<td :rowspan="conn.clients.length">
						<button @click="dropConnection(conn.addr)">Delete</button>
					</td>
				</tr>
				<template v-for="(client,i) in conn.clients">
					<tr v-if="i > 0">
						<td>
							{{ client.addr }}
						</td>
						<td>
							{{ client.createdAt }}
						</td>
						<td>
							<button @click="dropClient(conn.addr, client.addr)">Delete</button>
						</td>
					</tr>
				</template>
			</template>
		</tbody>
	</table>
</template>
