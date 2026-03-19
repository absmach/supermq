import type { SessionInfo } from "@/lib/api";

export async function getSessions(): Promise<SessionInfo[]> {
	const res = await fetch("/api/sessions", { cache: "no-store" });
	if (!res.ok) throw new Error(`Failed to fetch sessions: ${res.status}`);
	return res.json();
}

export async function getSession(id: string): Promise<SessionInfo> {
	const res = await fetch(`/api/sessions/${encodeURIComponent(id)}`, {
		cache: "no-store",
	});
	if (!res.ok)
		throw new Error(`Failed to fetch session "${id}": ${res.status}`);
	return res.json();
}
