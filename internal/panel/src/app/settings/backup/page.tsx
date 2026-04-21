"use client";

import { useEffect, useState, useCallback } from "react";
import { Button } from "@/components/ui/button";
import { PlaySquare, Download, Trash2, Database, Loader2, RefreshCw, ShieldAlert } from "lucide-react";
import { toast } from "sonner";
import { fetchAPI, getAPIUrl } from "@/lib/api-client";
import { formatDate, formatSize } from "@/lib/utils";
import { DataTable, DataTableColumnHeader } from "@/components/ui/data-table";
import { PageLoader } from "@/components/global/widget/loader";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
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

interface Backup {
  name: string;
  size: number;
  createdAt: string;
}

export default function BackupPage() {
  const [isBackingUp, setIsBackingUp] = useState(false);
  const [backups, setBackups] = useState<Backup[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);
  const [backupToDelete, setBackupToDelete] = useState<string | null>(null);

  const fetchBackups = useCallback(async () => {
    try {
      const res = await fetchAPI("/settings/backup");
      if (res.success) {
        setBackups(res.data || []);
      }
    } catch (error) {
      toast.error("Failed to fetch backups");
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchBackups();
  }, [fetchBackups]);

  const handleBackup = async () => {
    setIsBackingUp(true);
    try {
      const res = await fetchAPI("/settings/backup", { method: "POST" });
      if (res.success) {
        toast.success("Database backup created successfully!");
        fetchBackups();
      } else {
        toast.error(res.error?.message || "Backup failed");
      }
    } catch (error) {
      toast.error("An error occurred while creating backup");
    } finally {
      setIsBackingUp(false);
    }
  };

  const handleDeleteExecute = async () => {
    if (!backupToDelete) return;
    try {
      const res = await fetchAPI(`/settings/backup/${backupToDelete}`, { method: "DELETE" });
      if (res.success) {
        toast.success("Backup deleted successfully");
        setBackups(backups.filter((b) => b.name !== backupToDelete));
      } else {
        toast.error(res.error?.message || "Delete failed");
      }
    } catch (error) {
      toast.error("An error occurred while deleting backup");
    } finally {
      setIsDeleteConfirmOpen(false);
      setBackupToDelete(null);
    }
  };

  const handleDownload = (filename: string) => {
    const url = getAPIUrl(`/settings/backup/download/${filename}`);
    window.open(url, "_blank");
  };

  const columns = [
    {
      accessorKey: "name",
      header: ({ column }: any) => <DataTableColumnHeader column={column} title="File Name" icon={Database} />,
      cell: ({ row }: any) => (
        <div className="flex items-center gap-2 py-1">
          <span className="font-mono text-sm font-semibold tracking-tight text-foreground">{row.original.name}</span>
        </div>
      ),
    },
    {
      accessorKey: "createdAt",
      header: ({ column }: any) => <DataTableColumnHeader column={column} title="Date" />,
      cell: ({ row }: any) => (
        <span className="text-sm text-muted-foreground">{formatDate(row.original.createdAt)}</span>
      ),
    },
    {
      accessorKey: "size",
      header: ({ column }: any) => <DataTableColumnHeader column={column} title="Size" />,
      cell: ({ row }: any) => (
        <span className="text-sm text-muted-foreground font-mono">{formatSize(row.original.size)}</span>
      ),
    },
    {
      id: "actions",
      header: () => <div className="text-right px-4">Actions</div>,
      cell: ({ row }: any) => (
        <div className="flex items-center justify-end gap-2 px-4">
          <Button
            variant="outline"
            size="icon"
            onClick={(e) => {
              e.stopPropagation();
              handleDownload(row.original.name);
            }}
            className="h-8 w-8 transition-all shadow-sm"
          >
            <Download className="h-3.5 w-3.5" />
          </Button>
          <Button
            variant="outline"
            size="icon"
            onClick={(e) => {
              e.stopPropagation();
              setBackupToDelete(row.original.name);
              setIsDeleteConfirmOpen(true);
            }}
            className="h-8 w-8 hover:bg-destructive/10 hover:text-destructive hover:border-destructive/30 transition-all shadow-sm"
          >
            <Trash2 className="h-3.5 w-3.5" />
          </Button>
        </div>
      ),
    },
  ];

  if (isLoading && backups.length === 0) {
    return (
      <PageLoader
        title="Fetching Backups"
        description="Please wait while we establish a secure connection to your backup storage..."
      />
    );
  }

  return (
    <div className="flex flex-col gap-8">
      {/* Header */}
      <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div className="flex flex-col gap-1">
          <h1 className="text-2xl font-bold tracking-tight">Database Backups</h1>
          <p className="text-sm text-muted-foreground">Manage and store secure snapshots of your engine's database.</p>
        </div>
        <div className="flex flex-wrap items-center gap-3">
          <Button variant="outline" onClick={fetchBackups} disabled={isLoading} size="lg">
            <RefreshCw className={`mr-2 h-4 w-4 ${isLoading ? "animate-spin" : ""}`} />
            {isLoading ? "Syncing..." : "Sync"}
          </Button>
          <Button onClick={handleBackup} disabled={isBackingUp} size="lg">
            {isBackingUp ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <PlaySquare className="mr-2 h-4 w-4" />}
            {isBackingUp ? "Creating..." : "Create Backup"}
          </Button>
        </div>
      </div>

      <DataTable data={backups} columns={columns} searchKey="name" />

      {/* Delete Confirmation AlertDialog */}
      <AlertDialog open={isDeleteConfirmOpen} onOpenChange={setIsDeleteConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
            <AlertDialogDescription>
              This will permanently delete the backup file. This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={(e) => {
                e.preventDefault();
                handleDeleteExecute();
              }}
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
