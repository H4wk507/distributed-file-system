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
        toast.error(`błąd: ${file.name}`, {
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
    <div className="space-y-3">
      {/* Dropzone - surowy styl */}
      <div
        {...getRootProps()}
        className={`border-2 border-dashed p-6 text-center cursor-pointer transition-colors ${
          isDragActive
            ? "border-primary bg-primary/5"
            : "border-border hover:border-muted-foreground"
        }`}
      >
        <input {...getInputProps()} />
        <Upload
          className={`w-6 h-6 mx-auto mb-2 ${
            isDragActive ? "text-primary" : "text-muted-foreground"
          }`}
        />
        {isDragActive ? (
          <p className="text-xs text-primary font-mono">upuść pliki</p>
        ) : (
          <p className="text-xs text-muted-foreground font-mono">
            upuść pliki lub kliknij
          </p>
        )}
      </div>

      {/* Uploads list */}
      {uploads.length > 0 && (
        <div className="border border-border divide-y divide-border">
          {uploads.map((upload) => (
            <div
              key={upload.id}
              className="flex items-center gap-3 p-3 bg-card"
            >
              <div className="shrink-0">
                {upload.status === "completed" ? (
                  <CheckCircle2 className="w-4 h-4 text-success" />
                ) : upload.status === "error" ? (
                  <XCircle className="w-4 h-4 text-destructive" />
                ) : upload.status === "cancelled" ? (
                  <XCircle className="w-4 h-4 text-muted-foreground" />
                ) : (
                  <File className="w-4 h-4 text-muted-foreground" />
                )}
              </div>

              <div className="flex-1 min-w-0">
                <p className="text-xs font-mono truncate">{upload.fileName}</p>
                <div className="flex items-center gap-2 mt-1">
                  <Progress value={upload.progress} className="flex-1 h-1" />
                  <span className="text-xs font-mono text-muted-foreground shrink-0 w-8 text-right">
                    {Math.round(upload.progress)}%
                  </span>
                </div>
                {upload.error && (
                  <p className="text-xs text-destructive font-mono mt-1">
                    err: {upload.error}
                  </p>
                )}
                {upload.status === "cancelled" && (
                  <p className="text-xs text-muted-foreground font-mono mt-1">
                    anulowano
                  </p>
                )}
              </div>

              {(upload.status === "uploading" ||
                upload.status === "pending") && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="shrink-0 h-6 w-6"
                  onClick={() => cancelUpload(upload.id)}
                  title="anuluj"
                >
                  <X className="w-3.5 h-3.5" />
                </Button>
              )}
              {(upload.status === "completed" ||
                upload.status === "error" ||
                upload.status === "cancelled") && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="shrink-0 h-6 w-6"
                  onClick={() => clearUpload(upload.id)}
                  title="usuń"
                >
                  <X className="w-3.5 h-3.5" />
                </Button>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
