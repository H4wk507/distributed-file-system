import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuth } from "@/context/AuthContext";
import { useLogout } from "@/hooks/useAuth";
import {
  Copy,
  FileText,
  HardDrive,
  LogOut,
  Server,
  Upload,
} from "lucide-react";

export default function HomePage() {
  const { user } = useAuth();
  const logout = useLogout();

  return (
    <div className="min-h-screen flex flex-col">
      <header className="border-b bg-card/80 backdrop-blur-sm sticky top-0 z-50">
        <div className="max-w-5xl mx-auto px-6">
          <div className="flex items-center justify-between h-14">
            <div className="flex items-center gap-2">
              <Server className="w-5 h-5" />
              <span className="font-semibold">RSP</span>
            </div>

            <div className="flex items-center gap-4">
              <span className="text-sm text-muted-foreground">
                {user?.email}
              </span>
              <Button variant="ghost" size="sm" onClick={logout}>
                <LogOut className="w-4 h-4" />
              </Button>
            </div>
          </div>
        </div>
      </header>
      <main className="flex-1 max-w-5xl mx-auto px-6 py-8 w-full">
        <div className="mb-8">
          <h1 className="text-2xl font-semibold">Witaj, {user?.username}</h1>
          <p className="text-muted-foreground mt-1">Zarządzaj swoimi plikami</p>
        </div>
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
          <StatCard
            icon={<FileText className="w-5 h-5" />}
            label="Pliki"
            value="—"
          />
          <StatCard
            icon={<HardDrive className="w-5 h-5" />}
            label="Przestrzeń"
            value="—"
          />
          <StatCard
            icon={<Server className="w-5 h-5" />}
            label="Węzły"
            value="—"
          />
          <StatCard
            icon={<Copy className="w-5 h-5" />}
            label="Repliki"
            value="—"
          />
        </div>
        <Card className="mb-8">
          <CardHeader>
            <CardTitle className="text-base">Prześlij</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="border-2 border-dashed rounded-lg p-8 text-center hover:border-muted-foreground/50 transition-colors cursor-pointer">
              <Upload className="w-8 h-8 mx-auto text-muted-foreground mb-3" />
              <p className="text-sm text-muted-foreground">
                Upuść pliki tutaj lub kliknij, aby przeglądać
              </p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Pliki</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-center py-12 text-muted-foreground">
              <FileText className="w-8 h-8 mx-auto mb-3 opacity-50" />
              <p className="text-sm">Brak plików</p>
            </div>
          </CardContent>
        </Card>
      </main>
    </div>
  );
}

function StatCard({
  icon,
  label,
  value,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
}) {
  return (
    <Card>
      <CardContent className="p-4">
        <div className="flex items-center gap-3">
          <div className="text-muted-foreground">{icon}</div>
          <div>
            <p className="text-xs text-muted-foreground">{label}</p>
            <p className="text-lg font-semibold">{value}</p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
