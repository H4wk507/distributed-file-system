import type { FileInfo } from "@/api/types";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { formatBytes, formatDate } from "@/lib/formatters";
import { Download, Info, MoreHorizontal, Trash2 } from "lucide-react";
import { useState } from "react";

interface FileActionsProps {
  file: FileInfo;
  onDelete: (fileId: string) => void;
  onDownload: (fileId: string, filename: string) => void;
  isDeleting?: boolean;
}

export function FileActions({
  file,
  onDelete,
  onDownload,
  isDeleting,
}: FileActionsProps) {
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [showInfoDialog, setShowInfoDialog] = useState(false);

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon" className="h-7 w-7">
            <MoreHorizontal className="h-3.5 w-3.5" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem
            onClick={() => onDownload(file.id, file.filename)}
            className="text-xs"
          >
            <Download className="mr-2 h-3.5 w-3.5" />
            pobierz
          </DropdownMenuItem>
          <DropdownMenuItem
            onClick={() => setShowInfoDialog(true)}
            className="text-xs"
          >
            <Info className="mr-2 h-3.5 w-3.5" />
            szczegóły
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem
            onClick={() => setShowDeleteDialog(true)}
            className="text-destructive focus:text-destructive text-xs"
            disabled={isDeleting}
          >
            <Trash2 className="mr-2 h-3.5 w-3.5" />
            usuń
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      {/* Delete Dialog */}
      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>usuń plik</AlertDialogTitle>
            <AlertDialogDescription className="font-mono">
              usunąć "{file.filename}"? operacja nieodwracalna.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel className="text-xs">anuluj</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => {
                onDelete(file.id);
                setShowDeleteDialog(false);
              }}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/80 text-xs"
            >
              usuń
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Info Dialog */}
      <Dialog open={showInfoDialog} onOpenChange={setShowInfoDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>szczegóły pliku</DialogTitle>
            <DialogDescription className="font-mono">
              informacje o pliku i replikach
            </DialogDescription>
          </DialogHeader>
          <div className="border border-border divide-y divide-border">
            <InfoRow label="nazwa" value={file.filename} />
            <InfoRow label="rozmiar" value={formatBytes(file.size)} />
            <InfoRow label="typ" value={file.content_type} />
            <InfoRow label="hash" value={file.hash} mono />
            <InfoRow label="utworzono" value={formatDate(file.created_at)} />
            <InfoRow
              label="zmodyfikowano"
              value={formatDate(file.updated_at)}
            />
            <div className="p-3">
              <span className="text-xs font-mono uppercase text-muted-foreground">
                repliki
              </span>
              <div className="flex flex-wrap gap-1 mt-1.5">
                {file.replicas?.map((nodeId, i) => (
                  <span
                    key={i}
                    className="px-2 py-0.5 text-xs bg-muted border border-border font-mono"
                  >
                    {nodeId}
                  </span>
                )) || (
                  <span className="text-xs text-muted-foreground font-mono">
                    brak
                  </span>
                )}
              </div>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}

function InfoRow({
  label,
  value,
  mono,
}: {
  label: string;
  value: string;
  mono?: boolean;
}) {
  return (
    <div className="flex justify-between items-start gap-4 p-3">
      <span className="text-xs font-mono uppercase text-muted-foreground shrink-0">
        {label}
      </span>
      <span
        className={`text-xs text-right break-all ${mono ? "font-mono" : ""}`}
      >
        {value}
      </span>
    </div>
  );
}
