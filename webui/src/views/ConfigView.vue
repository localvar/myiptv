<template>
	<a-form :labelCol="{span: 4, offset: 5}" :wrapperCol="{span: 8}" :model="config" @finish="onSave">
		<a-form-item label="HTTP 服务地址：">
			<a-space-compact block>
				<a-select v-model:value="selectedIP">
					<template v-for="[iface, ips] in Object.entries(ifaceAndIPs)">
						<a-select-option v-for="ip in ips" :key="ip" :value="ip">
							{{ `${ip} (${iface})` }}
						</a-select-option>
					</template>
				</a-select>
				<a-input-number :min="1" :max="65535" v-model:value="selectedPort" />
			</a-space-compact>
		</a-form-item>
		<a-form-item label="电子节目单地址：">
			<a-input v-model:value="config.epgURL" />
		</a-form-item>
		<a-form-item label="组播网卡：">
			<a-select v-model:value="config.mcastIface">
				<a-select-option v-for="[iface] in Object.entries(ifaceAndIPs)" :key="iface" :value="iface">
					{{ iface }}
				</a-select-option>
			</a-select>
		</a-form-item>
		<a-form-item label="组播包大小：">
			<a-input-number v-model:value="config.mcastPacketSize" addon-after="字节"/>
		</a-form-item>
		<a-form-item label="缓冲区大小：">
			<a-input-number v-model:value="config.writeBufferSize" addon-after="字节"/>
		</a-form-item>
		<a-form-item label="数据接收超时：">
			<a-input-number v-model:value="config.readTimeout" addon-after="毫秒" />
		</a-form-item>

		<a-form-item :wrapperCol="{offset: 10, span: 8}">
			<a-space>
				<a-button type="primary" html-type="submit">保存</a-button>
				<a-button @click="onReset">重置</a-button>
			</a-space>
			<a-popconfirm title="确定要重启服务吗？如果您修改了 HTTP 服务地址，浏览器无法自动重新加载页面，需要您手动输入新的地址。" @confirm="onRestart">
				<a-button type="link" danger>重启服务</a-button>
			</a-popconfirm>
		</a-form-item>
	</a-form>	
</template>

<script setup lang="ts">
import {ref} from 'vue';
import {App} from 'ant-design-vue';
import { Config, getConfig, updateConfig, listInterfaceAndIPs, restart } from '../api/iptv';

const {message} = App.useApp();

const config = ref<Config>({} as Config);
const ifaceAndIPs = ref({});
const selectedIP = ref('');
const selectedPort = ref(7709);

listInterfaceAndIPs().then((o) => {
	ifaceAndIPs.value = o;
});

const onSave = () => {
	config.value.serverAddr = `${selectedIP.value}:${selectedPort.value}`;
	updateConfig(config.value).then(() => {
		message.success('保存成功');
	}).catch((e) => {
		message.error(`保存失败: ${e}`);
	})
};

const onReset = () => {
	getConfig().then((c) => {
		config.value = c
		if (c.serverAddr) {
			const [ip, port] = c.serverAddr.split(':');
			selectedIP.value = ip;
			selectedPort.value = parseInt(port);
		}
	});
}

const onRestart = () => {
	message.info('服务重启中...', 5, () => {
		window.location.pathname = '/';
	});
	restart();
}

onReset();
</script>
