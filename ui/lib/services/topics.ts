import type { TopicInfo } from "@/lib/api";

export async function getTopics(): Promise<TopicInfo[]> {
	const res = await fetch("/api/topics", { cache: "no-store" });
	if (!res.ok) throw new Error(`Failed to fetch topics: ${res.status}`);
	return res.json();
}
