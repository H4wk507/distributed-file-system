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
              className={`${isMaster ? "text-primary" : "text-muted-foreground"}`}
            >
              {isMaster ? (
                <Crown className="w-5 h-5" />
              ) : (
                <Server className="w-5 h-5" />
              )}
            </div>
            <div>
              <SheetTitle className="flex items-center gap-2">
                węzeł {node.id.slice(0, 8)}
                <Badge variant={isMaster ? "default" : "secondary"}>
                  {isMaster ? "master" : "storage"}
                </Badge>
              </SheetTitle>
              <SheetDescription>
                {node.ip}:{node.port}
              </SheetDescription>
            </div>
          </div>
        </SheetHeader>

        <Tabs defaultValue="overview" className="mt-6">
          <TabsList className="w-full">
            <TabsTrigger value="overview" className="flex-1">
              przegląd
            </TabsTrigger>
            <TabsTrigger value="files" className="flex-1">
              pliki
            </TabsTrigger>
            <TabsTrigger value="logs" className="flex-1">
              logi
            </TabsTrigger>
          </TabsList>

          <TabsContent value="overview" className="space-y-4 mt-4">
            {/* Status */}
            <section>
              <div className="flex items-center gap-2 mb-2">
                <Activity className="w-3.5 h-3.5 text-muted-foreground" />
                <span className="text-xs font-mono uppercase text-muted-foreground">
                  status
                </span>
              </div>
              <div className="border border-border divide-y divide-border">
                <StatItem
                  label="stan"
                  value={<StatusBadge status={node.status} />}
                />
                <StatItem
                  label="rola"
                  value={isMaster ? "master" : "storage"}
                />
              </div>
            </section>

            {/* Storage */}
            <section>
              <div className="flex items-center gap-2 mb-2">
                <HardDrive className="w-3.5 h-3.5 text-muted-foreground" />
                <span className="text-xs font-mono uppercase text-muted-foreground">
                  przestrzeń
                </span>
              </div>
              <div className="border border-border p-3 space-y-2">
                <div className="flex justify-between text-xs font-mono">
                  <span className="text-muted-foreground">wykorzystanie</span>
                  <span>
                    {formatBytes(node.storage_used)} /{" "}
                    {formatBytes(node.storage_total)}
                  </span>
                </div>
                <div className="h-2 bg-muted border border-border overflow-hidden">
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
                <p className="text-xs text-muted-foreground font-mono text-right">
                  {storagePercent}%
                </p>
              </div>
            </section>

            {/* Network */}
            <section>
              <div className="flex items-center gap-2 mb-2">
                <Network className="w-3.5 h-3.5 text-muted-foreground" />
                <span className="text-xs font-mono uppercase text-muted-foreground">
                  sieć
                </span>
              </div>
              <div className="border border-border divide-y divide-border">
                <StatItem label="ip" value={node.ip} />
                <StatItem label="port" value={node.port.toString()} />
              </div>
            </section>

            {/* Info */}
            <section>
              <div className="flex items-center gap-2 mb-2">
                <Clock className="w-3.5 h-3.5 text-muted-foreground" />
                <span className="text-xs font-mono uppercase text-muted-foreground">
                  info
                </span>
              </div>
              <div className="border border-border divide-y divide-border">
                <StatItem label="id" value={node.id} mono />
                <StatItem
                  label="heartbeat"
                  value={new Date(node.last_heartbeat).toLocaleString("pl-PL")}
                />
              </div>
            </section>
          </TabsContent>

          <TabsContent value="files" className="mt-4">
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-2">
                <FileText className="w-3.5 h-3.5 text-muted-foreground" />
                <span className="text-xs font-mono uppercase text-muted-foreground">
                  pliki na węźle
                </span>
              </div>
              <Badge variant="secondary">{node.files_count}</Badge>
            </div>
            <div className="border border-border bg-muted/30 text-center py-8">
              <FileText className="w-5 h-5 mx-auto mb-2 text-muted-foreground" />
              <p className="text-xs font-mono text-muted-foreground">
                lista niedostępna
              </p>
              <p className="text-xs font-mono text-muted-foreground">
                wymaga implementacji
              </p>
            </div>
          </TabsContent>

          <TabsContent value="logs" className="mt-4">
            <div className="flex items-center gap-2 mb-2">
              <Activity className="w-3.5 h-3.5 text-muted-foreground" />
              <span className="text-xs font-mono uppercase text-muted-foreground">
                ostatnie logi
              </span>
            </div>
            <div className="border border-border bg-muted/30 text-center py-8">
              <Activity className="w-5 h-5 mx-auto mb-2 text-muted-foreground" />
              <p className="text-xs font-mono text-muted-foreground">
                logi niedostępne
              </p>
              <p className="text-xs font-mono text-muted-foreground">
                wymaga websocket
              </p>
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
    <div className="flex justify-between items-start gap-4 p-3">
      <span className="text-xs font-mono uppercase text-muted-foreground">
        {label}
      </span>
      <span
        className={`text-xs text-right ${mono ? "font-mono break-all" : ""}`}
      >
        {value}
      </span>
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
