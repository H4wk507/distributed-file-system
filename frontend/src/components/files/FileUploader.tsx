import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import { formatBytes } from "@/lib/formatters";
import { CheckCircle2, File, Upload, X, XCircle } from "lucide-react";
import { useCallback, useState } from "react";
import { useDropzone } from "react-dropzone";
import { toast } from "sonner";

interface UploadingFile {
  id: string;
  file: File;
  progress: number;
  status: "uploading" | "success" | "error";
  error?: string;
}

interface FileUploaderProps {
  onUploadComplete?: (file: File) => void;
  maxSize?: number; // in bytes
  acceptedTypes?: string[];
}

const DEFAULT_MAX_SIZE = 100 * 1024 * 1024; // 100MB
const DEFAULT_ACCEPTED_TYPES = [
  "image/*",
  "video/*",
  "audio/*",
  "application/pdf",
  "application/zip",
  "application/x-rar-compressed",
  "application/x-7z-compressed",
  "application/msword",
  "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
  "application/vnd.ms-excel",
  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
  "application/vnd.ms-powerpoint",
  "application/vnd.openxmlformats-officedocument.presentationml.presentation",
  "text/*",
];

export function FileUploader({
  onUploadComplete,
  maxSize = DEFAULT_MAX_SIZE,
  acceptedTypes = DEFAULT_ACCEPTED_TYPES,
}: FileUploaderProps) {
  const [uploadingFiles, setUploadingFiles] = useState<UploadingFile[]>([]);

  const simulateUpload = useCallback(
    (file: File) => {
      const id = Math.random().toString(36).substring(7);

      setUploadingFiles((prev) => [
        ...prev,
        { id, file, progress: 0, status: "uploading" },
      ]);

      let progress = 0;
      const interval = setInterval(() => {
        progress += Math.random() * 30;
        if (progress >= 100) {
          progress = 100;
          clearInterval(interval);

          setUploadingFiles((prev) =>
            prev.map((f) =>
              f.id === id ? { ...f, progress: 100, status: "success" } : f,
            ),
          );

          toast.success("Plik przesłany", { description: file.name });
          onUploadComplete?.(file);

          setTimeout(() => {
            setUploadingFiles((prev) => prev.filter((f) => f.id !== id));
          }, 2000);
        } else {
          setUploadingFiles((prev) =>
            prev.map((f) => (f.id === id ? { ...f, progress } : f)),
          );
        }
      }, 200);

      return () => clearInterval(interval);
    },
    [onUploadComplete],
  );

  const onDrop = useCallback(
    (
      acceptedFiles: File[],
      fileRejections: { file: File; errors: readonly { message: string }[] }[],
    ) => {
      fileRejections.forEach(({ file, errors }) => {
        const errorMessages = errors.map((e) => e.message).join(", ");
        toast.error(`Nie można przesłać: ${file.name}`, {
          description: errorMessages,
        });
      });

      acceptedFiles.forEach((file) => {
        simulateUpload(file);
      });
    },
    [simulateUpload],
  );

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    maxSize,
    accept: acceptedTypes.reduce(
      (acc, type) => {
        acc[type] = [];
        return acc;
      },
      {} as Record<string, string[]>,
    ),
  });

  const cancelUpload = (id: string) => {
    setUploadingFiles((prev) => prev.filter((f) => f.id !== id));
    toast.info("Przesyłanie anulowane");
  };

  return (
    <div className="space-y-4">
      <div
        {...getRootProps()}
        className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors cursor-pointer ${
          isDragActive
            ? "border-primary bg-primary/5"
            : "hover:border-muted-foreground/50"
        }`}
      >
        <input {...getInputProps()} />
        <Upload
          className={`w-8 h-8 mx-auto mb-3 ${
            isDragActive ? "text-primary" : "text-muted-foreground"
          }`}
        />
        {isDragActive ? (
          <p className="text-sm text-primary font-medium">Upuść pliki tutaj</p>
        ) : (
          <>
            <p className="text-sm text-muted-foreground">
              Upuść pliki tutaj lub kliknij, aby przeglądać
            </p>
            <p className="text-xs text-muted-foreground mt-1">
              Maksymalny rozmiar: {formatBytes(maxSize)}
            </p>
          </>
        )}
      </div>
      {uploadingFiles.length > 0 && (
        <div className="space-y-3">
          {uploadingFiles.map((uploadFile) => (
            <div
              key={uploadFile.id}
              className="flex items-center gap-3 p-3 rounded-lg border bg-card"
            >
              <div className="shrink-0">
                {uploadFile.status === "success" ? (
                  <CheckCircle2 className="w-5 h-5 text-green-500" />
                ) : uploadFile.status === "error" ? (
                  <XCircle className="w-5 h-5 text-destructive" />
                ) : (
                  <File className="w-5 h-5 text-muted-foreground" />
                )}
              </div>

              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium truncate">
                  {uploadFile.file.name}
                </p>
                <div className="flex items-center gap-2 mt-1">
                  <Progress value={uploadFile.progress} className="flex-1" />
                  <span className="text-xs text-muted-foreground shrink-0">
                    {Math.round(uploadFile.progress)}%
                  </span>
                </div>
                {uploadFile.error && (
                  <p className="text-xs text-destructive mt-1">
                    {uploadFile.error}
                  </p>
                )}
              </div>

              {uploadFile.status === "uploading" && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="shrink-0 h-8 w-8"
                  onClick={() => cancelUpload(uploadFile.id)}
                >
                  <X className="w-4 h-4" />
                </Button>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
