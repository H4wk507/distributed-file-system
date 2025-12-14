import type { FileInfo } from "@/api/types";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { getFileIcon, type FileType } from "@/lib/file";
import { formatBytes } from "@/lib/formatters";
import {
  Archive,
  Download,
  File,
  FileAudio,
  FileImage,
  FileSpreadsheet,
  FileText,
  FileVideo,
} from "lucide-react";

const fileIcons: Record<FileType, React.ReactNode> = {
  image: <FileImage className="w-12 h-12 text-primary" />,
  video: <FileVideo className="w-12 h-12 text-primary" />,
  audio: <FileAudio className="w-12 h-12 text-success" />,
  pdf: <FileText className="w-12 h-12 text-destructive" />,
  archive: <Archive className="w-12 h-12 text-primary" />,
  doc: <FileText className="w-12 h-12 text-muted-foreground" />,
  spreadsheet: <FileSpreadsheet className="w-12 h-12 text-success" />,
  text: <FileText className="w-12 h-12 text-muted-foreground" />,
  file: <File className="w-12 h-12 text-muted-foreground" />,
};

interface FilePreviewProps {
  file: FileInfo | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onDownload: (fileId: string, filename: string) => void;
  previewUrl?: string;
}

export function FilePreview({
  file,
  open,
  onOpenChange,
  onDownload,
  previewUrl,
}: FilePreviewProps) {
  if (!file) return null;

  const fileType = getFileIcon(file.content_type);
  const isImage = fileType === "image";
  const isPdf = fileType === "pdf";
  const isVideo = fileType === "video";
  const isAudio = fileType === "audio";

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl max-h-[90vh] overflow-hidden flex flex-col">
        <DialogHeader>
          <DialogTitle className="truncate pr-8 font-mono text-xs">
            {file.filename}
          </DialogTitle>
        </DialogHeader>

        <div className="flex-1 overflow-auto min-h-0 border border-border">
          {isImage && previewUrl && (
            <div className="flex items-center justify-center bg-muted/30 p-4">
              <img
                src={previewUrl}
                alt={file.filename}
                className="max-w-full max-h-[55vh] object-contain"
              />
            </div>
          )}

          {isPdf && previewUrl && (
            <div className="w-full h-[55vh] bg-muted/30 overflow-hidden">
              <iframe
                src={previewUrl}
                className="w-full h-full"
                title={file.filename}
              />
            </div>
          )}

          {isVideo && previewUrl && (
            <div className="flex items-center justify-center bg-muted/30 p-4">
              <video
                src={previewUrl}
                controls
                className="max-w-full max-h-[55vh]"
              >
                brak obsługi wideo
              </video>
            </div>
          )}

          {isAudio && previewUrl && (
            <div className="flex flex-col items-center justify-center bg-muted/30 p-8 gap-4">
              {fileIcons[fileType]}
              <audio src={previewUrl} controls className="w-full max-w-md">
                brak obsługi audio
              </audio>
            </div>
          )}

          {!isImage && !isPdf && !isVideo && !isAudio && (
            <div className="flex flex-col items-center justify-center bg-muted/30 p-12 gap-4">
              {fileIcons[fileType]}
              <div className="text-center font-mono">
                <p className="text-sm">{file.filename}</p>
                <p className="text-xs text-muted-foreground">
                  {formatBytes(file.size)} • {file.content_type}
                </p>
                <p className="text-xs text-muted-foreground mt-2">
                  podgląd niedostępny
                </p>
              </div>
            </div>
          )}

          {(isImage || isPdf || isVideo || isAudio) && !previewUrl && (
            <div className="flex flex-col items-center justify-center bg-muted/30 p-12 gap-4">
              {fileIcons[fileType]}
              <div className="text-center font-mono">
                <p className="text-sm">{file.filename}</p>
                <p className="text-xs text-muted-foreground">
                  {formatBytes(file.size)}
                </p>
                <p className="text-xs text-muted-foreground mt-2">
                  pobierz aby wyświetlić
                </p>
              </div>
            </div>
          )}
        </div>

        <div className="flex justify-end pt-3 border-t border-border">
          <Button
            onClick={() => onDownload(file.id, file.filename)}
            className="text-xs"
          >
            <Download className="w-3.5 h-3.5 mr-1.5" />
            pobierz
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
