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
          <Button variant="ghost" size="icon" className="h-8 w-8">
            <MoreHorizontal className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem onClick={() => onDownload(file.id, file.filename)}>
            <Download className="mr-2 h-4 w-4" />
            Pobierz
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => setShowInfoDialog(true)}>
            <Info className="mr-2 h-4 w-4" />
            Szczegóły
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem
            onClick={() => setShowDeleteDialog(true)}
            className="text-destructive focus:text-destructive"
            disabled={isDeleting}
          >
            <Trash2 className="mr-2 h-4 w-4" />
            Usuń
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Usuń plik</AlertDialogTitle>
            <AlertDialogDescription>
              Czy na pewno chcesz usunąć plik "{file.filename}"? Ta operacja
              jest nieodwracalna i usunie wszystkie repliki pliku.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Anuluj</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => {
                onDelete(file.id);
                setShowDeleteDialog(false);
              }}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              Usuń
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
      <Dialog open={showInfoDialog} onOpenChange={setShowInfoDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Szczegóły pliku</DialogTitle>
            <DialogDescription>
              Informacje o pliku i jego replikach
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <InfoRow label="Nazwa" value={file.filename} />
            <InfoRow label="Rozmiar" value={formatBytes(file.size)} />
            <InfoRow label="Typ" value={file.content_type} />
            <InfoRow label="Hash" value={file.hash} mono />
            <InfoRow label="Utworzono" value={formatDate(file.created_at)} />
            <InfoRow
              label="Zmodyfikowano"
              value={formatDate(file.updated_at)}
            />
            <div className="space-y-1">
              <span className="text-sm font-medium">Repliki</span>
              <div className="flex flex-wrap gap-2">
                {file.replicas?.map((nodeId, i) => (
                  <span
                    key={i}
                    className="px-2 py-1 text-xs bg-muted rounded-md font-mono"
                  >
                    {nodeId}
                  </span>
                )) || (
                  <span className="text-sm text-muted-foreground">Brak</span>
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
    <div className="flex justify-between items-start gap-4">
      <span className="text-sm font-medium shrink-0">{label}</span>
      <span
        className={`text-sm text-muted-foreground text-right break-all ${
          mono ? "font-mono text-xs" : ""
        }`}
      >
        {value}
      </span>
    </div>
  );
}
