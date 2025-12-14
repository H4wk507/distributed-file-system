import { FileList } from "@/components/files/FileList";
import { FileUploader } from "@/components/files/FileUploader";
import { MainLayout } from "@/components/MainLayout";
import { Progress } from "@/components/ui/progress";
import { useFileDownload } from "@/hooks/useFileDownload";
import { useDeleteFile, useFiles } from "@/hooks/useFiles";
import { formatBytes } from "@/lib/formatters";
import {
  CheckCircle,
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
      {/* Nagłówek - surowy styl */}
      <div className="mb-6 pb-4 border-b border-border">
        <h1 className="text-sm font-mono uppercase tracking-wide text-primary">
          :: pliki ::
        </h1>
        <p className="text-xs text-muted-foreground font-mono mt-1">
          zarządzaj plikami w systemie rozproszonym
        </p>
      </div>

      {/* Stats - asymetryczny grid, bez kart */}
      <div className="grid grid-cols-3 gap-px bg-border mb-6">
        <StatBlock
          icon={<FileText className="w-4 h-4" />}
          label="pliki"
          value={total.toString()}
        />
        <StatBlock
          icon={<HardDrive className="w-4 h-4" />}
          label="przestrzeń"
          value={formatBytes(totalSize)}
        />
        <StatBlock
          icon={<Server className="w-4 h-4" />}
          label="węzły"
          value="3"
        />
      </div>

      {/* Upload - prostszy */}
      <section className="mb-6">
        <div className="flex items-center gap-2 mb-3">
          <span className="text-xs font-mono uppercase tracking-wide text-muted-foreground">
            [upload]
          </span>
          <div className="flex-1 h-px bg-border" />
        </div>
        <FileUploader />
      </section>

      {/* Downloads */}
      {downloads.length > 0 && (
        <section className="mb-6">
          <div className="flex items-center gap-2 mb-3">
            <span className="text-xs font-mono uppercase tracking-wide text-muted-foreground">
              [pobieranie{" "}
              {activeDownloads.length > 0 ? `(${activeDownloads.length})` : ""}]
            </span>
            <div className="flex-1 h-px bg-border" />
          </div>
          <div className="border border-border divide-y divide-border">
            {downloads.map((dl) => (
              <div key={dl.id} className="flex items-center gap-4 p-3 bg-card">
                <div className="flex-1 min-w-0">
                  <p className="text-xs font-mono truncate">{dl.fileName}</p>
                  <div className="flex items-center gap-2 mt-1.5">
                    <Progress value={dl.progress} className="flex-1 h-1" />
                    <span className="text-xs font-mono text-muted-foreground w-10 text-right">
                      {dl.progress}%
                    </span>
                  </div>
                  {dl.status === "error" && (
                    <p className="text-xs text-destructive font-mono mt-1">
                      err: {dl.error}
                    </p>
                  )}
                </div>
                <div className="flex items-center gap-1">
                  {dl.status === "completed" && (
                    <CheckCircle className="w-4 h-4 text-success" />
                  )}
                  {dl.status === "error" && (
                    <XCircle className="w-4 h-4 text-destructive" />
                  )}
                  {(dl.status === "downloading" || dl.status === "pending") && (
                    <button
                      onClick={() => cancelDownload(dl.id)}
                      className="p-1 hover:bg-muted"
                      title="Anuluj"
                    >
                      <X className="w-3.5 h-3.5 text-muted-foreground" />
                    </button>
                  )}
                  {(dl.status === "completed" ||
                    dl.status === "error" ||
                    dl.status === "cancelled") && (
                    <button
                      onClick={() => clearDownload(dl.id)}
                      className="p-1 hover:bg-muted"
                      title="Usuń"
                    >
                      <X className="w-3.5 h-3.5 text-muted-foreground" />
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        </section>
      )}

      {/* File list */}
      <section>
        <div className="flex items-center gap-2 mb-3">
          <span className="text-xs font-mono uppercase tracking-wide text-muted-foreground">
            [lista plików]
          </span>
          <div className="flex-1 h-px bg-border" />
        </div>
        <div className="border border-border">
          {isLoading ? (
            <div className="flex items-center justify-center py-12 bg-card">
              <Loader2 className="w-5 h-5 animate-spin text-muted-foreground" />
              <span className="ml-2 text-xs font-mono text-muted-foreground">
                ładowanie...
              </span>
            </div>
          ) : isError ? (
            <div className="text-center py-12 bg-card">
              <XCircle className="w-5 h-5 mx-auto mb-2 text-destructive" />
              <p className="text-xs font-mono text-destructive">
                błąd ładowania plików
              </p>
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
            <div className="text-center py-12 bg-card">
              <FileText className="w-5 h-5 mx-auto mb-2 text-muted-foreground" />
              <p className="text-xs font-mono text-muted-foreground">
                brak plików
              </p>
            </div>
          )}
        </div>
      </section>
    </MainLayout>
  );
}

function StatBlock({
  icon,
  label,
  value,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
}) {
  return (
    <div className="bg-card p-3">
      <div className="flex items-center gap-2 text-muted-foreground mb-1">
        {icon}
        <span className="text-xs font-mono uppercase">{label}</span>
      </div>
      <p className="text-lg font-mono text-primary">{value}</p>
    </div>
  );
}
