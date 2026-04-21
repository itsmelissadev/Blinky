import { Key, Copy, Type, Hash, CheckCircle2, Calendar as CalendarIcon, Braces, Link2 } from "lucide-react";
import { format } from "date-fns";
import { toast } from "sonner";
import { ColumnDef } from "@tanstack/react-table";

import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Badge } from "@/components/ui/badge";
import { DataTableColumnHeader } from "@/components/ui/data-table";

const formatHeader = (str: string) => {
  return str.replace(/_/g, " ").replace(/\b\w/g, (l) => l.toUpperCase());
};

export interface CollectionField {
  id: string;
  name: string;
  type: string;
  props: any;
}

export const TypeIconMap: Record<string, any> = {
  text: Type,
  number: Hash,
  boolean: CheckCircle2,
  date: CalendarIcon,
  json: Braces,
  relation: Link2,
  id: Key,
};

export const renderIdCell = (idValue: string) => {
  return (
    <div className="flex items-center gap-2 group/id">
      <code className="relative rounded bg-muted px-2 py-[0.1rem] font-mono text-xs font-medium text-foreground/80">
        {idValue}
      </code>
      <Button
        variant="ghost"
        size="icon"
        className="size-5 opacity-0 group-hover/id:opacity-100 transition-opacity text-muted-foreground hover:text-foreground"
        onClick={() => {
          navigator.clipboard.writeText(idValue || "");
          toast.success("ID copied");
        }}
      >
        <Copy className="size-3" />
      </Button>
    </div>
  );
};

export function getBaseColumns<TData>(): ColumnDef<TData>[] {
  return [
    {
      id: "select",
      header: ({ table }) => (
        <div className="flex items-center justify-center">
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
        <div className="flex items-center justify-center">
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
  ];
}

export function generateCollectionColumns(schema: CollectionField[]): ColumnDef<any>[] {
  const columns: ColumnDef<any>[] = [...getBaseColumns()];

  // Always add ID if not present in schema
  const hasIdInSchema = schema.find((f) => f.name.toLowerCase() === "id");
  if (!hasIdInSchema) {
    columns.push({
      accessorKey: "id",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Id" icon={Key} />,
      cell: ({ row }) => renderIdCell(row.getValue("id")?.toString() || ""),
    });
  }

  // Map schema fields
  const schemaColumns = schema.map((field) => ({
    accessorKey: field.name,
    header: ({ column }: any) => (
      <DataTableColumnHeader
        column={column}
        title={formatHeader(field.name)}
        icon={field.name.toLowerCase() === "id" ? Key : TypeIconMap[field.type]}
      />
    ),
    cell: ({ row }: any) => {
      const value = row.getValue(field.name);
      if (value === null || value === undefined)
        return <span className="text-muted-foreground/50 text-xs italic">null</span>;

      // Handle ID field rendering
      if (field.name.toLowerCase() === "id") {
        return renderIdCell(value.toString());
      }

      // Handle Date fields
      if (field.type === "date") {
        try {
          const date = new Date(value);
          return (
            <div className="flex flex-col">
              <span className="text-sm font-medium tabular-nums">{format(date, "MMM d, yyyy")}</span>
              <span className="text-[10px] text-muted-foreground tabular-nums">{format(date, "HH:mm:ss")} UTC</span>
            </div>
          );
        } catch {
          return <span className="text-sm">{value.toString()}</span>;
        }
      }

      // Handle Boolean fields
      if (typeof value === "boolean" || field.type === "boolean") {
        const isTrue = value === true || value === "true";
        return (
          <Badge
            variant={isTrue ? "default" : "secondary"}
            className="h-5 px-1.5 text-[10px] font-bold uppercase tracking-tight"
          >
            {isTrue ? "True" : "False"}
          </Badge>
        );
      }

      // Handle Number fields
      if (typeof value === "number" || field.type === "number") {
        return (
          <span className="font-mono text-sm font-medium text-foreground tabular-nums">
            {Number(value).toLocaleString()}
          </span>
        );
      }

      // Handle Text/String fields
      if (field.type === "text" || typeof value === "string") {
        const str = value.toString();
        return (
          <div className="max-w-[250px] truncate text-sm" title={str}>
            {str}
          </div>
        );
      }

      // Handle Relation fields
      if (field.type === "relation") {
        const isArray = Array.isArray(value);
        if (!value) return <span className="text-muted-foreground text-xs">none</span>;

        const count = isArray ? value.length : 1;
        return (
          <Badge variant="outline" className="gap-1.5 font-mono text-[10px] h-5">
            <Link2 className="size-3 text-muted-foreground" />
            {isArray ? `${count} refs` : value.toString().substring(0, 8)}
          </Badge>
        );
      }

      // Handle JSON/Object fields
      if (typeof value === "object" || field.type === "json") {
        const isArray = Array.isArray(value);
        const count = isArray ? value.length : Object.keys(value || {}).length;
        return (
          <Badge variant="outline" className="gap-1.5 font-mono text-[10px] h-5">
            <Braces className="size-3 text-muted-foreground" />
            {isArray ? "ARRAY" : "OBJECT"} ({count})
          </Badge>
        );
      }

      return <span className="text-sm">{value.toString()}</span>;
    },
  }));

  return [...columns, ...schemaColumns];
}
