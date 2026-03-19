"use client";

import { Edit, X } from "lucide-react";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";

const configItems = [
	{ key: "server.tcp.plain.addr", value: ":1883", category: "Server" },
	{ key: "server.health_addr", value: ":8081", category: "Server" },
	{ key: "server.shutdown_timeout", value: "30s", category: "Server" },
	{ key: "broker.max_message_size", value: "1048576", category: "Broker" },
	{ key: "broker.max_retained_messages", value: "10000", category: "Broker" },
	{ key: "session.max_sessions", value: "10000", category: "Session" },
	{ key: "session.max_offline_queue_size", value: "1000", category: "Session" },
	{ key: "storage.type", value: "badger", category: "Storage" },
	{ key: "storage.badger_dir", value: "/tmp/fluxmq/data", category: "Storage" },
	{ key: "cluster.enabled", value: "false", category: "Cluster" },
	{ key: "cluster.node_id", value: "broker-1", category: "Cluster" },
	{ key: "log.level", value: "debug", category: "Logging" },
	{ key: "log.format", value: "text", category: "Logging" },
];

const ConfigClient = () => {
	const [isEditing, setIsEditing] = useState(false);

	const grouped = configItems.reduce(
		(acc: Record<string, typeof configItems>, item) => {
			(acc[item.category] ??= []).push(item);
			return acc;
		},
		{},
	);

	return (
		<div className="p-8 space-y-6">
			<div className="flex items-center justify-between">
				<div>
					<h1 className="text-3xl font-bold text-flux-text mb-1">
						Configuration
					</h1>
					<p className="text-flux-text-muted">
						Broker settings from running config
					</p>
				</div>
				<Button
					variant="outline"
					size="sm"
					onClick={() => setIsEditing(!isEditing)}
					className={
						isEditing
							? "bg-flux-red/10 text-flux-red border-flux-red/30 hover:bg-flux-red/20 hover:text-flux-red"
							: "bg-flux-blue/10 text-flux-blue border-flux-blue/30 hover:bg-flux-blue/20 hover:text-flux-blue"
					}
				>
					{isEditing ? (
						<>
							<X size={16} />
							Cancel
						</>
					) : (
						<>
							<Edit size={16} />
							Edit
						</>
					)}
				</Button>
			</div>

			<div className="space-y-4">
				{Object.entries(grouped).map(([category, items]) => (
					<Card
						key={category}
						className="border-flux-card-border bg-flux-card overflow-hidden"
					>
						<div className="px-6 py-3 border-b border-flux-card-border bg-flux-hover">
							<h2 className="text-sm font-semibold text-flux-text-muted uppercase tracking-wide">
								{category}
							</h2>
						</div>
						<CardContent className="p-0">
							<div className="divide-y divide-flux-card-border">
								{items.map((item, i) => (
									<div
										key={i}
										className="px-6 py-3 flex items-center justify-between"
									>
										<p className="text-flux-text text-sm font-mono">
											{item.key}
										</p>
										{isEditing ? (
											<Input
												type="text"
												defaultValue={item.value}
												className="w-56 bg-flux-bg border-flux-card-border text-flux-text text-sm focus-visible:ring-flux-blue"
											/>
										) : (
											<span className="text-flux-text-muted text-sm font-mono">
												{item.value}
											</span>
										)}
									</div>
								))}
							</div>
						</CardContent>
					</Card>
				))}
			</div>

			{isEditing && (
				<div className="flex gap-3 justify-end">
					<Button
						variant="outline"
						size="sm"
						onClick={() => setIsEditing(false)}
						className="border-flux-card-border text-flux-text-muted hover:text-flux-text hover:bg-flux-hover"
					>
						Cancel
					</Button>
					<Button
						size="sm"
						className="bg-flux-green hover:bg-green-600 text-white border-0"
					>
						Save Changes
					</Button>
				</div>
			)}
		</div>
	);
};

export default ConfigClient;
