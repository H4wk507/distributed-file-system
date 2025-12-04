import type { NodeInfo } from "@/api/types";
import { Badge } from "@/components/ui/badge";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { formatBytes } from "@/lib/formatters";
import {
  Activity,
  Clock,
  Crown,
  FileText,
  HardDrive,
  Network,
  Server,
} from "lucide-react";

interface NodeDetailProps {
  node: NodeInfo | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function NodeDetail({ node, open, onOpenChange }: NodeDetailProps) {
  if (!node) return null;

  const isMaster = node.role === "master";
  const storagePercent =
    node.storage_total > 0
      ? Math.round((node.storage_used / node.storage_total) * 100)
      : 0;

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="overflow-y-auto">
        <SheetHeader>
          <div className="flex items-center gap-3">
            <div
              className={`w-12 h-12 rounded-lg flex items-center justify-center ${
                isMaster
                  ? "bg-amber-500/10 text-amber-500"
                  : "bg-primary/10 text-primary"
              }`}
            >
              {isMaster ? (
                <Crown className="w-6 h-6" />
              ) : (
                <Server className="w-6 h-6" />
              )}
            </div>
            <div>
              <SheetTitle className="flex items-center gap-2">
                Węzeł {node.id.slice(0, 8)}
                <Badge variant={isMaster ? "default" : "secondary"}>
                  {isMaster ? "Master" : "Storage"}
                </Badge>
              </SheetTitle>
              <SheetDescription>
                {node.address}:{node.port}
              </SheetDescription>
            </div>
          </div>
        </SheetHeader>

        <Tabs defaultValue="overview" className="mt-6">
          <TabsList className="w-full">
            <TabsTrigger value="overview" className="flex-1">
              Przegląd
            </TabsTrigger>
            <TabsTrigger value="files" className="flex-1">
              Pliki
            </TabsTrigger>
            <TabsTrigger value="logs" className="flex-1">
              Logi
            </TabsTrigger>
          </TabsList>

          <TabsContent value="overview" className="space-y-6 mt-4">
            {/* Status */}
            <div className="space-y-3">
              <h4 className="text-sm font-medium flex items-center gap-2">
                <Activity className="w-4 h-4" />
                Status
              </h4>
              <div className="grid grid-cols-2 gap-3">
                <StatItem
                  label="Stan"
                  value={
                    <StatusBadge
                      status={node.status}
                    />
                  }
                />
                <StatItem label="Rola" value={isMaster ? "Master" : "Storage"} />
              </div>
            </div>

            {/* Storage */}
            <div className="space-y-3">
              <h4 className="text-sm font-medium flex items-center gap-2">
                <HardDrive className="w-4 h-4" />
                Przestrzeń dyskowa
              </h4>
              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Wykorzystanie</span>
                  <span>
                    {formatBytes(node.storage_used)} /{" "}
                    {formatBytes(node.storage_total)}
                  </span>
                </div>
                <div className="h-2 bg-muted rounded-full overflow-hidden">
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
                <p className="text-xs text-muted-foreground text-right">
                  {storagePercent}% wykorzystane
                </p>
              </div>
            </div>

            {/* Network */}
            <div className="space-y-3">
              <h4 className="text-sm font-medium flex items-center gap-2">
                <Network className="w-4 h-4" />
                Sieć
              </h4>
              <div className="grid grid-cols-2 gap-3">
                <StatItem label="Adres IP" value={node.address} />
                <StatItem label="Port" value={node.port.toString()} />
              </div>
            </div>

            {/* Info */}
            <div className="space-y-3">
              <h4 className="text-sm font-medium flex items-center gap-2">
                <Clock className="w-4 h-4" />
                Informacje
              </h4>
              <div className="space-y-2">
                <StatItem label="ID węzła" value={node.id} mono />
                <StatItem
                  label="Ostatni heartbeat"
                  value={new Date(node.last_heartbeat).toLocaleString("pl-PL")}
                />
              </div>
            </div>
          </TabsContent>

          <TabsContent value="files" className="mt-4">
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <h4 className="text-sm font-medium flex items-center gap-2">
                  <FileText className="w-4 h-4" />
                  Pliki na węźle
                </h4>
                <Badge variant="secondary">{node.files_count}</Badge>
              </div>
              <div className="text-center py-8 text-muted-foreground">
                <FileText className="w-8 h-8 mx-auto mb-2 opacity-50" />
                <p className="text-sm">Lista plików niedostępna</p>
                <p className="text-xs">Wymaga implementacji backendu</p>
              </div>
            </div>
          </TabsContent>

          <TabsContent value="logs" className="mt-4">
            <div className="space-y-3">
              <h4 className="text-sm font-medium">Ostatnie logi</h4>
              <div className="text-center py-8 text-muted-foreground">
                <Activity className="w-8 h-8 mx-auto mb-2 opacity-50" />
                <p className="text-sm">Logi niedostępne</p>
                <p className="text-xs">Wymaga implementacji WebSocket</p>
              </div>
            </div>
          </TabsContent>
        </Tabs>
      </SheetContent>
    </Sheet>
  );
}

function StatItem({
  label,
  value,
  mono,
}: {
  label: string;
  value: React.ReactNode;
  mono?: boolean;
}) {
  return (
    <div className="space-y-1">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p
        className={`text-sm ${mono ? "font-mono text-xs break-all" : "font-medium"}`}
      >
        {value}
      </p>
    </div>
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

