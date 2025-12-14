import { FileList } from "@/components/files/FileList";
import { FileUploader } from "@/components/files/FileUploader";
import { MainLayout } from "@/components/MainLayout";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { useFileDownload } from "@/hooks/useFileDownload";
import { useDeleteFile, useFiles } from "@/hooks/useFiles";
import { formatBytes } from "@/lib/formatters";
import {
  CheckCircle,
  Download,
  FileText,
  HardDrive,
  Loader2,
  Server,
  X,
  XCircle,
} from "lucide-react";
import { useState } from "react";

export default function HomePage() {
  const [page, setPage] = useState(1);
  const perPage = 10;

  const { data, isLoading, isError } = useFiles(page, perPage);
  const deleteFileMutation = useDeleteFile();

  const { downloads, downloadFile, cancelDownload, clearDownload } =
    useFileDownload();

  const activeDownloads = downloads.filter(
    (d) => d.status === "downloading" || d.status === "pending",
  );

  const files = data?.files ?? [];
  const total = data?.total ?? 0;
  const totalPages = Math.ceil(total / perPage);

  const totalSize = files.reduce((acc, f) => acc + f.size, 0);

  const handleDelete = (fileId: string) => {
    deleteFileMutation.mutate(fileId);
  };

  const handleDownload = (fileId: string, filename: string) => {
    downloadFile(fileId, filename);
  };

  return (
    <MainLayout>
      <div className="mb-8">
        <h1 className="text-2xl font-semibold">Pliki</h1>
        <p className="text-muted-foreground mt-1">
          Zarządzaj swoimi plikami w systemie rozproszonym
        </p>
      </div>

      <div className="grid grid-cols-2 lg:grid-cols-3 gap-4 mb-8">
        <StatCard
          icon={<FileText className="w-5 h-5" />}
          label="Pliki"
          value={total.toString()}
        />
        <StatCard
          icon={<HardDrive className="w-5 h-5" />}
          label="Przestrzeń"
          value={formatBytes(totalSize)}
        />
        <StatCard
          icon={<Server className="w-5 h-5" />}
          label="Węzły"
          value="3"
        />
      </div>

      <Card className="mb-8">
        <CardHeader>
          <CardTitle className="text-base">Prześlij pliki</CardTitle>
        </CardHeader>
        <CardContent>
          <FileUploader />
        </CardContent>
      </Card>

      {downloads.length > 0 && (
        <Card className="mb-8">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <Download className="w-4 h-4" />
              Pobieranie
              {activeDownloads.length > 0 && ` (${activeDownloads.length})`}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {downloads.map((dl) => (
              <div
                key={dl.id}
                className="flex items-center gap-4 p-3 bg-muted/50 rounded-lg"
              >
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium truncate">{dl.fileName}</p>
                  <div className="flex items-center gap-2 mt-1">
                    <Progress value={dl.progress} className="flex-1 h-2" />
                    <span className="text-xs text-muted-foreground w-12 text-right">
                      {dl.progress}%
                    </span>
                  </div>
                  {dl.status === "error" && (
                    <p className="text-xs text-destructive mt-1">
                      Błąd: {dl.error}
                    </p>
                  )}
                </div>
                <div className="flex items-center gap-1">
                  {dl.status === "completed" && (
                    <CheckCircle className="w-5 h-5 text-green-500" />
                  )}
                  {dl.status === "error" && (
                    <XCircle className="w-5 h-5 text-destructive" />
                  )}
                  {(dl.status === "downloading" || dl.status === "pending") && (
                    <button
                      onClick={() => cancelDownload(dl.id)}
                      className="p-1 hover:bg-muted rounded"
                      title="Anuluj"
                    >
                      <X className="w-4 h-4 text-muted-foreground" />
                    </button>
                  )}
                  {(dl.status === "completed" ||
                    dl.status === "error" ||
                    dl.status === "cancelled") && (
                    <button
                      onClick={() => clearDownload(dl.id)}
                      className="p-1 hover:bg-muted rounded"
                      title="Usuń"
                    >
                      <X className="w-4 h-4 text-muted-foreground" />
                    </button>
                  )}
                </div>
              </div>
            ))}
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Twoje pliki</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
            </div>
          ) : isError ? (
            <div className="text-center py-12 text-muted-foreground">
              <XCircle className="w-8 h-8 mx-auto mb-3 opacity-50 text-destructive" />
              <p className="text-sm">Błąd ładowania plików</p>
            </div>
          ) : files.length > 0 ? (
            <FileList
              files={files}
              page={page}
              totalPages={totalPages}
              onPageChange={setPage}
              onDelete={handleDelete}
              onDownload={handleDownload}
              isDeleting={deleteFileMutation.isPending}
            />
          ) : (
            <div className="text-center py-12 text-muted-foreground">
              <FileText className="w-8 h-8 mx-auto mb-3 opacity-50" />
              <p className="text-sm">Brak plików</p>
            </div>
          )}
        </CardContent>
      </Card>
    </MainLayout>
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
