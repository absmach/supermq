"use client";

import { Search } from "lucide-react";
import { useEffect, useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "@/components/ui/table";
import { TablePagination } from "@/components/ui/table-pagination";
import type { ClientInfo } from "@/lib/api";
import { getClients } from "@/lib/services/clients";

const ClientsClient = () => {
	const [clients, setClients] = useState<ClientInfo[]>([]);
	const [page, setPage] = useState(1);
	const [limit, setLimit] = useState(10);

	useEffect(() => {
		getClients().then(setClients).catch(console.error);
	}, []);

	const totalPages = Math.max(1, Math.ceil(clients.length / limit));
	const paginated = clients.slice((page - 1) * limit, page * limit);

	return (
		<div className="p-8 space-y-6">
			<div>
				<h1 className="text-3xl font-bold text-flux-text mb-1">Clients</h1>
				<p className="text-flux-text-muted">Manage and monitor MQTT clients</p>
			</div>

			<Card className="border-flux-card-border bg-flux-card">
				<CardContent className="p-6">
					<div className="flex items-center justify-between mb-6">
						<div className="relative flex-1 max-w-xs">
							<Search
								className="absolute left-3 top-1/2 -translate-y-1/2 text-flux-blue dark:text-flux-orange"
								size={18}
							/>
							<Input
								type="text"
								placeholder="Search clients..."
								className="pl-10 bg-flux-bg border-flux-card-border text-flux-text placeholder:text-flux-text-muted focus-visible:ring-flux-blue"
							/>
						</div>
					</div>

					<Table>
						<TableHeader>
							<TableRow className="border-flux-card-border hover:bg-transparent">
								<TableHead>Client ID</TableHead>
								<TableHead>Status</TableHead>
								<TableHead>Protocol</TableHead>
								<TableHead>Subscriptions</TableHead>
								<TableHead>Connected At</TableHead>
							</TableRow>
						</TableHeader>
						<TableBody>
							{paginated.map((client, i) => (
								<TableRow
									key={i}
									className="border-flux-card-border hover:bg-flux-hover"
								>
									<TableCell className="text-flux-text font-medium text-sm">
										{client.client_id}
									</TableCell>
									<TableCell>
										<Badge
											variant="outline"
											className={`inline-flex items-center gap-1.5 ${
												client.connected
													? "bg-flux-green/10 text-flux-green border-flux-green/20"
													: "bg-flux-red/10 text-flux-red border-flux-red/20"
											}`}
										>
											<span
												className={`w-1.5 h-1.5 rounded-full ${client.connected ? "bg-flux-green" : "bg-flux-red"}`}
											/>
											{client.connected ? "Connected" : "Disconnected"}
										</Badge>
									</TableCell>
									<TableCell>
										<Badge
											variant="outline"
											className="bg-flux-blue/10 text-flux-blue border-flux-blue/20"
										>
											{client.protocol ?? "MQTT"}
										</Badge>
									</TableCell>
									<TableCell className="text-flux-text text-sm">
										{client.subscriptions}
									</TableCell>
									<TableCell className="text-flux-text-muted text-sm">
										{new Date(client.connected_at).toLocaleString()}
									</TableCell>
								</TableRow>
							))}
						</TableBody>
					</Table>
					<TablePagination
						page={page}
						limit={limit}
						totalPages={totalPages}
						totalItems={clients.length}
						setPage={setPage}
						setLimit={setLimit}
						itemLabel="clients"
					/>
				</CardContent>
			</Card>
		</div>
	);
};

export default ClientsClient;
