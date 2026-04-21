"use client";
"use no memo";

import * as React from "react";
import {
  ColumnDef,
  ColumnFiltersState,
  SortingState,
  VisibilityState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table";
import { ChevronDown } from "lucide-react";

import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
  DropdownMenuGroup,
} from "@/components/ui/dropdown-menu";
import { cn } from "@/lib/utils";
import { ArrowDown, ArrowUp, ArrowUpDown } from "lucide-react";

interface DataTableColumnHeaderProps<TData, TValue> extends React.HTMLAttributes<HTMLDivElement> {
  column: any;
  title: string;
  icon?: React.ComponentType<{ className?: string }>;
}

export function DataTableColumnHeader<TData, TValue>({
  column,
  title,
  icon: Icon,
  className,
}: DataTableColumnHeaderProps<TData, TValue>) {
  if (!column.getCanSort()) {
    return (
      <div className={cn("text-sm font-semibold tracking-tight text-foreground", className)}>
        <div className="flex items-center gap-2">
          {Icon && <Icon className="size-3.5 opacity-80" />}
          {title}
        </div>
      </div>
    );
  }

  const isSorted = column.getIsSorted();
  const SortIcon = isSorted === "asc" ? ArrowUp : isSorted === "desc" ? ArrowDown : ArrowUpDown;

  return (
    <div className={cn("flex items-center space-x-2", className)}>
      <Button
        variant="ghost"
        size="sm"
        className="-ml-3 h-8 gap-2 text-sm font-semibold tracking-tight text-foreground hover:bg-transparent cursor-default"
        onClick={() => column.toggleSorting(isSorted === "asc")}
      >
        {Icon && <Icon className="size-3.5 opacity-80" />}
        <span>{title}</span>
        {isSorted && <SortIcon className="ml-1 size-3.5 text-foreground transition-none" />}
      </Button>
    </div>
  );
}

interface DataTableProps<TData, TValue> {
  columns: ColumnDef<TData, TValue>[];
  data: TData[];
  searchKey?: string;
  onRowSelectionChange?: (selectedRows: TData[]) => void;
  onRowClick?: (row: TData) => void;
}

export function DataTable<TData, TValue>({
  columns,
  data,
  searchKey,
  onRowSelectionChange,
  onRowClick,
}: DataTableProps<TData, TValue>) {
  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>([]);
  const [columnVisibility, setColumnVisibility] = React.useState<VisibilityState>({});
  const [rowSelection, setRowSelection] = React.useState({});

  // eslint-disable-next-line react-hooks/incompatible-library
  const table = useReactTable({
    data,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: (updater) => {
      const nextSelection = typeof updater === "function" ? updater(rowSelection) : updater;
      setRowSelection(nextSelection);
    },
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      rowSelection,
    },
  });

  // Sync selection to parent
  React.useEffect(() => {
    if (onRowSelectionChange) {
      const selected = table.getFilteredSelectedRowModel().rows.map((row) => row.original);
      onRowSelectionChange(selected);
    }
  }, [rowSelection, table, onRowSelectionChange]);

  return (
    <div className="w-full">
      <div className="flex items-center py-4 gap-2">
        {searchKey && table.getColumn(searchKey) && (
          <Input
            placeholder={`Search by ${searchKey}...`}
            value={(table.getColumn(searchKey)?.getFilterValue() as string) ?? ""}
            onChange={(event) => table.getColumn(searchKey)?.setFilterValue(event.target.value)}
            className="max-w-sm h-9"
          />
        )}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" className="ml-auto h-9">
              Columns <ChevronDown className="ml-2 h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuGroup>
              {table
                .getAllColumns()
                .filter((column) => column.getCanHide())
                .map((column) => {
                  return (
                    <DropdownMenuCheckboxItem
                      key={column.id}
                      className="capitalize"
                      checked={column.getIsVisible()}
                      onCheckedChange={(value) => column.toggleVisibility(!!value)}
                    >
                      {typeof column.columnDef.header === "string"
                        ? column.columnDef.header
                        : column.id.replace(/_/g, " ")}
                    </DropdownMenuCheckboxItem>
                  );
                })}
            </DropdownMenuGroup>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
      <div className="rounded-md border">
        {/* Desktop View */}
        <div className="hidden md:block">
          <Table>
            <TableHeader>
              {table.getHeaderGroups().map((headerGroup) => (
                <TableRow key={headerGroup.id}>
                  {headerGroup.headers.map((header) => (
                    <TableHead
                      key={header.id}
                      className={cn(
                        "whitespace-nowrap h-10 text-sm font-semibold text-foreground",
                        header.id === "select" && "w-12",
                      )}
                    >
                      {header.id === "select" ? (
                        <div className="flex items-center justify-center">
                          {flexRender(header.column.columnDef.header, header.getContext())}
                        </div>
                      ) : header.isPlaceholder ? null : (
                        flexRender(header.column.columnDef.header, header.getContext())
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
                    className={cn(onRowClick && "cursor-pointer")}
                    onClick={(e) => {
                      const target = e.target as HTMLElement;
                      if (target.closest("button") || target.closest("a") || target.closest('input[type="checkbox"]'))
                        return;
                      onRowClick?.(row.original);
                    }}
                  >
                    {row.getVisibleCells().map((cell) => (
                      <TableCell
                        key={cell.id}
                        className={cn("whitespace-nowrap", cell.column.id === "select" && "w-12")}
                      >
                        {cell.column.id === "select" ? (
                          <div className="flex items-center justify-center">
                            {flexRender(cell.column.columnDef.cell, cell.getContext())}
                          </div>
                        ) : (
                          flexRender(cell.column.columnDef.cell, cell.getContext())
                        )}
                      </TableCell>
                    ))}
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell colSpan={columns.length} className="h-24 text-center text-muted-foreground">
                    No results.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>

        {/* Mobile View (Simplified Vertical View) */}
        <div className="md:hidden divide-y">
          {table.getRowModel().rows?.length ? (
            table.getRowModel().rows.map((row) => (
              <div
                key={row.id}
                className={cn("p-4 flex flex-col gap-4", onRowClick && "cursor-pointer active:bg-muted/50")}
                onClick={(e) => {
                  const target = e.target as HTMLElement;
                  if (target.closest("button") || target.closest("a") || target.closest('input[type="checkbox"]'))
                    return;
                  onRowClick?.(row.original);
                }}
              >
                {/* Header Row (Select & Actions) */}
                <div className="flex items-center justify-between">
                  {(() => {
                    const selectCell = row.getVisibleCells().find((c) => c.column.id === "select");
                    const actionsCell = row.getVisibleCells().find((c) => c.column.id === "actions");

                    return (
                      <>
                        <div className="flex items-center gap-3">
                          {selectCell && (
                            <div className="flex items-center justify-center shrink-0">
                              {flexRender(selectCell.column.columnDef.cell, selectCell.getContext())}
                            </div>
                          )}
                          <span className="text-xs font-bold text-muted-foreground">Row #{row.id}</span>
                        </div>
                        {actionsCell && (
                          <div className="flex items-center gap-2 px-1">
                            {flexRender(actionsCell.column.columnDef.cell, actionsCell.getContext())}
                          </div>
                        )}
                      </>
                    );
                  })()}
                </div>

                {/* Content Grid */}
                <div className="grid gap-4">
                  {row
                    .getVisibleCells()
                    .filter((cell) => cell.column.id !== "select" && cell.column.id !== "actions")
                    .map((cell) => (
                      <div key={cell.id} className="flex flex-col gap-1 px-1">
                        <span className="text-[11px] font-bold text-muted-foreground/60 uppercase">
                          {typeof cell.column.columnDef.header === "string"
                            ? cell.column.columnDef.header
                            : cell.column.id.replace(/_/g, " ")}
                        </span>
                        <div className="text-sm font-medium">
                          {flexRender(cell.column.columnDef.cell, cell.getContext())}
                        </div>
                      </div>
                    ))}
                </div>
              </div>
            ))
          ) : (
            <div className="p-8 text-center text-muted-foreground text-sm">No results found.</div>
          )}
        </div>
      </div>
      <div className="flex items-center justify-end space-x-2 py-4">
        <div className="flex-1 text-sm text-muted-foreground">
          {table.getFilteredSelectedRowModel().rows.length} of {table.getFilteredRowModel().rows.length} row(s)
          selected.
        </div>
        <div className="space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.previousPage()}
            disabled={!table.getCanPreviousPage()}
            className="h-8 rounded-lg px-3"
          >
            Previous
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.nextPage()}
            disabled={!table.getCanNextPage()}
            className="h-8 rounded-lg px-3"
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  );
}
