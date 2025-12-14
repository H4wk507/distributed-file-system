import type { NodeInfo } from "@/api/types";
import { formatBytes } from "@/lib/formatters";
import { Activity, Crown, HardDrive, Server } from "lucide-react";

interface NodeCardProps {
  node: NodeInfo;
  onClick?: () => void;
}

export function NodeCard({ node, onClick }: NodeCardProps) {
  const isMaster = node.role === "master";

  const storagePercent =
    node.storage_total > 0
      ? Math.round((node.storage_used / node.storage_total) * 100)
      : 0;

  return (
    <div
      className={`border cursor-pointer transition-colors hover:border-primary bg-card ${
        isMaster ? "border-primary" : "border-border"
      }`}
      onClick={onClick}
    >
      {/* Header */}
      <div className="flex items-start justify-between p-3 border-b border-border">
        <div className="flex items-center gap-2">
          <div
            className={`${isMaster ? "text-primary" : "text-muted-foreground"}`}
          >
            {isMaster ? (
              <Crown className="w-4 h-4" />
            ) : (
              <Server className="w-4 h-4" />
            )}
          </div>
          <div>
            <p className="font-mono text-xs truncate max-w-[100px]">
              {node.id.slice(0, 8)}...
            </p>
            <p className="text-xs text-muted-foreground font-mono">
              {node.ip}:{node.port}
            </p>
          </div>
        </div>
        <div className="flex flex-col items-end gap-1">
          <span
            className={`text-xs font-mono uppercase px-1.5 py-0.5 border ${
              isMaster
                ? "border-primary text-primary bg-primary/10"
                : "border-border text-muted-foreground"
            }`}
          >
            {isMaster ? "master" : "storage"}
          </span>
          <StatusBadge status={node.status} />
        </div>
      </div>

      {/* Content */}
      <div className="p-3 space-y-2">
        {/* Storage */}
        <div className="space-y-1">
          <div className="flex items-center justify-between text-xs">
            <span className="text-muted-foreground flex items-center gap-1 font-mono">
              <HardDrive className="w-3 h-3" />
              storage
            </span>
            <span className="font-mono">
              {formatBytes(node.storage_used)} /{" "}
              {formatBytes(node.storage_total)}
            </span>
          </div>
          <div className="h-1 bg-muted border border-border overflow-hidden">
            <div
              className={`h-full transition-all ${
                storagePercent > 90
                  ? "bg-destructive"
                  : storagePercent > 70
                    ? "bg-primary"
                    : "bg-success"
              }`}
              style={{ width: `${storagePercent}%` }}
            />
          </div>
        </div>

        {/* Files count */}
        <div className="flex items-center justify-between text-xs">
          <span className="text-muted-foreground flex items-center gap-1 font-mono">
            <Activity className="w-3 h-3" />
            files
          </span>
          <span className="font-mono text-primary">{node.files_count}</span>
        </div>
      </div>
    </div>
  );
}

function StatusBadge({ status }: { status: NodeInfo["status"] }) {
  const config = {
    online: {
      label: "online",
      className: "border-success text-success bg-success/10",
    },
    offline: {
      label: "offline",
      className: "border-destructive text-destructive bg-destructive/10",
    },
    unknown: {
      label: "unknown",
      className: "border-muted-foreground text-muted-foreground bg-muted",
    },
  };

  const { label, className } = config[status];

  return (
    <span
      className={`inline-flex items-center gap-1 px-1.5 py-0.5 text-xs font-mono uppercase border ${className}`}
    >
      <span
        className={`w-1.5 h-1.5 ${
          status === "online"
            ? "bg-success"
            : status === "offline"
              ? "bg-destructive"
              : "bg-muted-foreground"
        }`}
      />
      {label}
    </span>
  );
}
