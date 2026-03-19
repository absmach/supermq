import { NextResponse } from "next/server";
import { MOCK_STATUS } from "@/lib/api";

const API_URL = process.env.FLUXMQ_API_URL || "";

export async function GET() {
	if (API_URL) {
		// -- Real backend --
		try {
			const res = await fetch(`${API_URL}/cluster/status`, {
				cache: "no-store",
			});
			if (!res.ok) {
				return NextResponse.json(
					{ error: `Backend returned ${res.status}` },
					{ status: res.status },
				);
			}
			return NextResponse.json(await res.json());
		} catch (err) {
			console.error("Failed to fetch broker status:", err);
			return NextResponse.json(
				{ error: "Could not reach FluxMQ broker" },
				{ status: 503 },
			);
		}
	}

	// -- Mock fallback (no FLUXMQ_API_URL set) --
	return NextResponse.json(MOCK_STATUS);
}
