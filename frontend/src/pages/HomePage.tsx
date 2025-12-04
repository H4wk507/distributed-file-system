import type { FileInfo } from "@/api/types";
import { FileList } from "@/components/files/FileList";
import { FileUploader } from "@/components/files/FileUploader";
import { MainLayout } from "@/components/MainLayout";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { formatBytes } from "@/lib/formatters";
import { Copy, FileText, HardDrive, Server } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

// TODO: remove
const mockFiles: FileInfo[] = [
  {
    id: "1",
    filename: "raport-roczny-2024.pdf",
    size: 2457600,
    content_type: "application/pdf",
    hash: "abc123",
    owner_id: "user1",
    replicas: ["node1", "node2", "node3"],
    created_at: "2024-12-01T10:30:00Z",
    updated_at: "2024-12-01T10:30:00Z",
  },
  {
    id: "2",
    filename: "zdjecie-wakacje.jpg",
    size: 4521984,
    content_type: "image/jpeg",
    hash: "def456",
    owner_id: "user1",
    replicas: ["node1", "node2"],
    created_at: "2024-11-28T15:45:00Z",
    updated_at: "2024-11-28T15:45:00Z",
  },
  {
    id: "3",
    filename: "prezentacja-projekt.pptx",
    size: 8912345,
    content_type:
      "application/vnd.openxmlformats-officedocument.presentationml.presentation",
    hash: "ghi789",
    owner_id: "user1",
    replicas: ["node2", "node3"],
    created_at: "2024-11-25T09:15:00Z",
    updated_at: "2024-11-26T14:20:00Z",
  },
  {
    id: "4",
    filename: "dane-sprzedaz.xlsx",
    size: 156789,
    content_type:
      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
    hash: "jkl012",
    owner_id: "user1",
    replicas: ["node1", "node3"],
    created_at: "2024-11-20T11:00:00Z",
    updated_at: "2024-11-20T11:00:00Z",
  },
  {
    id: "5",
    filename: "backup-baza-danych.zip",
    size: 52428800,
    content_type: "application/zip",
    hash: "mno345",
    owner_id: "user1",
    replicas: ["node1", "node2", "node3"],
    created_at: "2024-11-15T23:00:00Z",
    updated_at: "2024-11-15T23:00:00Z",
  },
  {
    id: "6",
    filename: "nagranie-spotkanie.mp4",
    size: 157286400,
    content_type: "video/mp4",
    hash: "pqr678",
    owner_id: "user1",
    replicas: ["node2"],
    created_at: "2024-11-10T14:30:00Z",
    updated_at: "2024-11-10T14:30:00Z",
  },
  {
    id: "7",
    filename: "notatki.txt",
    size: 4521,
    content_type: "text/plain",
    hash: "stu901",
    owner_id: "user1",
    replicas: ["node1", "node2", "node3"],
    created_at: "2024-11-05T08:45:00Z",
    updated_at: "2024-12-02T16:30:00Z",
  },
  {
    id: "8",
    filename: "logo-firma.png",
    size: 89456,
    content_type: "image/png",
    hash: "vwx234",
    owner_id: "user1",
    replicas: ["node1", "node2"],
    created_at: "2024-10-20T10:00:00Z",
    updated_at: "2024-10-20T10:00:00Z",
  },
  {
    id: "9",
    filename: "logo-firma2.png",
    size: 89456,
    content_type: "image/png",
    hash: "vwx2345",
    owner_id: "user1",
    replicas: ["node1", "node2"],
    created_at: "2024-10-20T10:00:00Z",
    updated_at: "2024-10-20T10:00:00Z",
  },
  {
    id: "10",
    filename: "logo-firma3.png",
    size: 89456,
    content_type: "image/png",
    hash: "vwx23456",
    owner_id: "user1",
    replicas: ["node1", "node2"],
    created_at: "2024-10-20T10:00:00Z",
    updated_at: "2024-10-20T10:00:00Z",
  },
  {
    id: "11",
    filename: "logo-firma4.png",
    size: 89456,
    content_type: "image/png",
    hash: "vwx234567",
    owner_id: "user1",
    replicas: ["node1", "node2"],
    created_at: "2024-10-20T10:00:00Z",
    updated_at: "2024-10-20T10:00:00Z",
  },
];

export default function HomePage() {
  const [page, setPage] = useState(1);
  const [files, setFiles] = useState(mockFiles);
  const [isDeleting, setIsDeleting] = useState(false);
  const perPage = 10;

  const total = files.length;
  const totalPages = Math.ceil(total / perPage);

  const totalSize = files.reduce((acc, f) => acc + f.size, 0);
  const totalReplicas = files.reduce(
    (acc, f) => acc + (f.replicas?.length || 0),
    0,
  );

  const handleDelete = (fileId: string) => {
    setIsDeleting(true);
    setTimeout(() => {
      setFiles((prev) => prev.filter((f) => f.id !== fileId));
      toast.success("Plik usunięty");
      setIsDeleting(false);
    }, 500);
  };

  const handleDownload = (_fileId: string, filename: string) => {
    toast.success("Pobieranie rozpoczęte", { description: filename });
  };

  return (
    <MainLayout>
      <div className="mb-8">
        <h1 className="text-2xl font-semibold">Pliki</h1>
        <p className="text-muted-foreground mt-1">
          Zarządzaj swoimi plikami w systemie rozproszonym
        </p>
      </div>

      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
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
        <StatCard
          icon={<Copy className="w-5 h-5" />}
          label="Repliki"
          value={totalReplicas.toString()}
        />
      </div>

      <Card className="mb-8">
        <CardHeader>
          <CardTitle className="text-base">Prześlij pliki</CardTitle>
        </CardHeader>
        <CardContent>
          <FileUploader
            onUploadComplete={(file) => {
              const newFile: FileInfo = {
                id: Math.random().toString(36).substring(7),
                filename: file.name,
                size: file.size,
                content_type: file.type || "application/octet-stream",
                hash: Math.random().toString(36).substring(7),
                owner_id: "user1",
                replicas: ["node1", "node2"],
                created_at: new Date().toISOString(),
                updated_at: new Date().toISOString(),
              };
              setFiles((prev) => [newFile, ...prev]);
            }}
          />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Twoje pliki</CardTitle>
        </CardHeader>
        <CardContent>
          {files.length > 0 ? (
            <FileList
              files={files}
              page={page}
              totalPages={totalPages}
              onPageChange={setPage}
              onDelete={handleDelete}
              onDownload={handleDownload}
              isDeleting={isDeleting}
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
