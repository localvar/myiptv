import axios from 'axios';

export interface Channel {
	name: string;
	displayName: string;
	logo: string;
	hide: boolean;
	sources: string[];

	selected?: boolean; // this field is only used in the frontend.
}

export interface ChannelGroup {
	name: string;
	channels: Channel[];
}

export const listChannelGroups = () => {
	return axios.get<ChannelGroup[]>('/api/channel-groups')
		.then(res => {
			res.data.forEach(g => {
				if (!g.channels) {
					g.channels = [];
				} else {
					g.channels.forEach(c => !c.sources && (c.sources = []));
				}});
			return res.data;
		});
}

export const updateChannelGroups = (groups: ChannelGroup[]) => {
	groups = groups.map(g => ({
		name: g.name,
		channels: g.channels.map(c => ({
			name: c.name,
			displayName: c.displayName,
			logo: c.logo,
			hide: c.hide,
			sources: c.sources
		}))
	}));

	return axios.put('/api/channel-groups', groups);
}

export interface RelayClient {
	addr: string;
	createdAt: string;
}

export interface RelayConnection {
	addr: string;
	createdAt: string;
	clients: RelayClient[];
}

export const listRelayConnections = () => {
	return axios.get<RelayConnection[]>('/api/relays').then(res => res.data);
}

export const dropRelayConnection = (addr: string) => {
	return axios.delete(`/api/relays/${addr}`);
}

export const dropRelayClient = (addr: string, clientAddr: string) => {
	return axios.delete(`/api/relays/${addr}/${clientAddr}`);
}

export interface Programme {
	title: string;
	start: Date;
	end: Date;
	desc: string;
}

export const getEpg = (channel: string) => {
	return axios.get<any[]>(`/api/epg/${channel}`)
		.then(res => {
			return res.data.map(p => {
				return {
					title: p.title,
					start: new Date(p.start),
					end: new Date(p.end),
					desc: p.desc
				} as Programme
			})
		})
		.catch(() => {
			return [];
		});
}

export interface Config {
	serverAddr: string;
	epgURL: string;
	mcastIface: string;
	mcastPacketSize: number;
	writeBufferSize: number;
	readTimeout: number;
}

export const getConfig = () => {
	return axios.get<Config>('/api/config').then(res => res.data);
}

export const listInterfaceAndIPs = () => {
	return axios.get<Map<string, string[]>>('/api/interfaces-and-ips').then(res => res.data);
}

export const updateConfig = (config: Config) => {
	return axios.put('/api/config', config);
}

export const restart = () => {
	return axios.post('/api/restart');
}
