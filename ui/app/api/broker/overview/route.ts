import { NextResponse } from "next/server";
import {
	type BrokerStatus,
	MOCK_NODES,
	MOCK_STATUS,
	type NodeInfo,
} from "@/lib/api";

const API_URL = process.env.FLUXMQ_API_URL || "";
console.log("Using API URL:", API_URL);

export interface OverviewResponse {
	status: BrokerStatus;
	nodes: NodeInfo[];
}

export async function GET() {
	if (API_URL) {
		try {
			const res = await fetch(`${API_URL}/api/v1/overview`, {
				cache: "no-store",
			});
			if (!res.ok) {
				return NextResponse.json(
					{ error: `Backend returned ${res.status}` },
					{ status: res.status },
				);
			}
			const data = await res.json();

			const status: BrokerStatus = {
				node_id: data.node_id,
				is_leader: data.is_leader,
				cluster_mode: data.cluster_mode,
				node_count: data.cluster?.nodes?.length ?? 0,
				sessions:
					data.sessions?.connected ?? data.stats?.connections?.current ?? 0,
				messages_received: data.stats?.messages?.received ?? 0,
				messages_sent: data.stats?.messages?.sent ?? 0,
				bytes_received: data.stats?.bytes?.received ?? 0,
				bytes_sent: data.stats?.bytes?.sent ?? 0,
				subscriptions:
					data.stats?.by_protocol?.mqtt?.subscriptions?.active ?? 0,
				retained_messages:
					data.stats?.by_protocol?.mqtt?.subscriptions?.retained_messages ?? 0,
				uptime_seconds: data.uptime_seconds ?? 0,
				auth_errors: data.stats?.by_protocol?.mqtt?.errors?.auth ?? 0,
			};

			const nodes: NodeInfo[] = (data.cluster?.nodes ?? []).map(
				(n: {
					id: string;
					address: string;
					leader: boolean;
					uptime_seconds: number;
				}) => ({
					node_id: n.id,
					is_leader: n.leader,
					addr: n.address,
					uptime_seconds: n.uptime_seconds,
				}),
			);

			return NextResponse.json({ status, nodes } satisfies OverviewResponse);
		} catch (err) {
			console.error("Failed to fetch broker overview:", err);
			return NextResponse.json(
				{ error: "Could not reach FluxMQ broker" },
				{ status: 503 },
			);
		}
	}

	// Mock fallback
	return NextResponse.json({
		status: MOCK_STATUS,
		nodes: MOCK_NODES,
	} satisfies OverviewResponse);
}
