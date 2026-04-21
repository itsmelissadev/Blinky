"use client";

import { useEffect, useState, useCallback, useMemo } from "react";
import { Button } from "@/components/ui/button";
import { RefreshCw, Plus, Trash2, Settings, Key, Eye, EyeOff, Loader2, Edit } from "lucide-react";
import { toast } from "sonner";
import { Input } from "@/components/ui/input";
import { fetchAPI } from "@/lib/api-client";
import { DataTable, DataTableColumnHeader } from "@/components/ui/data-table";
import { Checkbox } from "@/components/ui/checkbox";
import { PageLoader } from "@/components/global/widget/loader";
import { Label } from "@/components/ui/label";
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

interface EnvVar {
  key: string;
  value: string;
}

const SYSTEM_KEYS = [
  "POSTGRESQL_DB_HOST",
  "POSTGRESQL_DB_PORT",
  "POSTGRESQL_DB_USER",
  "POSTGRESQL_DB_PASSWORD",
  "POSTGRESQL_DB_NAME",
  "POSTGRESQL_FOLDER_PATH",
  "POSTGRESQL_DATA_PATH",
  "PUBLIC_API_PORT",
  "PUBLIC_API_HOST",
  "ADMIN_PANEL_HOST",
  "ADMIN_PANEL_PORT",
];

export default function EnvironmentsPage() {
  const [vars, setVars] = useState<EnvVar[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [selectedKeys, setSelectedKeys] = useState<Record<string, boolean>>({});
  const [isDeleting, setIsDeleting] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [keysToDelete, setKeysToDelete] = useState<string[]>([]);
  const [editVar, setEditVar] = useState<EnvVar | null>(null);
  const [showInTable, setShowInTable] = useState<Record<string, boolean>>({});

  const fetchEnv = useCallback(async () => {
    try {
      const res = await fetchAPI("/settings/environments");
      if (res.success) {
        setVars(res.data || []);
        setSelectedKeys({});
      }
    } catch (error) {
      toast.error("Failed to fetch environment variables");
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchEnv();
  }, [fetchEnv]);

  const { systemVars, userVars } = useMemo(() => {
    const sys: EnvVar[] = [];
    const usr: EnvVar[] = [];
    vars.forEach((v) => {
      if (SYSTEM_KEYS.includes(v.key)) {
        sys.push(v);
      } else {
        usr.push(v);
      }
    });
    return { systemVars: sys, userVars: usr };
  }, [vars]);

  const handleEdit = (v: EnvVar) => {
    setEditVar(v);
    setEditDialogOpen(true);
  };

  const handleCreateNew = () => {
    setEditVar(null);
    setEditDialogOpen(true);
  };

  const confirmDelete = (keys: string[]) => {
    setKeysToDelete(keys);
    setDeleteDialogOpen(true);
  };

  const handleDeleteExecute = async () => {
    setIsDeleting(true);
    try {
      const res = await fetchAPI("/settings/environments", {
        method: "DELETE",
        body: JSON.stringify({ keys: keysToDelete }),
      });
      if (res.success) {
        toast.success(keysToDelete.length > 1 ? "Variables deleted" : "Variable deleted");
        await fetchEnv();
      } else {
        toast.error(res.error?.message || "Delete failed");
      }
    } catch (error) {
      toast.error("An error occurred");
    } finally {
      setIsDeleting(false);
      setDeleteDialogOpen(false);
      setKeysToDelete([]);
    }
  };

  const toggleSelectAll = (isSystem: boolean, checked: boolean) => {
    if (isSystem) return;
    const newSelected = { ...selectedKeys };
    userVars.forEach((v) => {
      newSelected[v.key] = checked;
    });
    setSelectedKeys(newSelected);
  };

  const toggleSelectRow = (key: string, checked: boolean) => {
    setSelectedKeys((prev) => ({ ...prev, [key]: checked }));
  };

  const selectedCount = useMemo(() => Object.keys(selectedKeys).filter((k) => selectedKeys[k]).length, [selectedKeys]);

  const getColumns = (isSystem: boolean) => {
    const cols: any[] = [];

    if (!isSystem) {
      cols.push({
        id: "select",
        header: () => (
          <div className="flex items-center justify-center w-5">
            <Checkbox
              checked={userVars.length > 0 && userVars.every((v) => selectedKeys[v.key])}
              onCheckedChange={(value) => toggleSelectAll(false, !!value)}
              aria-label="Select all"
            />
          </div>
        ),
        cell: ({ row }: any) => (
          <div className="flex items-center justify-center w-5">
            <Checkbox
              checked={!!selectedKeys[row.original.key]}
              onCheckedChange={(value) => toggleSelectRow(row.original.key, !!value)}
              onClick={(e) => e.stopPropagation()}
              aria-label="Select row"
            />
          </div>
        ),
      });
    }

    cols.push(
      {
        accessorKey: "key",
        header: ({ column }: any) => <DataTableColumnHeader column={column} title="Constant" icon={Key} />,
        cell: ({ row }: any) => (
          <div className="flex items-center gap-2 py-1">
            <span className="font-mono text-sm font-semibold tracking-tight text-foreground">{row.original.key}</span>
          </div>
        ),
      },
      {
        accessorKey: "value",
        header: ({ column }: any) => <DataTableColumnHeader column={column} title="Value" icon={Key} />,
        cell: ({ row }: any) => {
          const isRevealed = showInTable[row.original.key];
          return (
            <div className="flex items-center gap-2 group">
              <span className="font-mono text-sm font-medium text-muted-foreground/80">
                {isRevealed ? row.original.value : "••••••••••••••••"}
              </span>
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  setShowInTable((prev) => ({ ...prev, [row.original.key]: !isRevealed }));
                }}
                className="ml-auto opacity-0 group-hover:opacity-100 transition-opacity p-1 hover:bg-muted rounded text-muted-foreground hover:text-foreground"
              >
                {isRevealed ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
              </button>
            </div>
          );
        },
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
                handleEdit(row.original);
              }}
              className="h-8 w-8 hover:bg-primary/10 transition-all shadow-sm"
            >
              {isSystem ? <Settings className="h-3.5 w-3.5" /> : <Edit className="h-3.5 w-3.5" />}
            </Button>
            {!isSystem && (
              <Button
                variant="outline"
                size="icon"
                onClick={(e) => {
                  e.stopPropagation();
                  confirmDelete([row.original.key]);
                }}
                className="h-8 w-8 hover:bg-destructive/10 hover:text-destructive hover:border-destructive/30 transition-all shadow-sm"
              >
                <Trash2 className="h-3.5 w-3.5" />
              </Button>
            )}
          </div>
        ),
      },
    );

    return cols;
  };

  if (isLoading && vars.length === 0) {
    return (
      <PageLoader
        title="Fetching Environments"
        description="Please wait while we establish a secure connection to your system configurations..."
      />
    );
  }

  return (
    <div className="flex flex-col gap-12">
      {/* Header */}
      <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div className="flex flex-col gap-1">
          <h1 className="text-2xl font-bold tracking-tight">Environments</h1>
          <p className="text-sm text-muted-foreground">
            Manage your environment variables and system secrets stored in the .env file.
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-3">
          {selectedCount > 0 && (
            <Button
              variant="destructive"
              onClick={() => confirmDelete(Object.keys(selectedKeys).filter((k) => selectedKeys[k]))}
              disabled={isDeleting}
              size="lg"
              className="flex-1 sm:flex-none animate-in fade-in slide-in-from-right-2"
            >
              {isDeleting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Trash2 className="mr-2 h-4 w-4" />}
              Delete ({selectedCount})
            </Button>
          )}
          <Button variant="outline" onClick={fetchEnv} disabled={isLoading} size="lg" className="flex-1 sm:flex-none">
            <RefreshCw className={`mr-2 h-4 w-4 ${isLoading ? "animate-spin" : ""}`} />
            Sync
          </Button>
          <Button onClick={handleCreateNew} size="lg" className="w-full sm:w-auto">
            <Plus className="mr-2 h-4 w-4" />
            Add Variable
          </Button>
        </div>
      </div>

      <div className="space-y-6">
        <div>
          <h2 className="text-lg font-semibold tracking-tight">User Variables</h2>
          <p className="text-sm text-muted-foreground">Custom environment variables for your application.</p>
        </div>
        <DataTable data={userVars} columns={getColumns(false)} searchKey="key" onRowClick={(row) => handleEdit(row)} />
      </div>

      <div className="space-y-6">
        <div>
          <h2 className="text-lg font-semibold tracking-tight">System Constants</h2>
          <p className="text-sm text-muted-foreground">Core engine configurations that cannot be deleted.</p>
        </div>
        <DataTable data={systemVars} columns={getColumns(true)} searchKey="key" onRowClick={(row) => handleEdit(row)} />
      </div>

      <EnvironmentEditDialog
        open={editDialogOpen}
        onOpenChange={setEditDialogOpen}
        onSuccess={fetchEnv}
        initialData={editVar}
      />

      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
            <AlertDialogDescription>
              This action cannot be undone. This will permanently delete {keysToDelete.length}{" "}
              {keysToDelete.length > 1 ? "variables" : "variable"} from your configuration file.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isDeleting}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              onClick={(e) => {
                e.preventDefault();
                handleDeleteExecute();
              }}
              disabled={isDeleting}
            >
              {isDeleting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : "Delete Permanently"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

function EnvironmentEditDialog({
  open,
  onOpenChange,
  onSuccess,
  initialData,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
  initialData: EnvVar | null;
}) {
  const [key, setKey] = useState("");
  const [value, setValue] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  const isSystem = initialData && SYSTEM_KEYS.includes(initialData.key);

  useEffect(() => {
    if (initialData) {
      setKey(initialData.key);
      setValue(initialData.value);
    } else {
      setKey("");
      setValue("");
    }
  }, [initialData, open]);

  const handleSave = async () => {
    if (!key) {
      toast.error("Key is required");
      return;
    }
    setIsLoading(true);
    try {
      const res = await fetchAPI("/settings/environments", {
        method: "PATCH",
        body: JSON.stringify({
          oldKey: initialData?.key || "",
          key: key,
          value: value,
        }),
      });
      if (res.success) {
        toast.success(initialData ? "Variable updated" : "Variable added");
        onOpenChange(false);
        onSuccess();
      } else {
        toast.error(res.error?.message || "Operation failed");
      }
    } catch (error) {
      toast.error("An error occurred");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{initialData ? "Edit Variable" : "Add Environment Variable"}</AlertDialogTitle>
          <AlertDialogDescription>
            {initialData
              ? "Update the configuration for this environment variable."
              : "Add a new environment variable to your system configuration."}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="key">Constant Name</Label>
            <Input
              id="key"
              value={key}
              onChange={(e) => setKey(e.target.value.toUpperCase())}
              placeholder="e.g. DATABASE_URL"
              className="font-mono uppercase"
              disabled={!!isSystem}
            />
            {isSystem && <p className="text-[10px] text-muted-foreground">System constant names cannot be modified.</p>}
          </div>
          <div className="grid gap-2">
            <Label htmlFor="value">Value</Label>
            <Input
              id="value"
              value={value}
              onChange={(e) => setValue(e.target.value)}
              placeholder="Enter variable value"
              className="font-mono"
            />
          </div>
        </div>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isLoading}>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={(e) => {
              e.preventDefault();
              handleSave();
            }}
            disabled={isLoading}
          >
            {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            {initialData ? "Save Changes" : "Create Variable"}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
