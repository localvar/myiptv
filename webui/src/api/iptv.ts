import axios from 'axios';

export type Channel = {
	name: string;
	displayName: string;
	logo: string;
	hide: boolean;
	sources: string[];
}

export type ChannelGroup = {
	name: string;
	channels: Channel[];
}

export const listChannelGroups = () => {
	return axios.get<ChannelGroup[]>('/api/channel-groups').then(res => res.data);
}

export type RelayClient = {
	addr: string;
	createdAt: string;
}

export type RelayConnection = {
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

export type Programme = {
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