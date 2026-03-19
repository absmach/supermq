import type { ClientInfo } from "@/lib/api";

export async function getClients(): Promise<ClientInfo[]> {
	const res = await fetch("/api/clients", { cache: "no-store" });
	if (!res.ok) throw new Error(`Failed to fetch clients: ${res.status}`);
	return res.json();
}
