import { useEffect, useState, useMemo } from "react";
import { fetchAPI } from "@/lib/api-client";
import { DataTable } from "@/components/ui/data-table";
import { PageLoader } from "@/components/global/widget/loader";
import { Plus, RefreshCw, Trash2, Edit, Calendar } from "lucide-react";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { DataTableColumnHeader } from "@/components/ui/data-table";
import { Checkbox } from "@/components/ui/checkbox";
import { cn } from "@/lib/utils";
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

export default function AdminsPage() {
  const [admins, setAdmins] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [adminToDelete, setAdminToDelete] = useState<string | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const fetchAdmins = async () => {
    setLoading(true);
    try {
      const res = await fetchAPI("/admins");
      if (res.success) setAdmins(res.data || []);
    } catch (e) {
      toast.error("Failed to fetch admins");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAdmins();
  }, []);

  const handleDelete = (id: string) => {
    setAdminToDelete(id);
    setDeleteDialogOpen(true);
  };

  const confirmDelete = async () => {
    if (!adminToDelete) return;

    setIsDeleting(true);
    try {
      const res = await fetchAPI("/admins/user", {
        method: "DELETE",
        body: JSON.stringify({ id: adminToDelete }),
      });
      if (res.success) {
        toast.success("Admin deleted successfully");
        fetchAdmins();
      } else {
        toast.error(res.error?.message || "Failed to delete admin");
      }
    } catch (e) {
      toast.error("An unexpected error occurred while deleting");
    } finally {
      setIsDeleting(false);
      setAdminToDelete(null);
      setDeleteDialogOpen(false);
    }
  };

  const columns = useMemo(
    () => [
      {
        id: "select",
        header: ({ table }: any) => (
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
        cell: ({ row }: any) => (
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
        accessorKey: "nickname",
        header: ({ column }: any) => <DataTableColumnHeader column={column} title="Administrator" />,
        cell: ({ row }: any) => {
          const initials = row.original.nickname?.substring(0, 2).toUpperCase() || "AD";
          return (
            <div className="flex items-center gap-3 py-1 group">
              <Avatar className="h-9 w-9 rounded-full shadow-sm border border-border/50">
                <AvatarImage src={row.original.avatar} alt={row.original.nickname} />
                <AvatarFallback className="rounded-full bg-primary/5 text-primary font-bold text-xs">
                  {initials}
                </AvatarFallback>
              </Avatar>
              <div className="flex flex-col">
                <span className="font-semibold text-base text-foreground tracking-tight group-hover:text-primary transition-colors">
                  {row.original.nickname}
                </span>
                <span className="text-xs text-muted-foreground font-medium">{row.original.email}</span>
              </div>
            </div>
          );
        },
      },
      {
        accessorKey: "username",
        header: ({ column }: any) => <DataTableColumnHeader column={column} title="Username" />,
        cell: ({ row }: any) => (
          <div className="flex items-center gap-1">
            <span className="font-mono text-sm font-medium text-foreground/90 group-hover:text-primary transition-colors">
              @{row.original.username}
            </span>
          </div>
        ),
      },
      {
        accessorKey: "created_at",
        header: ({ column }: any) => <DataTableColumnHeader column={column} title="Created At" icon={Calendar} />,
        cell: ({ row }: any) => {
          const date = new Date(row.original.created_at);
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
        accessorKey: "updated_at",
        header: ({ column }: any) => <DataTableColumnHeader column={column} title="Updated At" icon={Calendar} />,
        cell: ({ row }: any) => {
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
        id: "actions",
        header: () => <div className="text-right">Actions</div>,
        cell: ({ row }: any) => {
          return (
            <div className="flex justify-end gap-2">
              <Button
                variant="outline"
                size="icon"
                className="h-8 w-8 hover:bg-primary/10 hover:border-primary/30 hover:text-primary transition-all shadow-sm"
                onClick={() => toast.info("Edit Admin coming soon!")}
              >
                <Edit className="h-3.5 w-3.5" />
              </Button>
              <Button
                variant="outline"
                size="icon"
                className="h-8 w-8 hover:bg-destructive/10 hover:border-destructive/30 hover:text-destructive transition-all shadow-sm"
                onClick={() => handleDelete(row.original.id)}
              >
                <Trash2 className="h-3.5 w-3.5" />
              </Button>
            </div>
          );
        },
      },
    ],
    [],
  );

  if (loading && admins.length === 0) {
    return <PageLoader title="Loading Administrators" description="Fetching system superusers..." />;
  }

  return (
    <div className="flex flex-col gap-8 animate-in fade-in slide-in-from-bottom-2 duration-500">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div className="flex flex-col gap-1">
          <h1 className="text-2xl font-bold tracking-tight">Admins</h1>
          <p className="text-sm text-muted-foreground">
            Manage superuser accounts that have full access to the Blinky infrastructure.
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-3">
          <Button variant="outline" onClick={fetchAdmins} disabled={loading} size="lg" className="flex-1 sm:flex-none">
            <RefreshCw className={cn("mr-2 h-4 w-4", loading && "animate-spin")} />
            Sync
          </Button>
          <Button size="lg" className="w-full sm:w-auto" onClick={() => toast.info("Adding admins coming soon!")}>
            <Plus className="mr-2 h-4 w-4" />
            New Admin
          </Button>
        </div>
      </div>

      <DataTable data={admins} columns={columns} searchKey="nickname" />

      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Admin Account?</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete this administrator? This will revoke all access for this user. This action
              cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isDeleting}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              onClick={(e) => {
                e.preventDefault();
                confirmDelete();
              }}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              disabled={isDeleting}
            >
              {isDeleting ? "Deleting..." : "Delete Permanently"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
