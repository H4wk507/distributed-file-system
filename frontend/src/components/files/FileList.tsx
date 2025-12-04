import type { FileInfo } from "@/api/types";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { getFileIcon, type FileType } from "@/lib/file";
import { formatBytes, formatDate } from "@/lib/formatters";
import {
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  useReactTable,
  type ColumnDef,
  type SortingState,
} from "@tanstack/react-table";
import {
  Archive,
  ArrowUpDown,
  ChevronLeft,
  ChevronRight,
  File,
  FileAudio,
  FileImage,
  FileSpreadsheet,
  FileText,
  FileVideo,
  Search,
  Trash2,
} from "lucide-react";
import { useMemo, useState } from "react";
import { FileActions } from "./FileActions";
import { FilePreview } from "./FilePreview";

const fileIcons: Record<FileType, React.ReactNode> = {
  image: <FileImage className="w-5 h-5 text-pink-500" />,
  video: <FileVideo className="w-5 h-5 text-purple-500" />,
  audio: <FileAudio className="w-5 h-5 text-green-500" />,
  pdf: <FileText className="w-5 h-5 text-red-500" />,
  archive: <Archive className="w-5 h-5 text-yellow-600" />,
  doc: <FileText className="w-5 h-5 text-blue-500" />,
  spreadsheet: <FileSpreadsheet className="w-5 h-5 text-emerald-500" />,
  text: <FileText className="w-5 h-5 text-gray-500" />,
  file: <File className="w-5 h-5 text-muted-foreground" />,
};

interface FileListProps {
  files: FileInfo[];
  page: number;
  totalPages: number;
  onPageChange: (page: number) => void;
  onDelete: (fileId: string) => void;
  onDownload: (fileId: string, filename: string) => void;
  isDeleting?: boolean;
}

export function FileList({
  files,
  page,
  totalPages,
  onPageChange,
  onDelete,
  onDownload,
  isDeleting,
}: FileListProps) {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [globalFilter, setGlobalFilter] = useState("");
  const [rowSelection, setRowSelection] = useState({});
  const [previewFile, setPreviewFile] = useState<FileInfo | null>(null);

  const columns: ColumnDef<FileInfo>[] = useMemo(
    () => [
      {
        id: "select",
        header: ({ table }) => (
          <Checkbox
            checked={
              table.getIsAllPageRowsSelected() ||
              (table.getIsSomePageRowsSelected() && "indeterminate")
            }
            onCheckedChange={(value) =>
              table.toggleAllPageRowsSelected(!!value)
            }
            aria-label="Zaznacz wszystkie"
          />
        ),
        cell: ({ row }) => (
          <Checkbox
            checked={row.getIsSelected()}
            onCheckedChange={(value) => row.toggleSelected(!!value)}
            aria-label="Zaznacz wiersz"
          />
        ),
        enableSorting: false,
      },
      {
        accessorKey: "filename",
        header: ({ column }) => (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
            className="-ml-3"
          >
            Nazwa
            <ArrowUpDown className="ml-2 h-4 w-4" />
          </Button>
        ),
        cell: ({ row }) => {
          const file = row.original;
          const iconType = getFileIcon(file.content_type);
          return (
            <button
              onClick={() => setPreviewFile(file)}
              className="flex items-center gap-3 hover:underline text-left"
            >
              {fileIcons[iconType]}
              <span className="font-medium truncate max-w-[300px]">
                {file.filename}
              </span>
            </button>
          );
        },
      },
      {
        accessorKey: "size",
        header: ({ column }) => (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
            className="-ml-3"
          >
            Rozmiar
            <ArrowUpDown className="ml-2 h-4 w-4" />
          </Button>
        ),
        cell: ({ row }) => formatBytes(row.original.size),
      },
      {
        accessorKey: "created_at",
        header: ({ column }) => (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
            className="-ml-3"
          >
            Data
            <ArrowUpDown className="ml-2 h-4 w-4" />
          </Button>
        ),
        cell: ({ row }) => formatDate(row.original.created_at),
      },
      {
        accessorKey: "replicas",
        header: "Repliki",
        cell: ({ row }) => (
          <span className="text-muted-foreground">
            {row.original.replicas?.length || 0}
          </span>
        ),
        enableSorting: false,
      },
      {
        id: "actions",
        cell: ({ row }) => {
          const file = row.original;
          return (
            <FileActions
              file={file}
              onDelete={onDelete}
              onDownload={onDownload}
              isDeleting={isDeleting}
            />
          );
        },
      },
    ],
    [onDelete, onDownload, isDeleting],
  );

  const table = useReactTable({
    data: files,
    columns,
    state: {
      sorting,
      globalFilter,
      rowSelection,
    },
    onSortingChange: setSorting,
    onGlobalFilterChange: setGlobalFilter,
    onRowSelectionChange: setRowSelection,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
  });

  const selectedCount = Object.keys(rowSelection).length;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between gap-4">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Szukaj plików..."
            value={globalFilter}
            onChange={(e) => setGlobalFilter(e.target.value)}
            className="pl-9"
          />
        </div>
        {selectedCount > 0 && (
          <div className="flex items-center gap-2">
            <span className="text-sm text-muted-foreground">
              Zaznaczono: {selectedCount}
            </span>
            <Button
              variant="destructive"
              size="sm"
              onClick={() => {
                const selectedIds = table
                  .getSelectedRowModel()
                  .rows.map((row) => row.original.id);
                selectedIds.forEach((id) => onDelete(id));
                setRowSelection({});
              }}
              disabled={isDeleting}
            >
              <Trash2 className="mr-2 h-4 w-4" />
              Usuń zaznaczone
            </Button>
          </div>
        )}
      </div>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(
                          header.column.columnDef.header,
                          header.getContext(),
                        )}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && "selected"}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext(),
                      )}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell
                  colSpan={columns.length}
                  className="h-24 text-center"
                >
                  Brak plików
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            Strona {page} z {totalPages}
          </p>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => onPageChange(page - 1)}
              disabled={page <= 1}
            >
              <ChevronLeft className="h-4 w-4" />
              Poprzednia
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => onPageChange(page + 1)}
              disabled={page >= totalPages}
            >
              Następna
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}
      <FilePreview
        file={previewFile}
        open={!!previewFile}
        onOpenChange={(open) => !open && setPreviewFile(null)}
        onDownload={onDownload}
      />
    </div>
  );
}
