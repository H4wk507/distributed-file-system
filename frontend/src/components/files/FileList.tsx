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
  image: <FileImage className="w-4 h-4 text-primary" />,
  video: <FileVideo className="w-4 h-4 text-primary" />,
  audio: <FileAudio className="w-4 h-4 text-success" />,
  pdf: <FileText className="w-4 h-4 text-destructive" />,
  archive: <Archive className="w-4 h-4 text-primary" />,
  doc: <FileText className="w-4 h-4 text-muted-foreground" />,
  spreadsheet: <FileSpreadsheet className="w-4 h-4 text-success" />,
  text: <FileText className="w-4 h-4 text-muted-foreground" />,
  file: <File className="w-4 h-4 text-muted-foreground" />,
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
            aria-label="zaznacz wszystkie"
          />
        ),
        cell: ({ row }) => (
          <Checkbox
            checked={row.getIsSelected()}
            onCheckedChange={(value) => row.toggleSelected(!!value)}
            aria-label="zaznacz"
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
            className="-ml-3 text-xs"
          >
            nazwa
            <ArrowUpDown className="ml-1 h-3 w-3" />
          </Button>
        ),
        cell: ({ row }) => {
          const file = row.original;
          const iconType = getFileIcon(file.content_type);
          return (
            <button
              onClick={() => setPreviewFile(file)}
              className="flex items-center gap-2 hover:text-primary text-left"
            >
              {fileIcons[iconType]}
              <span className="font-mono text-xs truncate max-w-[240px]">
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
            className="-ml-3 text-xs"
          >
            rozmiar
            <ArrowUpDown className="ml-1 h-3 w-3" />
          </Button>
        ),
        cell: ({ row }) => (
          <span className="font-mono text-xs text-muted-foreground">
            {formatBytes(row.original.size)}
          </span>
        ),
      },
      {
        accessorKey: "created_at",
        header: ({ column }) => (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
            className="-ml-3 text-xs"
          >
            data
            <ArrowUpDown className="ml-1 h-3 w-3" />
          </Button>
        ),
        cell: ({ row }) => (
          <span className="font-mono text-xs text-muted-foreground">
            {formatDate(row.original.created_at)}
          </span>
        ),
      },
      {
        accessorKey: "replicas",
        header: () => <span className="text-xs">repliki</span>,
        cell: ({ row }) => (
          <span className="font-mono text-xs text-muted-foreground">
            {row.original.replicas_count}
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
    <div>
      {/* Toolbar */}
      <div className="flex items-center justify-between gap-3 p-3 border-b border-border bg-muted/30">
        <div className="relative flex-1 max-w-xs">
          <Search className="absolute left-2 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
          <Input
            placeholder="szukaj..."
            value={globalFilter}
            onChange={(e) => setGlobalFilter(e.target.value)}
            className="pl-7 h-7 text-xs"
          />
        </div>
        {selectedCount > 0 && (
          <div className="flex items-center gap-2">
            <span className="text-xs text-muted-foreground font-mono">
              [{selectedCount}]
            </span>
            <Button
              variant="destructive"
              size="sm"
              className="h-7 text-xs"
              onClick={() => {
                const selectedIds = table
                  .getSelectedRowModel()
                  .rows.map((row) => row.original.id);
                selectedIds.forEach((id) => onDelete(id));
                setRowSelection({});
              }}
              disabled={isDeleting}
            >
              <Trash2 className="mr-1 h-3 w-3" />
              usuń
            </Button>
          </div>
        )}
      </div>

      {/* Table */}
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
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </TableCell>
                ))}
              </TableRow>
            ))
          ) : (
            <TableRow>
              <TableCell
                colSpan={columns.length}
                className="h-20 text-center text-xs text-muted-foreground font-mono"
              >
                brak plików
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between p-3 border-t border-border bg-muted/30">
          <p className="text-xs text-muted-foreground font-mono">
            strona {page}/{totalPages}
          </p>
          <div className="flex items-center gap-1">
            <Button
              variant="outline"
              size="sm"
              className="h-7 text-xs"
              onClick={() => onPageChange(page - 1)}
              disabled={page <= 1}
            >
              <ChevronLeft className="h-3 w-3 mr-1" />
              prev
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="h-7 text-xs"
              onClick={() => onPageChange(page + 1)}
              disabled={page >= totalPages}
            >
              next
              <ChevronRight className="h-3 w-3 ml-1" />
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
