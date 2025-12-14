import type { NodeInfo } from "@/api/types";
import { MainLayout } from "@/components/MainLayout";
import { NodeCard } from "@/components/nodes/NodeCard";
import { NodeDetail } from "@/components/nodes/NodeDetail";
import { useNodes } from "@/hooks/useNodes";
import { Activity, Crown, HardDrive, Server } from "lucide-react";
import { useState } from "react";

export default function NodesPage() {
  const [selectedNode, setSelectedNode] = useState<NodeInfo | null>(null);

  const { data: nodes = [] } = useNodes();
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
      {/* Nagłówek */}
      <div className="mb-6 pb-4 border-b border-border">
        <h1 className="text-sm font-mono uppercase tracking-wide text-primary">
          :: węzły ::
        </h1>
        <p className="text-xs text-muted-foreground font-mono mt-1">
          zarządzaj węzłami w klastrze
        </p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-4 gap-px bg-border mb-6">
        <StatBlock
          icon={<Server className="w-4 h-4" />}
          label="wszystkie"
          value={nodes.length.toString()}
        />
        <StatBlock
          icon={<Activity className="w-4 h-4" />}
          label="online"
          value={onlineNodes.length.toString()}
          highlight={onlineNodes.length === nodes.length}
        />
        <StatBlock
          icon={<Crown className="w-4 h-4" />}
          label="master"
          value={masterNode?.ip || "—"}
        />
        <StatBlock
          icon={<HardDrive className="w-4 h-4" />}
          label="przestrzeń"
          value={`${Math.round((usedStorage / totalStorage) * 100) || 0}%`}
        />
      </div>

      {/* Master Node */}
      {masterNode && (
        <section className="mb-6">
          <div className="flex items-center gap-2 mb-3">
            <span className="text-xs font-mono uppercase tracking-wide text-muted-foreground">
              [master]
            </span>
            <div className="flex-1 h-px bg-border" />
          </div>
          <div className="max-w-sm">
            <NodeCard
              node={masterNode}
              onClick={() => setSelectedNode(masterNode)}
            />
          </div>
        </section>
      )}

      {/* Storage Nodes */}
      <section>
        <div className="flex items-center gap-2 mb-3">
          <span className="text-xs font-mono uppercase tracking-wide text-muted-foreground">
            [storage] ({storageNodes.length})
          </span>
          <div className="flex-1 h-px bg-border" />
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {storageNodes.map((node) => (
            <NodeCard
              key={node.id}
              node={node}
              onClick={() => setSelectedNode(node)}
            />
          ))}
        </div>
      </section>

      {/* Node Detail Panel */}
      <NodeDetail
        node={selectedNode}
        open={!!selectedNode}
        onOpenChange={(open) => !open && setSelectedNode(null)}
      />
    </MainLayout>
  );
}

function StatBlock({
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
    <div className="bg-card p-3">
      <div
        className={`flex items-center gap-2 mb-1 ${highlight ? "text-success" : "text-muted-foreground"}`}
      >
        {icon}
        <span className="text-xs font-mono uppercase">{label}</span>
      </div>
      <p
        className={`text-lg font-mono ${highlight ? "text-success" : "text-primary"}`}
      >
        {value}
      </p>
    </div>
  );
}
