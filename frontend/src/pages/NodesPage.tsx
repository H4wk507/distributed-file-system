import type { NodeInfo } from "@/api/types";
import { MainLayout } from "@/components/MainLayout";
import { NodeCard } from "@/components/nodes/NodeCard";
import { NodeDetail } from "@/components/nodes/NodeDetail";
import { Card, CardContent } from "@/components/ui/card";
import { Activity, Crown, HardDrive, Server } from "lucide-react";
import { useState } from "react";

// TODO: remove - mock data
const mockNodes: NodeInfo[] = [
  {
    id: "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    address: "192.168.1.10",
    port: 8081,
    role: "master",
    status: "online",
    storage_used: 0,
    storage_total: 0,
    files_count: 0,
    last_heartbeat: new Date().toISOString(),
  },
  {
    id: "b2c3d4e5-f6a7-8901-bcde-f12345678901",
    address: "192.168.1.11",
    port: 8082,
    role: "storage",
    status: "online",
    storage_used: 52428800,
    storage_total: 107374182400,
    files_count: 156,
    last_heartbeat: new Date().toISOString(),
  },
  {
    id: "c3d4e5f6-a7b8-9012-cdef-123456789012",
    address: "192.168.1.12",
    port: 8083,
    role: "storage",
    status: "online",
    storage_used: 31457280000,
    storage_total: 107374182400,
    files_count: 234,
    last_heartbeat: new Date().toISOString(),
  },
  {
    id: "d4e5f6a7-b8c9-0123-def0-234567890123",
    address: "192.168.1.13",
    port: 8084,
    role: "storage",
    status: "offline",
    storage_used: 85899345920,
    storage_total: 107374182400,
    files_count: 89,
    last_heartbeat: new Date(Date.now() - 300000).toISOString(),
  },
];

export default function NodesPage() {
  const [selectedNode, setSelectedNode] = useState<NodeInfo | null>(null);

  const nodes = mockNodes;
  const masterNode = nodes.find((n) => n.role === "master");
  const storageNodes = nodes.filter((n) => n.role === "storage");
  const onlineNodes = nodes.filter((n) => n.status === "online");
  const totalStorage = storageNodes.reduce(
    (acc, n) => acc + n.storage_total,
    0,
  );
  const usedStorage = storageNodes.reduce((acc, n) => acc + n.storage_used, 0);

  return (
    <MainLayout>
      <div className="mb-8">
        <h1 className="text-2xl font-semibold">Węzły</h1>
        <p className="text-muted-foreground mt-1">
          Zarządzaj węzłami w klastrze rozproszonego systemu plików
        </p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <StatCard
          icon={<Server className="w-5 h-5" />}
          label="Wszystkie węzły"
          value={nodes.length.toString()}
        />
        <StatCard
          icon={<Activity className="w-5 h-5" />}
          label="Online"
          value={onlineNodes.length.toString()}
          highlight={onlineNodes.length === nodes.length}
        />
        <StatCard
          icon={<Crown className="w-5 h-5" />}
          label="Master"
          value={masterNode?.address || "—"}
        />
        <StatCard
          icon={<HardDrive className="w-5 h-5" />}
          label="Przestrzeń"
          value={`${Math.round((usedStorage / totalStorage) * 100) || 0}%`}
        />
      </div>

      {/* Master Node */}
      {masterNode && (
        <div className="mb-6">
          <h2 className="text-sm font-medium text-muted-foreground mb-3">
            Węzeł Master
          </h2>
          <div className="max-w-sm">
            <NodeCard
              node={masterNode}
              onClick={() => setSelectedNode(masterNode)}
            />
          </div>
        </div>
      )}

      {/* Storage Nodes */}
      <div>
        <h2 className="text-sm font-medium text-muted-foreground mb-3">
          Węzły Storage ({storageNodes.length})
        </h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {storageNodes.map((node) => (
            <NodeCard
              key={node.id}
              node={node}
              onClick={() => setSelectedNode(node)}
            />
          ))}
        </div>
      </div>

      {/* Node Detail Panel */}
      <NodeDetail
        node={selectedNode}
        open={!!selectedNode}
        onOpenChange={(open) => !open && setSelectedNode(null)}
      />
    </MainLayout>
  );
}

function StatCard({
  icon,
  label,
  value,
  highlight,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
  highlight?: boolean;
}) {
  return (
    <Card>
      <CardContent className="p-4">
        <div className="flex items-center gap-3">
          <div
            className={`${highlight ? "text-green-500" : "text-muted-foreground"}`}
          >
            {icon}
          </div>
          <div>
            <p className="text-xs text-muted-foreground">{label}</p>
            <p className="text-lg font-semibold">{value}</p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
