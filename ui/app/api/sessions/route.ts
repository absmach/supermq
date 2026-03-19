import { NextResponse } from "next/server";
import { MOCK_SESSIONS } from "@/lib/api";

const API_URL = process.env.FLUXMQ_API_URL || "";

export async function GET() {
	if (API_URL) {
		try {
			const res = await fetch(`${API_URL}/api/v1/sessions`, {
				cache: "no-store",
			});
			if (!res.ok) {
				return NextResponse.json(
					{ error: `Backend returned ${res.status}` },
					{ status: res.status },
				);
			}
			const data = await res.json();
			// Backend returns { sessions: [...], next_page_token: "..." }
			return NextResponse.json(data.sessions ?? []);
		} catch (err) {
			console.error("Failed to fetch sessions:", err);
			return NextResponse.json(
				{ error: "Could not reach FluxMQ broker" },
				{ status: 503 },
			);
		}
	}

	return NextResponse.json(MOCK_SESSIONS);
}
