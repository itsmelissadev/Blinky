import { useEffect, useState, useCallback } from "react";
import { Plus, RefreshCw } from "lucide-react";
import { useNavigate } from "react-router-dom";

import { Button } from "@/components/ui/button";
import { fetchAPI } from "@/lib/api-client";
import { DataTable } from "@/components/ui/data-table";
import { CollectionDialog } from "@/components/collections-dialog";

import { Collection, getCollectionColumns } from "@/components/global/collections/table/collections-table";
import { PageLoader } from "@/components/global/widget/loader";

export default function CollectionsPage() {
  const navigate = useNavigate();
  const [collections, setCollections] = useState<Collection[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editCollection, setEditCollection] = useState<Collection | null>(null);

  const fetchCollections = useCallback(async () => {
    try {
      const json = await fetchAPI("/collections");
      if (json.success) {
        const items = json.data || [];
        setCollections(items.filter((col: any) => !col.is_system));
      }
    } catch (error) {
      console.error("Failed to fetch collections:", error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchCollections();
  }, [fetchCollections]);

  const handleManualSync = async () => {
    setLoading(true);
    await fetchCollections();
  };

  const handleCreateNew = () => {
    setEditCollection(null);
    setDialogOpen(true);
  };

  const handleEdit = (col: Collection) => {
    setEditCollection(col);
    setDialogOpen(true);
  };

  const columns = getCollectionColumns(handleEdit);

  if (loading && collections.length === 0) {
    return (
      <PageLoader
        title="Fetching Collections"
        description="Please wait while we establish a secure connection to your database..."
      />
    );
  }

  return (
    <div className="flex flex-col gap-8">
      {/* Header */}
      <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div className="flex flex-col gap-1">
          <h1 className="text-2xl font-bold tracking-tight">Collections</h1>
          <p className="text-sm text-muted-foreground">
            Manage your database collections and schema structures in real-time.
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-3">
          <Button
            variant="outline"
            onClick={handleManualSync}
            disabled={loading}
            size="lg"
            className="flex-1 sm:flex-none"
          >
            <RefreshCw className={`mr-2 h-4 w-4 ${loading ? "animate-spin" : ""}`} />
            Sync
          </Button>
          <Button onClick={handleCreateNew} size="lg" className="w-full sm:w-auto">
            <Plus className="mr-2 h-4 w-4" />
            New Collection
          </Button>
        </div>
      </div>

      <DataTable
        data={collections || []}
        columns={columns}
        searchKey="name"
        onRowClick={(row) => navigate(`/collections/${row.name}`)}
      />

      <CollectionDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        onSuccess={fetchCollections}
        mode={editCollection ? "edit" : "create"}
        initialData={
          editCollection
            ? {
                name: editCollection.name,
                fields: editCollection.schema as any,
              }
            : undefined
        }
      />
    </div>
  );
}
