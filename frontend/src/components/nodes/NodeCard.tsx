import type { NodeInfo } from "@/api/types";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
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
    <Card
      className={`cursor-pointer transition-all hover:shadow-md hover:border-primary/50 ${
        isMaster ? "ring-2 ring-amber-500/50" : ""
      }`}
      onClick={onClick}
    >
      <CardContent className="p-4">
        <div className="flex items-start justify-between mb-3">
          <div className="flex items-center gap-2">
            <div
              className={`w-10 h-10 rounded-lg flex items-center justify-center ${
                isMaster
                  ? "bg-amber-500/10 text-amber-500"
                  : "bg-primary/10 text-primary"
              }`}
            >
              {isMaster ? (
                <Crown className="w-5 h-5" />
              ) : (
                <Server className="w-5 h-5" />
              )}
            </div>
            <div>
              <p className="font-medium text-sm truncate max-w-[120px]">
                {node.id.slice(0, 8)}...
              </p>
              <p className="text-xs text-muted-foreground">
                {node.address}:{node.port}
              </p>
            </div>
          </div>
          <div className="flex flex-col items-end gap-1">
            <Badge variant={isMaster ? "default" : "secondary"}>
              {isMaster ? "Master" : "Storage"}
            </Badge>
            <StatusBadge status={node.status} />
          </div>
        </div>

        <div className="space-y-3">
          {/* Storage */}
          <div className="space-y-1">
            <div className="flex items-center justify-between text-xs">
              <span className="text-muted-foreground flex items-center gap-1">
                <HardDrive className="w-3 h-3" />
                Przestrzeń
              </span>
              <span>
                {formatBytes(node.storage_used)} / {formatBytes(node.storage_total)}
              </span>
            </div>
            <div className="h-1.5 bg-muted rounded-full overflow-hidden">
              <div
                className={`h-full transition-all ${
                  storagePercent > 90
                    ? "bg-destructive"
                    : storagePercent > 70
                      ? "bg-amber-500"
                      : "bg-primary"
                }`}
                style={{ width: `${storagePercent}%` }}
              />
            </div>
          </div>

          {/* Stats */}
          <div className="flex items-center justify-between text-xs">
            <span className="text-muted-foreground flex items-center gap-1">
              <Activity className="w-3 h-3" />
              Pliki
            </span>
            <span className="font-medium">{node.files_count}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

function StatusBadge({ status }: { status: NodeInfo["status"] }) {
  const config = {
    online: { label: "Online", className: "bg-green-500/10 text-green-600" },
    offline: { label: "Offline", className: "bg-red-500/10 text-red-600" },
    unknown: { label: "Nieznany", className: "bg-gray-500/10 text-gray-600" },
  };

  const { label, className } = config[status];

  return (
    <span
      className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium ${className}`}
    >
      <span
        className={`w-1.5 h-1.5 rounded-full ${
          status === "online"
            ? "bg-green-500"
            : status === "offline"
              ? "bg-red-500"
              : "bg-gray-500"
        }`}
      />
      {label}
    </span>
  );
}

