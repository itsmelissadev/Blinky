import { ColumnDef } from "@tanstack/react-table";
import { Link } from "react-router-dom";
import { Layers, HardDrive, FileText, Settings, Eye, Calendar } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { DataTableColumnHeader } from "@/components/ui/data-table";

interface SubProps {
  is_nullable?: boolean;
  is_unique?: boolean;
  is_hidden?: boolean;
  default?: unknown;
  default_bool?: boolean;
  default_now?: boolean;
  text?: { min?: number; max?: number; regex?: string };
  number?: { min?: number; max?: number };
}

interface CollectionField {
  name: string;
  type: any;
  props: SubProps;
}

export interface Collection {
  name: string;
  schema: CollectionField[];
  total_documents: number;
  total_bytes: number;
  created_at: string;
  updated_at: string;
}

function formatBytes(bytes: number, decimals = 2) {
  if (bytes === 0) return "0 Bytes";
  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ["Bytes", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + " " + sizes[i];
}

export const getCollectionColumns = (onEdit: (col: Collection) => void): ColumnDef<Collection>[] => [
  {
    id: "select",
    header: ({ table }) => (
      <div className="flex items-center justify-center w-5">
        <Checkbox
          checked={
            (table.getIsAllPageRowsSelected()
              ? true
              : table.getIsSomePageRowsSelected()
                ? "indeterminate"
                : false) as any
          }
          onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
          aria-label="Select all"
        />
      </div>
    ),
    cell: ({ row }) => (
      <div className="flex items-center justify-center w-5">
        <Checkbox
          checked={row.getIsSelected()}
          onCheckedChange={(value) => row.toggleSelected(!!value)}
          aria-label="Select row"
        />
      </div>
    ),
    enableSorting: false,
    enableHiding: false,
  },
  {
    accessorKey: "name",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Name" />,
    cell: ({ row }) => (
      <Link to={`/collections/${row.original.name}`} className="flex items-center py-1 group">
        <span className="font-semibold text-base text-foreground tracking-tight group-hover:text-primary transition-colors">
          {row.original.name}
        </span>
      </Link>
    ),
  },
  {
    accessorKey: "schema",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Schema" icon={Layers} />,
    cell: ({ row }) => (
      <div className="flex items-center gap-2">
        <Layers className="h-3.5 w-3.5 text-muted-foreground/70" />
        <span className="text-sm font-medium text-foreground">{row.original.schema?.length || 0} Fields</span>
      </div>
    ),
  },
  {
    accessorKey: "total_documents",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Total Documents" icon={HardDrive} />,
    cell: ({ row }) => (
      <div className="flex items-center gap-2">
        <HardDrive className="h-3.5 w-3.5 text-muted-foreground/70" />
        <span className="text-sm font-medium text-foreground">
          {(row.original.total_documents || 0).toLocaleString()}
        </span>
      </div>
    ),
  },
  {
    accessorKey: "total_bytes",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Total Bytes" icon={FileText} />,
    cell: ({ row }) => (
      <div className="flex items-center gap-2">
        <FileText className="h-3.5 w-3.5 text-muted-foreground/70" />
        <span className="text-sm font-medium text-foreground">{formatBytes(row.original.total_bytes || 0)}</span>
      </div>
    ),
  },
  {
    accessorKey: "updated_at",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Updated At" icon={Calendar} />,
    cell: ({ row }) => {
      const date = new Date(row.original.updated_at);
      return (
        <div className="flex flex-col">
          <span className="text-sm font-medium text-foreground">
            {date.toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric", timeZone: "UTC" })}
          </span>
          <span className="text-[10px] text-muted-foreground font-medium">
            {date.toLocaleTimeString("en-US", {
              hour: "2-digit",
              minute: "2-digit",
              second: "2-digit",
              hour12: false,
              timeZone: "UTC",
            })}{" "}
            UTC
          </span>
        </div>
      );
    },
  },
  {
    accessorKey: "created_at",
    header: ({ column }) => <DataTableColumnHeader column={column} title="Created At" icon={Calendar} />,
    cell: ({ row }) => {
      const date = new Date(row.original.created_at);
      return (
        <div className="flex flex-col">
          <span className="text-sm font-medium text-foreground">
            {date.toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric", timeZone: "UTC" })}
          </span>
          <span className="text-xs text-muted-foreground/70 font-medium">
            {date.toLocaleTimeString("en-US", {
              hour: "2-digit",
              minute: "2-digit",
              second: "2-digit",
              hour12: false,
              timeZone: "UTC",
            })}{" "}
            UTC
          </span>
        </div>
      );
    },
  },
  {
    id: "actions",
    header: () => <div className="text-right">Actions</div>,
    cell: ({ row }) => (
      <div className="flex items-center justify-end gap-2">
        <Button
          variant="outline"
          size="icon"
          onClick={() => onEdit(row.original)}
          className="h-8 w-8 hover:bg-primary/10 hover:border-primary/30 hover:text-primary transition-all shadow-sm"
          title="Edit Schema"
        >
          <Settings className="h-3.5 w-3.5" />
        </Button>
        <Link to={`/collections/${row.original.name}`}>
          <Button
            variant="outline"
            size="icon"
            className="h-8 w-8 hover:bg-primary/10 hover:border-primary/30 hover:text-primary transition-all shadow-sm group"
            title="View Records"
          >
            <Eye className="h-3.5 w-3.5 group-hover:scale-110 transition-transform" />
          </Button>
        </Link>
      </div>
    ),
  },
];
