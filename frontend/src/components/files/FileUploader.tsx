import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import { useFileUpload } from "@/hooks/useFileUpload";
import { CheckCircle2, File, Upload, X, XCircle } from "lucide-react";
import { useCallback } from "react";
import { useDropzone } from "react-dropzone";
import { toast } from "sonner";

interface FileUploaderProps {
  acceptedTypes?: string[];
}

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
  acceptedTypes = DEFAULT_ACCEPTED_TYPES,
}: FileUploaderProps) {
  const { uploads, uploadFile, cancelUpload, clearUpload } = useFileUpload();

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
        uploadFile(file);
      });
    },
    [uploadFile],
  );

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: acceptedTypes.reduce(
      (acc, type) => {
        acc[type] = [];
        return acc;
      },
      {} as Record<string, string[]>,
    ),
  });

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
          </>
        )}
      </div>
      {uploads.length > 0 && (
        <div className="space-y-3">
          {uploads.map((upload) => (
            <div
              key={upload.id}
              className="flex items-center gap-3 p-3 rounded-lg border bg-card"
            >
              <div className="shrink-0">
                {upload.status === "completed" ? (
                  <CheckCircle2 className="w-5 h-5 text-green-500" />
                ) : upload.status === "error" ? (
                  <XCircle className="w-5 h-5 text-destructive" />
                ) : upload.status === "cancelled" ? (
                  <XCircle className="w-5 h-5 text-muted-foreground" />
                ) : (
                  <File className="w-5 h-5 text-muted-foreground" />
                )}
              </div>

              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium truncate">
                  {upload.fileName}
                </p>
                <div className="flex items-center gap-2 mt-1">
                  <Progress value={upload.progress} className="flex-1" />
                  <span className="text-xs text-muted-foreground shrink-0">
                    {Math.round(upload.progress)}%
                  </span>
                </div>
                {upload.error && (
                  <p className="text-xs text-destructive mt-1">
                    {upload.error}
                  </p>
                )}
                {upload.status === "cancelled" && (
                  <p className="text-xs text-muted-foreground mt-1">
                    Anulowano
                  </p>
                )}
              </div>

              {(upload.status === "uploading" ||
                upload.status === "pending") && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="shrink-0 h-8 w-8"
                  onClick={() => cancelUpload(upload.id)}
                  title="Anuluj"
                >
                  <X className="w-4 h-4" />
                </Button>
              )}
              {(upload.status === "completed" ||
                upload.status === "error" ||
                upload.status === "cancelled") && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="shrink-0 h-8 w-8"
                  onClick={() => clearUpload(upload.id)}
                  title="Usuń"
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
