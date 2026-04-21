import { useEffect, useState, useCallback, useMemo } from "react";
import { useParams, Link } from "react-router-dom";
import { Database, RefreshCw, Plus, Settings, ArrowLeft, Calendar, Layers, HardDrive, Edit } from "lucide-react";

import { Button } from "@/components/ui/button";
import { DataTable } from "@/components/ui/data-table";
import { fetchAPI } from "@/lib/api-client";
import { Badge } from "@/components/ui/badge";
import { CollectionDialog } from "@/components/collections-dialog";
import { RecordDialog } from "@/components/record-dialog";
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
import { toast } from "sonner";
import { Trash2 } from "lucide-react";

import { generateCollectionColumns } from "@/components/global/collections/table/columns";
import { PageLoader } from "@/components/global/widget/loader";
import { NotFound } from "@/components/global/widget/not-found";
import { formatSize } from "@/lib/utils";

interface CollectionField {
  id: string;
  name: string;
  type: string;
  props: any;
}

interface Collection {
  name: string;
  schema: CollectionField[];
  is_system?: boolean;
  total_documents: number;
  total_bytes: number;
  created_at: string;
  updated_at: string;
}


export default function CollectionPreviewPage() {
  const { name } = useParams();
  const [collection, setCollection] = useState<Collection | null>(null);
  const [records, setRecords] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [selectedRows, setSelectedRows] = useState<any[]>([]);
  const [deleting, setDeleting] = useState(false);
  const [showDeleteAlert, setShowDeleteAlert] = useState(false);
  const [recordDialogOpen, setRecordDialogOpen] = useState(false);
  const [recordToEdit, setRecordToEdit] = useState<any>(null);
  const [notFound, setNotFound] = useState(false);

  const fetchData = useCallback(
    async (isManual = false) => {
      try {
        if (isManual) setLoading(true);
        const [metaRes, recordsRes] = await Promise.all([
          fetchAPI(`/collection/${name}`),
          fetchAPI(`/collection/${name}/records`),
        ]);

        if (metaRes.success && !metaRes.data?.is_system) {
          setCollection(metaRes.data);
        } else {
          setNotFound(true);
        }
        if (recordsRes.success) {
          setRecords(recordsRes.data || []);
        }
        setSelectedRows([]); // Clear selection on refresh
      } catch (error) {
        console.error("Failed to fetch collection data:", error);
      } finally {
        setLoading(false);
      }
    },
    [name],
  );

  useEffect(() => {
    let isMounted = true;

    async function initialLoad() {
      try {
        const [metaRes, recordsRes] = await Promise.all([
          fetchAPI(`/collection/${name}`),
          fetchAPI(`/collection/${name}/records`),
        ]);

        if (isMounted) {
          if (metaRes.success && !metaRes.data?.is_system) {
            setCollection(metaRes.data);
            if (recordsRes.success) setRecords(recordsRes.data || []);
          } else {
            setNotFound(true);
          }
          setLoading(false);
        }
      } catch (error) {
        if (isMounted) {
          console.error("Initial load failed:", error);
          setLoading(false);
        }
      }
    }

    initialLoad();

    return () => {
      isMounted = false;
    };
  }, [name]);

  const columns = useMemo(() => {
    const cols = generateCollectionColumns(collection?.schema || []);
    if (cols.length > 0 && !cols.find((c: any) => c.id === "actions")) {
      cols.push({
        id: "actions",
        header: () => <div className="text-right">Actions</div>,
        cell: ({ row }: any) => (
          <div className="flex items-center justify-end gap-2">
            <Button
              variant="ghost"
              size="icon"
              onClick={(e) => {
                e.stopPropagation();
                setRecordToEdit(row.original);
                setRecordDialogOpen(true);
              }}
            >
              <Edit className="size-4" />
            </Button>
          </div>
        ),
      });
    }
    return cols;
  }, [collection?.schema]);

  const handleBulkDelete = async () => {
    if (selectedRows.length === 0) return;

    setDeleting(true);
    try {
      const ids = selectedRows.map((row) => row.id);
      const res = await fetchAPI(`/collection/${name}/records/bulk-delete`, {
        method: "POST",
        body: JSON.stringify({ ids }),
      });

      if (res.success) {
        toast.success(`Successfully deleted ${res.data.deleted_count} records`);
        fetchData();
        setShowDeleteAlert(false);
      } else {
        toast.error("Failed to delete records");
      }
    } catch {
      toast.error("An error occurred while deleting records");
    } finally {
      setDeleting(false);
    }
  };

  if (loading && !collection) {
    return (
      <PageLoader
        title={`Loading ${name}`}
        description="Please wait while we fetch the collection schema and its records..."
      />
    );
  }

  if (notFound) {
    return (
      <NotFound title="Collection Not Found" message={`The collection "${name}" does not exist or has been deleted.`} />
    );
  }

  return (
    <div className="flex flex-col gap-8">
      <CollectionDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        onSuccess={() => fetchData(false)}
        mode="edit"
        initialData={
          collection
            ? {
                name: collection.name,
                fields: collection.schema as any,
              }
            : undefined
        }
      />

      <RecordDialog
        open={recordDialogOpen}
        onOpenChange={(op) => {
          setRecordDialogOpen(op);
          if (!op) setTimeout(() => setRecordToEdit(null), 200);
        }}
        collectionName={name || ""}
        schema={collection?.schema || []}
        recordToEdit={recordToEdit}
        onSuccess={fetchData}
      />

      {/* Header */}
      <div className="flex flex-col gap-6">
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <Link to="/collections" className="hover:text-foreground transition-colors flex items-center gap-1">
            <ArrowLeft className="size-4" />
            Collections
          </Link>
          <span className="opacity-40">/</span>
          <span className="text-foreground font-medium">{name}</span>
        </div>

        <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div className="flex items-start gap-4">
            <div className="hidden sm:flex size-12 shrink-0 items-center justify-center rounded-lg bg-muted text-muted-foreground">
              <Database className="size-6" />
            </div>
            <div className="flex flex-col gap-1 min-w-0">
              <h1 className="text-2xl font-bold tracking-tight truncate">{name}</h1>
              <div className="flex flex-wrap items-center gap-3 text-sm text-muted-foreground">
                <div className="flex items-center gap-1.5">
                  <Calendar className="size-3.5" />
                  <span>Created {collection ? new Date(collection.created_at).toLocaleDateString() : "..."}</span>
                </div>
                <div className="flex items-center gap-1.5">
                  <Layers className="size-3.5" />
                  <span>{collection?.schema.length || 0} Fields</span>
                </div>
                <div className="flex items-center gap-1.5">
                  <HardDrive className="size-3.5" />
                  <span>{formatSize(collection?.total_bytes || 0)}</span>
                </div>
              </div>
            </div>
          </div>

          <div className="flex flex-wrap items-center gap-3">
            <Button
              variant="outline"
              size="lg"
              onClick={() => fetchData(true)}
              disabled={loading}
              className="flex-1 sm:flex-none"
            >
              <RefreshCw className={`mr-2 h-4 w-4 ${loading ? "animate-spin" : ""}`} />
              Refresh
            </Button>
            <Button variant="outline" size="lg" onClick={() => setDialogOpen(true)} className="flex-1 sm:flex-none">
              <Settings className="mr-2 h-4 w-4" />
              Schema
            </Button>
            <Button
              size="lg"
              onClick={() => {
                setRecordToEdit(null);
                setRecordDialogOpen(true);
              }}
              className="w-full sm:w-auto"
            >
              <Plus className="mr-2 h-4 w-4" />
              New Record
            </Button>
          </div>
        </div>
      </div>

      {/* Data Table */}
      <div className="flex flex-col gap-4">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold flex items-center gap-2">
            <Layers className="h-5 w-5 text-muted-foreground" />
            Records
            <Badge variant="secondary" className="ml-2 font-mono">
              {collection?.total_documents !== undefined && collection.total_documents > 0
                ? collection.total_documents
                : records.length}
            </Badge>
          </h2>
          {selectedRows.length > 0 && (
            <Button variant="destructive" size="sm" onClick={() => setShowDeleteAlert(true)} disabled={deleting}>
              <Trash2 className="mr-2 h-4 w-4" />
              Delete {selectedRows.length} Selected
            </Button>
          )}
        </div>

        <DataTable
          data={records}
          columns={columns}
          searchKey={collection?.schema[0]?.name?.toLowerCase() || "id"}
          onRowSelectionChange={setSelectedRows}
          onRowClick={(row) => {
            setRecordToEdit(row);
            setRecordDialogOpen(true);
          }}
        />
      </div>

      <AlertDialog open={showDeleteAlert} onOpenChange={setShowDeleteAlert}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete {selectedRows.length} Records?</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete the selected records from <strong>{name}</strong>? This action cannot be
              undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction variant="destructive" onClick={handleBulkDelete}>
              Delete Permanently
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
