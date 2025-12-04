import { MainLayout } from "@/components/MainLayout";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { formatBytes } from "@/lib/formatters";
import {
  Activity,
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
  uptime: 864000, // 10 days in seconds
};

const mockAlerts = [
  {
    id: "1",
    type: "node_offline",
    severity: "error" as const,
    message: "Węzeł d4e5f6a7... jest offline od 5 minut",
    timestamp: new Date(Date.now() - 300000).toISOString(),
  },
  {
    id: "2",
    type: "low_storage",
    severity: "warning" as const,
    message: "Węzeł c3d4e5f6... ma 80% wykorzystanej przestrzeni",
    timestamp: new Date(Date.now() - 600000).toISOString(),
  },
];

const mockLogs = [
  {
    level: "info",
    message: "Plik raport-2024.pdf przesłany pomyślnie",
    timestamp: new Date(Date.now() - 1000).toISOString(),
  },
  {
    level: "info",
    message: "Replikacja pliku do node2 zakończona",
    timestamp: new Date(Date.now() - 2000).toISOString(),
  },
  {
    level: "warning",
    message: "Wysokie wykorzystanie CPU na node3",
    timestamp: new Date(Date.now() - 5000).toISOString(),
  },
  {
    level: "info",
    message: "Heartbeat od node1 otrzymany",
    timestamp: new Date(Date.now() - 10000).toISOString(),
  },
  {
    level: "error",
    message: "Połączenie z node4 utracone",
    timestamp: new Date(Date.now() - 15000).toISOString(),
  },
  {
    level: "info",
    message: "Elekcja zakończona - master: node1",
    timestamp: new Date(Date.now() - 20000).toISOString(),
  },
  {
    level: "info",
    message: "Synchronizacja metadanych rozpoczęta",
    timestamp: new Date(Date.now() - 25000).toISOString(),
  },
  {
    level: "info",
    message: "Plik backup.zip usunięty",
    timestamp: new Date(Date.now() - 30000).toISOString(),
  },
  {
    level: "warning",
    message: "Opóźnienie replikacji > 500ms",
    timestamp: new Date(Date.now() - 35000).toISOString(),
  },
  {
    level: "info",
    message: "Nowy węzeł node5 dołączył do klastra",
    timestamp: new Date(Date.now() - 40000).toISOString(),
  },
];

export default function MonitoringPage() {
  const [metrics] = useState(mockMetrics);
  const [alerts, setAlerts] = useState(mockAlerts);
  const [logs] = useState(mockLogs);
  const [logFilter, setLogFilter] = useState("");
  const [logLevel, setLogLevel] = useState<string>("all");

  // Animated counter
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
      <div className="mb-8">
        <h1 className="text-2xl font-semibold">Monitoring</h1>
        <p className="text-muted-foreground mt-1">
          Monitoruj stan systemu i przeglądaj logi
        </p>
      </div>

      {/* System Health Cards */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <HealthCard
          icon={<FileText className="w-5 h-5" />}
          label="Pliki"
          value={displayedFiles.toLocaleString()}
          status="good"
        />
        <HealthCard
          icon={<Server className="w-5 h-5" />}
          label="Węzły aktywne"
          value={`${metrics.activeNodes}/${metrics.totalNodes}`}
          status={
            metrics.activeNodes === metrics.totalNodes ? "good" : "warning"
          }
        />
        <HealthCard
          icon={<HardDrive className="w-5 h-5" />}
          label="Przestrzeń"
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
        <HealthCard
          icon={<Clock className="w-5 h-5" />}
          label="Uptime"
          value={formatUptime(metrics.uptime)}
          status="good"
        />
      </div>

      <div className="grid lg:grid-cols-2 gap-6">
        {/* Alerts Panel */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-base flex items-center gap-2">
              <AlertTriangle className="w-4 h-4" />
              Alerty
            </CardTitle>
            <Badge variant={alerts.length > 0 ? "destructive" : "secondary"}>
              {alerts.length}
            </Badge>
          </CardHeader>
          <CardContent>
            {alerts.length > 0 ? (
              <div className="space-y-3">
                {alerts.map((alert) => (
                  <AlertItem
                    key={alert.id}
                    alert={alert}
                    onDismiss={() => dismissAlert(alert.id)}
                  />
                ))}
              </div>
            ) : (
              <div className="text-center py-8 text-muted-foreground">
                <CheckCircle2 className="w-8 h-8 mx-auto mb-2 text-green-500" />
                <p className="text-sm">Brak aktywnych alertów</p>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Quick Stats */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-base flex items-center gap-2">
              <Activity className="w-4 h-4" />
              Statystyki
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <StatRow label="Replikacje w toku" value="0" />
              <StatRow label="Pliki oczekujące" value="0" />
              <StatRow label="Średni czas odpowiedzi" value="45ms" />
              <StatRow label="Operacje/min" value="12" />
              <StatRow label="Transfer" value="2.4 MB/s" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Logs */}
      <Card className="mt-6">
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle className="text-base">Logi systemowe</CardTitle>
          <div className="flex items-center gap-2">
            <div className="relative">
              <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <Input
                placeholder="Szukaj..."
                value={logFilter}
                onChange={(e) => setLogFilter(e.target.value)}
                className="pl-8 h-8 w-[150px]"
              />
            </div>
            <select
              value={logLevel}
              onChange={(e) => setLogLevel(e.target.value)}
              className="h-8 px-2 rounded-md border bg-background text-sm"
            >
              <option value="all">Wszystkie</option>
              <option value="info">Info</option>
              <option value="warning">Warning</option>
              <option value="error">Error</option>
            </select>
            <Button variant="outline" size="sm">
              <Download className="w-4 h-4 mr-1" />
              Export
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="h-[300px] overflow-y-auto space-y-1 font-mono text-xs">
            {filteredLogs.map((log, i) => (
              <LogEntry key={i} log={log} />
            ))}
          </div>
        </CardContent>
      </Card>
    </MainLayout>
  );
}

function HealthCard({
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
    good: "text-green-500",
    warning: "text-amber-500",
    error: "text-red-500",
  };

  return (
    <Card>
      <CardContent className="p-4">
        <div className="flex items-start justify-between">
          <div className={statusColors[status]}>{icon}</div>
          <div
            className={`w-2 h-2 rounded-full ${
              status === "good"
                ? "bg-green-500"
                : status === "warning"
                  ? "bg-amber-500"
                  : "bg-red-500"
            }`}
          />
        </div>
        <div className="mt-3">
          <p className="text-2xl font-semibold">{value}</p>
          <p className="text-xs text-muted-foreground">{label}</p>
          {subtitle && (
            <p className="text-xs text-muted-foreground mt-1">{subtitle}</p>
          )}
        </div>
      </CardContent>
    </Card>
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
      className={`flex items-start gap-3 p-3 rounded-lg ${
        alert.severity === "error" ? "bg-red-500/10" : "bg-amber-500/10"
      }`}
    >
      {alert.severity === "error" ? (
        <XCircle className="w-4 h-4 text-red-500 shrink-0 mt-0.5" />
      ) : (
        <AlertTriangle className="w-4 h-4 text-amber-500 shrink-0 mt-0.5" />
      )}
      <div className="flex-1 min-w-0">
        <p className="text-sm">{alert.message}</p>
        <p className="text-xs text-muted-foreground mt-1">
          {new Date(alert.timestamp).toLocaleString("pl-PL")}
        </p>
      </div>
      <Button
        variant="ghost"
        size="icon"
        className="h-6 w-6"
        onClick={onDismiss}
      >
        <X className="w-3 h-3" />
      </Button>
    </div>
  );
}

function StatRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-sm text-muted-foreground">{label}</span>
      <span className="text-sm font-medium">{value}</span>
    </div>
  );
}

function LogEntry({
  log,
}: {
  log: { level: string; message: string; timestamp: string };
}) {
  const levelColors = {
    info: "text-blue-500",
    warning: "text-amber-500",
    error: "text-red-500",
  };

  return (
    <div className="flex items-start gap-2 py-1 px-2 hover:bg-muted/50 rounded">
      <span className="text-muted-foreground shrink-0">
        {new Date(log.timestamp).toLocaleTimeString("pl-PL")}
      </span>
      <span
        className={`uppercase font-medium shrink-0 w-14 ${levelColors[log.level as keyof typeof levelColors] || "text-gray-500"}`}
      >
        [{log.level}]
      </span>
      <span className="text-foreground">{log.message}</span>
    </div>
  );
}
