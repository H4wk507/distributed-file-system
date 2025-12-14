import { MainLayout } from "@/components/MainLayout";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { formatBytes } from "@/lib/formatters";
import {
  AlertTriangle,
  CheckCircle2,
  Clock,
  Download,
  FileText,
  HardDrive,
  Search,
  Server,
  X,
  XCircle,
} from "lucide-react";
import { useEffect, useState } from "react";

// Mock data
const mockMetrics = {
  totalFiles: 1247,
  activeNodes: 3,
  totalNodes: 4,
  storageUsed: 89743892480,
  storageTotal: 429496729600,
  uptime: 864000,
};

const mockAlerts = [
  {
    id: "1",
    type: "node_offline",
    severity: "error" as const,
    message: "węzeł d4e5f6a7... offline od 5 min",
    timestamp: new Date(Date.now() - 300000).toISOString(),
  },
  {
    id: "2",
    type: "low_storage",
    severity: "warning" as const,
    message: "węzeł c3d4e5f6... 80% przestrzeni",
    timestamp: new Date(Date.now() - 600000).toISOString(),
  },
];

const mockLogs = [
  {
    level: "info",
    message: "plik raport-2024.pdf przesłany",
    timestamp: new Date(Date.now() - 1000).toISOString(),
  },
  {
    level: "info",
    message: "replikacja do node2 zakończona",
    timestamp: new Date(Date.now() - 2000).toISOString(),
  },
  {
    level: "warning",
    message: "wysokie CPU na node3",
    timestamp: new Date(Date.now() - 5000).toISOString(),
  },
  {
    level: "info",
    message: "heartbeat od node1",
    timestamp: new Date(Date.now() - 10000).toISOString(),
  },
  {
    level: "error",
    message: "połączenie z node4 utracone",
    timestamp: new Date(Date.now() - 15000).toISOString(),
  },
  {
    level: "info",
    message: "elekcja zakończona - master: node1",
    timestamp: new Date(Date.now() - 20000).toISOString(),
  },
  {
    level: "info",
    message: "synchronizacja metadanych",
    timestamp: new Date(Date.now() - 25000).toISOString(),
  },
  {
    level: "info",
    message: "plik backup.zip usunięty",
    timestamp: new Date(Date.now() - 30000).toISOString(),
  },
  {
    level: "warning",
    message: "opóźnienie replikacji > 500ms",
    timestamp: new Date(Date.now() - 35000).toISOString(),
  },
  {
    level: "info",
    message: "nowy węzeł node5 dołączył",
    timestamp: new Date(Date.now() - 40000).toISOString(),
  },
];

export default function MonitoringPage() {
  const [metrics] = useState(mockMetrics);
  const [alerts, setAlerts] = useState(mockAlerts);
  const [logs] = useState(mockLogs);
  const [logFilter, setLogFilter] = useState("");
  const [logLevel, setLogLevel] = useState<string>("all");

  const [displayedFiles, setDisplayedFiles] = useState(0);

  useEffect(() => {
    const duration = 1000;
    const steps = 30;
    const increment = metrics.totalFiles / steps;
    let current = 0;
    const timer = setInterval(() => {
      current += increment;
      if (current >= metrics.totalFiles) {
        setDisplayedFiles(metrics.totalFiles);
        clearInterval(timer);
      } else {
        setDisplayedFiles(Math.floor(current));
      }
    }, duration / steps);
    return () => clearInterval(timer);
  }, [metrics.totalFiles]);

  const dismissAlert = (id: string) => {
    setAlerts((prev) => prev.filter((a) => a.id !== id));
  };

  const filteredLogs = logs.filter((log) => {
    const matchesFilter = log.message
      .toLowerCase()
      .includes(logFilter.toLowerCase());
    const matchesLevel = logLevel === "all" || log.level === logLevel;
    return matchesFilter && matchesLevel;
  });

  const storagePercent = Math.round(
    (metrics.storageUsed / metrics.storageTotal) * 100,
  );
  const formatUptime = (seconds: number) => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    return `${days}d ${hours}h`;
  };

  return (
    <MainLayout>
      {/* Nagłówek */}
      <div className="mb-6 pb-4 border-b border-border">
        <h1 className="text-sm font-mono uppercase tracking-wide text-primary">
          :: monitoring ::
        </h1>
        <p className="text-xs text-muted-foreground font-mono mt-1">
          stan systemu i logi
        </p>
      </div>

      {/* Health Stats */}
      <div className="grid grid-cols-4 gap-px bg-border mb-6">
        <HealthBlock
          icon={<FileText className="w-4 h-4" />}
          label="pliki"
          value={displayedFiles.toLocaleString()}
          status="good"
        />
        <HealthBlock
          icon={<Server className="w-4 h-4" />}
          label="węzły"
          value={`${metrics.activeNodes}/${metrics.totalNodes}`}
          status={
            metrics.activeNodes === metrics.totalNodes ? "good" : "warning"
          }
        />
        <HealthBlock
          icon={<HardDrive className="w-4 h-4" />}
          label="przestrzeń"
          value={`${storagePercent}%`}
          subtitle={`${formatBytes(metrics.storageUsed)} / ${formatBytes(metrics.storageTotal)}`}
          status={
            storagePercent > 90
              ? "error"
              : storagePercent > 70
                ? "warning"
                : "good"
          }
        />
        <HealthBlock
          icon={<Clock className="w-4 h-4" />}
          label="uptime"
          value={formatUptime(metrics.uptime)}
          status="good"
        />
      </div>

      <div className="grid lg:grid-cols-2 gap-6 mb-6">
        {/* Alerts */}
        <section>
          <div className="flex items-center gap-2 mb-3">
            <span className="text-xs font-mono uppercase tracking-wide text-muted-foreground">
              [alerty]
            </span>
            <Badge variant={alerts.length > 0 ? "destructive" : "secondary"}>
              {alerts.length}
            </Badge>
            <div className="flex-1 h-px bg-border" />
          </div>
          <div className="border border-border bg-card">
            {alerts.length > 0 ? (
              <div className="divide-y divide-border">
                {alerts.map((alert) => (
                  <AlertItem
                    key={alert.id}
                    alert={alert}
                    onDismiss={() => dismissAlert(alert.id)}
                  />
                ))}
              </div>
            ) : (
              <div className="text-center py-8">
                <CheckCircle2 className="w-5 h-5 mx-auto mb-2 text-success" />
                <p className="text-xs font-mono text-muted-foreground">
                  brak alertów
                </p>
              </div>
            )}
          </div>
        </section>

        {/* Quick Stats */}
        <section>
          <div className="flex items-center gap-2 mb-3">
            <span className="text-xs font-mono uppercase tracking-wide text-muted-foreground">
              [statystyki]
            </span>
            <div className="flex-1 h-px bg-border" />
          </div>
          <div className="border border-border bg-card divide-y divide-border">
            <StatRow label="replikacje_w_toku" value="0" />
            <StatRow label="pliki_oczekujące" value="0" />
            <StatRow label="avg_response_time" value="45ms" />
            <StatRow label="ops_per_min" value="12" />
            <StatRow label="transfer_rate" value="2.4 MB/s" />
          </div>
        </section>
      </div>

      {/* Logs */}
      <section>
        <div className="flex items-center gap-2 mb-3">
          <span className="text-xs font-mono uppercase tracking-wide text-muted-foreground">
            [logi systemowe]
          </span>
          <div className="flex-1 h-px bg-border" />
          <div className="flex items-center gap-2">
            <div className="relative">
              <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground" />
              <Input
                placeholder="szukaj..."
                value={logFilter}
                onChange={(e) => setLogFilter(e.target.value)}
                className="pl-7 h-7 w-32 text-xs"
              />
            </div>
            <select
              value={logLevel}
              onChange={(e) => setLogLevel(e.target.value)}
              className="h-7 px-2 border border-border bg-card text-xs font-mono"
            >
              <option value="all">all</option>
              <option value="info">info</option>
              <option value="warning">warn</option>
              <option value="error">err</option>
            </select>
            <Button variant="outline" size="sm" className="h-7 text-xs">
              <Download className="w-3.5 h-3.5 mr-1" />
              export
            </Button>
          </div>
        </div>
        <div className="border border-border bg-card">
          <div className="h-[280px] overflow-y-auto font-mono text-xs">
            {filteredLogs.map((log, i) => (
              <LogEntry key={i} log={log} />
            ))}
          </div>
        </div>
      </section>
    </MainLayout>
  );
}

function HealthBlock({
  icon,
  label,
  value,
  subtitle,
  status,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
  subtitle?: string;
  status: "good" | "warning" | "error";
}) {
  const statusColors = {
    good: "text-success",
    warning: "text-primary",
    error: "text-destructive",
  };

  return (
    <div className="bg-card p-3">
      <div className="flex items-center justify-between mb-1">
        <div className={`flex items-center gap-2 ${statusColors[status]}`}>
          {icon}
          <span className="text-xs font-mono uppercase text-muted-foreground">
            {label}
          </span>
        </div>
        <div
          className={`w-2 h-2 ${
            status === "good"
              ? "bg-success"
              : status === "warning"
                ? "bg-primary"
                : "bg-destructive"
          }`}
        />
      </div>
      <p className={`text-lg font-mono ${statusColors[status]}`}>{value}</p>
      {subtitle && (
        <p className="text-xs text-muted-foreground font-mono mt-0.5">
          {subtitle}
        </p>
      )}
    </div>
  );
}

function AlertItem({
  alert,
  onDismiss,
}: {
  alert: {
    id: string;
    severity: "error" | "warning";
    message: string;
    timestamp: string;
  };
  onDismiss: () => void;
}) {
  return (
    <div
      className={`flex items-start gap-3 p-3 ${
        alert.severity === "error"
          ? "border-l-2 border-l-destructive"
          : "border-l-2 border-l-primary"
      }`}
    >
      {alert.severity === "error" ? (
        <XCircle className="w-4 h-4 text-destructive shrink-0 mt-0.5" />
      ) : (
        <AlertTriangle className="w-4 h-4 text-primary shrink-0 mt-0.5" />
      )}
      <div className="flex-1 min-w-0">
        <p className="text-xs font-mono">{alert.message}</p>
        <p className="text-xs text-muted-foreground font-mono mt-0.5">
          {new Date(alert.timestamp).toLocaleTimeString("pl-PL")}
        </p>
      </div>
      <button onClick={onDismiss} className="p-1 hover:bg-muted">
        <X className="w-3 h-3 text-muted-foreground" />
      </button>
    </div>
  );
}

function StatRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between px-3 py-2">
      <span className="text-xs font-mono text-muted-foreground">{label}</span>
      <span className="text-xs font-mono text-primary">{value}</span>
    </div>
  );
}

function LogEntry({
  log,
}: {
  log: { level: string; message: string; timestamp: string };
}) {
  const levelColors: Record<string, string> = {
    info: "text-muted-foreground",
    warning: "text-primary",
    error: "text-destructive",
  };

  return (
    <div className="flex items-start gap-2 px-3 py-1.5 hover:bg-muted/30 border-b border-border last:border-b-0">
      <span className="text-muted-foreground shrink-0 w-16">
        {new Date(log.timestamp).toLocaleTimeString("pl-PL")}
      </span>
      <span
        className={`uppercase shrink-0 w-10 ${levelColors[log.level] || "text-muted-foreground"}`}
      >
        {log.level === "warning"
          ? "warn"
          : log.level === "error"
            ? "err"
            : log.level}
      </span>
      <span className="text-foreground">{log.message}</span>
    </div>
  );
}
