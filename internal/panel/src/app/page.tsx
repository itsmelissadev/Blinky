import { useState, ReactNode } from "react";
import { Database, Plus, LayoutGrid, RefreshCw, Activity } from "lucide-react";

import { Button } from "@/components/ui/button";
import { CollectionDialog } from "@/components/collections-dialog";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { DataTable } from "@/components/ui/data-table";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

export default function Home() {
  const [isNewDialogOpen, setIsNewDialogOpen] = useState(false);

  return (
    <div className="flex flex-col gap-6">
      <CollectionDialog open={isNewDialogOpen} onOpenChange={setIsNewDialogOpen} onSuccess={() => {}} />

      <div className="flex items-center justify-between">
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight">Dashboard</h1>
          <p className="text-sm text-muted-foreground">Overview of your Blinky engine instance</p>
        </div>
        <Button onClick={() => setIsNewDialogOpen(true)}>
          <Plus className="mr-2 size-4" />
          New Collection
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <MetricCard title="Collections" value="12" icon={<Database className="size-4 text-indigo-500" />} />
        <MetricCard title="Total Records" value="45,231" icon={<LayoutGrid className="size-4 text-emerald-500" />} />
        <MetricCard
          title="System Load"
          value="14%"
          icon={<Activity className="size-4 text-amber-500" />}
          description="Normal operation"
        />
      </div>

      <div className="grid gap-4">
        <Card className="col-span-full">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
            <div className="space-y-1">
              <CardTitle>Recent Activity</CardTitle>
              <CardDescription>Latest operations executed on the engine</CardDescription>
            </div>
            <Button variant="ghost" size="icon" className="size-8">
              <RefreshCw className="size-4 text-muted-foreground" />
            </Button>
          </CardHeader>
          <CardContent>
            <DataTable
              data={[
                { op: "Insert Record", col: "orders", time: "2m ago", lat: "12ms", variant: "default" as const },
                { op: "Update Schema", col: "products", time: "5m ago", lat: "45ms", variant: "outline" as const },
                { op: "SQL Query", col: "users", time: "12m ago", lat: "5ms", variant: "secondary" as const },
                {
                  op: "Engine Migration",
                  col: "categories",
                  time: "20m ago",
                  lat: "210ms",
                  variant: "default" as const,
                },
                { op: "Delete Record", col: "logs", time: "45m ago", lat: "8ms", variant: "destructive" as const },
              ]}
              columns={[
                {
                  accessorKey: "op",
                  header: "Operation",
                  cell: ({ row }) => {
                    const variant = row.original.variant;
                    return (
                      <div className="flex items-center gap-2 font-medium">
                        <div
                          className={cn(
                            "size-2 rounded-full",
                            variant === "default" && "bg-indigo-500",
                            variant === "outline" && "bg-zinc-500",
                            variant === "secondary" && "bg-emerald-500",
                            variant === "destructive" && "bg-rose-500",
                          )}
                        />
                        {row.original.op}
                      </div>
                    );
                  },
                },
                {
                  accessorKey: "col",
                  header: "Collection",
                  cell: ({ row }) => (
                    <Badge variant="outline" className="font-mono text-[10px] bg-muted/30">
                      {row.original.col}
                    </Badge>
                  ),
                },
                {
                  accessorKey: "time",
                  header: "Time",
                  cell: ({ row }) => <span className="text-xs text-muted-foreground">{row.original.time}</span>,
                },
                {
                  accessorKey: "lat",
                  header: () => <div className="text-right">Execution</div>,
                  cell: ({ row }) => <div className="text-right font-mono text-xs">{row.original.lat}</div>,
                },
              ]}
            />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

function MetricCard({
  title,
  value,
  icon,
  description,
}: {
  title: string;
  value: string;
  icon: React.ReactNode;
  description?: string;
}) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        {icon}
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold tabular-nums">{value}</div>
        {description && <p className="text-xs text-muted-foreground mt-1">{description}</p>}
      </CardContent>
    </Card>
  );
}
