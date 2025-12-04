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
  image: <FileImage className="w-16 h-16 text-pink-500" />,
  video: <FileVideo className="w-16 h-16 text-purple-500" />,
  audio: <FileAudio className="w-16 h-16 text-green-500" />,
  pdf: <FileText className="w-16 h-16 text-red-500" />,
  archive: <Archive className="w-16 h-16 text-yellow-600" />,
  doc: <FileText className="w-16 h-16 text-blue-500" />,
  spreadsheet: <FileSpreadsheet className="w-16 h-16 text-emerald-500" />,
  text: <FileText className="w-16 h-16 text-gray-500" />,
  file: <File className="w-16 h-16 text-muted-foreground" />,
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
          <DialogTitle className="truncate pr-8">{file.filename}</DialogTitle>
        </DialogHeader>

        <div className="flex-1 overflow-auto min-h-0">
          {isImage && previewUrl && (
            <div className="flex items-center justify-center bg-muted/30 rounded-lg p-4">
              <img
                src={previewUrl}
                alt={file.filename}
                className="max-w-full max-h-[60vh] object-contain rounded"
              />
            </div>
          )}

          {isPdf && previewUrl && (
            <div className="w-full h-[60vh] bg-muted/30 rounded-lg overflow-hidden">
              <iframe
                src={previewUrl}
                className="w-full h-full"
                title={file.filename}
              />
            </div>
          )}

          {isVideo && previewUrl && (
            <div className="flex items-center justify-center bg-muted/30 rounded-lg p-4">
              <video
                src={previewUrl}
                controls
                className="max-w-full max-h-[60vh] rounded"
              >
                Twoja przeglądarka nie obsługuje odtwarzania wideo.
              </video>
            </div>
          )}

          {isAudio && previewUrl && (
            <div className="flex flex-col items-center justify-center bg-muted/30 rounded-lg p-8 gap-4">
              {fileIcons[fileType]}
              <audio src={previewUrl} controls className="w-full max-w-md">
                Twoja przeglądarka nie obsługuje odtwarzania audio.
              </audio>
            </div>
          )}

          {!isImage && !isPdf && !isVideo && !isAudio && (
            <div className="flex flex-col items-center justify-center bg-muted/30 rounded-lg p-12 gap-4">
              {fileIcons[fileType]}
              <div className="text-center">
                <p className="font-medium">{file.filename}</p>
                <p className="text-sm text-muted-foreground">
                  {formatBytes(file.size)} • {file.content_type}
                </p>
                <p className="text-sm text-muted-foreground mt-2">
                  Podgląd niedostępny dla tego typu pliku
                </p>
              </div>
            </div>
          )}

          {(isImage || isPdf || isVideo || isAudio) && !previewUrl && (
            <div className="flex flex-col items-center justify-center bg-muted/30 rounded-lg p-12 gap-4">
              {fileIcons[fileType]}
              <div className="text-center">
                <p className="font-medium">{file.filename}</p>
                <p className="text-sm text-muted-foreground">
                  {formatBytes(file.size)}
                </p>
                <p className="text-sm text-muted-foreground mt-2">
                  Pobierz plik, aby go wyświetlić
                </p>
              </div>
            </div>
          )}
        </div>

        <div className="flex justify-end pt-4 border-t">
          <Button onClick={() => onDownload(file.id, file.filename)}>
            <Download className="w-4 h-4 mr-2" />
            Pobierz
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
