import * as React from "react";
import { Plus, Trash2, GripVertical, Settings2, ChevronDown, Info, MoreVertical, Eraser, Trash } from "lucide-react";
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription, SheetFooter } from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { ScrollArea } from "@/components/ui/scroll-area";
import { toast } from "sonner";
import { cn } from "@/lib/utils";

import { fetchAPI } from "@/lib/api-client";
import { useSchema, type FieldType, type FieldDefinition } from "@/hooks/use-schema";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { ButtonGroup } from "./ui/button-group";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuGroup,
  DropdownMenuLabel,
} from "@/components/ui/dropdown-menu";
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

interface CollectionDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
  mode?: "create" | "edit";
  initialData?: {
    name: string;
    fields: FieldDefinition[];
  };
}

const PropertyRow = ({
  label,
  description,
  bottomBorder,
  disableTopPadding,
  disableBottomPadding,
  children,
  className,
}: {
  label: string;
  description: string;
  bottomBorder?: boolean;
  disableTopPadding?: boolean;
  disableBottomPadding?: boolean;
  children: React.ReactNode;
  className?: string;
}) => (
  <div
    className={cn(
      `flex items-center justify-between ${disableTopPadding ? "pt-0" : "pt-3"} ${disableBottomPadding ? "pb-0" : "pb-3"} ${bottomBorder ? "border-b" : "border-none"}`,
      className,
    )}
  >
    <div className="space-y-0.5">
      <div className="flex items-center gap-1.5">
        <Label className="text-sm font-medium">{label}</Label>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button variant="ghost" size="icon" className="size-4 h-auto w-auto p-0 opacity-50 hover:opacity-100">
              <Info className="size-3.5" />
            </Button>
          </TooltipTrigger>
          <TooltipContent side="right" className="max-w-[200px]">
            {description}
          </TooltipContent>
        </Tooltip>
      </div>
      <p className="text-xs text-muted-foreground">{description}</p>
    </div>
    <div className="flex items-center">{children}</div>
  </div>
);

const generateId = () => {
  if (typeof window !== "undefined" && window.crypto && window.crypto.randomUUID) {
    return window.crypto.randomUUID();
  }
  return Math.random().toString(36).substring(2, 11) + Date.now().toString(36);
};

export function CollectionDialog({
  open,
  onOpenChange,
  onSuccess,
  mode = "create",
  initialData,
}: CollectionDialogProps) {
  const { types, loading: schemaLoading, getTypeConfig } = useSchema();
  const [name, setName] = React.useState("");
  const [loading, setLoading] = React.useState(false);
  const [fields, setFields] = React.useState<FieldDefinition[]>([]);
  const [collectionsList, setCollectionsList] = React.useState<{ name: string }[]>([]);
  const [activeSettings, setActiveSettings] = React.useState<Record<number, boolean>>({});
  const [draggedIndex, setDraggedIndex] = React.useState<number | null>(null);
  const [dragOverIndex, setDragOverIndex] = React.useState<number | null>(null);
  const [activeDragHandleIndex, setActiveDragHandleIndex] = React.useState<number | null>(null);
  const [showDeleteAlert, setShowDeleteAlert] = React.useState(false);
  const [showTruncateAlert, setShowTruncateAlert] = React.useState(false);

  const [showSaveAlert, setShowSaveAlert] = React.useState(false);
  const [changes, setChanges] = React.useState<
    { type: "added" | "removed" | "modified"; name: string; details?: string }[]
  >([]);

  const [lastOpen, setLastOpen] = React.useState(false);
  const [lastMode, setLastMode] = React.useState<"create" | "edit">("create");

  const isEdit = mode === "edit";

  // Sync state when dialog opens or mode changes
  if (open !== lastOpen || mode !== lastMode) {
    setLastOpen(open);
    setLastMode(mode);

    if (open) {
      if (isEdit && initialData) {
        setName(initialData.name);
        setFields(initialData.fields.map((f) => ({ ...f, id: f.id || generateId() })));
      } else if (!isEdit) {
        setName("");
        setFields([
          {
            id: generateId(),
            name: "id",
            type: "id",
            props: { id: { auto_len: 16 }, is_nullable: false, is_unique: true },
          },
          {
            id: generateId(),
            name: "updated_at",
            type: "date",
            props: { auto_timestamp: "create_update", is_nullable: false },
          },
          {
            id: generateId(),
            name: "created_at",
            type: "date",
            props: { auto_timestamp: "create", is_nullable: false },
          },
        ]);
      }
      setActiveSettings({});
    }
  }

  React.useEffect(() => {
    if (open) {
      fetchAPI("/collections").then((res) => {
        if (res.success) setCollectionsList(res.data || []);
      });
    }
  }, [open]);

  const toggleSettings = (index: number) => {
    setActiveSettings((prev) => ({ ...prev, [index]: !prev[index] }));
  };

  const addField = () => {
    const newField: FieldDefinition = {
      id: generateId(),
      name: `field_${fields.length}`,
      type: "text" as FieldType,
      props: { is_nullable: false, is_unique: false },
    };

    const newFields = [...fields];
    // Find index of first date system field (created_at/updated_at)
    const insertIndex = newFields.findIndex((f) => f.name === "created_at" || f.name === "updated_at");

    if (insertIndex !== -1) {
      newFields.splice(insertIndex, 0, newField);
    } else {
      newFields.push(newField);
    }
    setFields(newFields);
  };

  const moveField = (from: number, to: number) => {
    if (fields[from].type === "id" || fields[to].type === "id") return;
    const newFields = [...fields];
    const item = newFields.splice(from, 1)[0];
    newFields.splice(to, 0, item);
    setFields(newFields);
  };

  const removeField = (index: number) => {
    if (fields[index].type === "id") return;
    setFields(fields.filter((_, i) => i !== index));
    const newSettings = { ...activeSettings };
    delete newSettings[index];
    setActiveSettings(newSettings);
  };

  const updateField = (index: number, updates: Partial<FieldDefinition>) => {
    const newFields = [...fields];
    const newField = { ...newFields[index], ...updates };

    // Reset boolean props if type changed to boolean
    if (updates.type === "boolean") {
      newField.props = {
        ...newField.props,
        is_nullable: false,
        is_unique: false,
      };
    }

    newFields[index] = newField;
    setFields(newFields);
  };

  const getChanges = () => {
    if (!initialData) return [];
    const changesList: { type: "added" | "removed" | "modified"; name: string; details?: string }[] = [];
    const oldFields = initialData.fields;
    const newFields = fields;

    // Added
    newFields.forEach((nf) => {
      if (!oldFields.find((of) => of.id === nf.id)) {
        changesList.push({ type: "added", name: nf.name });
      }
    });

    // Removed
    oldFields.forEach((of) => {
      if (!newFields.find((nf) => nf.id === of.id)) {
        changesList.push({ type: "removed", name: of.name });
      }
    });

    // Modified
    newFields.forEach((nf) => {
      const of = oldFields.find((f) => f.id === nf.id);
      if (of) {
        const details: string[] = [];

        if (of.name !== nf.name) {
          details.push(`Rename: "${of.name}" → "${nf.name}"`);
        }

        if (of.type !== nf.type) {
          details.push(`Type: ${of.type.toUpperCase()} → ${nf.type.toUpperCase()}`);
        }

        // Check common props
        if (of.props.is_nullable !== nf.props.is_nullable) {
          details.push(`Nullable: ${of.props.is_nullable} → ${nf.props.is_nullable}`);
        }
        if (of.props.is_unique !== nf.props.is_unique) {
          details.push(`Unique: ${of.props.is_unique} → ${nf.props.is_unique}`);
        }
        if (of.props.default !== nf.props.default) {
          const oldDef = of.props.default ?? "none";
          const newDef = nf.props.default ?? "none";
          details.push(`Default: "${oldDef}" → "${newDef}"`);
        }

        // Check Type Specific Props
        if (nf.type === "number" && of.type === "number") {
          const o = of.props.number || {};
          const n = nf.props.number || {};
          if (o.min !== n.min) details.push(`Min: ${o.min ?? "none"} → ${n.min ?? "none"}`);
          if (o.max !== n.max) details.push(`Max: ${o.max ?? "none"} → ${n.max ?? "none"}`);
          if (o.no_decimals !== n.no_decimals) details.push(`No Decimals: ${!!o.no_decimals} → ${!!n.no_decimals}`);
          if (o.no_zero !== n.no_zero) details.push(`No Zero: ${!!o.no_zero} → ${!!n.no_zero}`);
        }

        if (nf.type === "text" && of.type === "text") {
          const o = of.props.text || {};
          const n = nf.props.text || {};
          if (o.min !== n.min) details.push(`Min Length: ${o.min ?? "none"} → ${n.min ?? "none"}`);
          if (o.max !== n.max) details.push(`Max Length: ${o.max ?? "none"} → ${n.max ?? "none"}`);
        }

        if (details.length > 0) {
          changesList.push({
            type: "modified",
            name: nf.name,
            details: details.join(" | "),
          });
        }
      }
    });

    return changesList;
  };

  const handleSubmit = async (bypassConfirm = false) => {
    if (!name) {
      toast.error("Collection name is required");
      return;
    }

    // Client-side validation
    for (const field of fields) {
      const { props, type, name: fName } = field;

      // Text length validation
      if (type === "text" && props.default) {
        const d = props.default as string;
        if (props.text?.min !== undefined && d.length < props.text.min) {
          toast.error(`Field "${fName}": Default value shorter than min length (${props.text.min})`);
          return;
        }
        if (props.text?.max !== undefined && d.length > props.text.max) {
          toast.error(`Field "${fName}": Default value longer than max length (${props.text.max})`);
          return;
        }
      }

      // Numeric range validation
      if (type === "number" && props.default !== undefined && props.default !== "") {
        const val = parseFloat(props.default as string);
        const opts = props.number;
        if (opts?.min !== undefined && val < opts.min) {
          toast.error(`Field "${fName}": Default value ${val} is less than min ${opts.min}`);
          return;
        }
        if (opts?.max !== undefined && val > opts.max) {
          toast.error(`Field "${fName}": Default value ${val} is greater than max ${opts.max}`);
          return;
        }
      }
    }

    if (isEdit && !bypassConfirm) {
      const detectedChanges = getChanges();
      if (detectedChanges.length > 0) {
        setChanges(detectedChanges);
        setShowSaveAlert(true);
        return;
      }
    }

    setLoading(true);
    try {
      if (isEdit) {
        const res = await fetchAPI(`/collection/${initialData?.name}`, {
          method: "PATCH",
          body: JSON.stringify({ name, schema: fields }),
        });

        if (res.success) {
          toast.success("Collection updated successfuly");
          onSuccess();
          onOpenChange(false);
        } else {
          toast.error(res.error?.message || "Failed to update collection");
        }
      } else {
        const json = await fetchAPI("/collection", {
          method: "POST",
          body: JSON.stringify({ name, schema: fields }),
        });
        if (json.success) {
          toast.success("Collection created successfuly");
          onSuccess();
          onOpenChange(false);
        } else {
          toast.error(json.error?.message || "Failed to create collection");
        }
      }
    } catch {
      toast.error("Network connection failure");
    } finally {
      setLoading(false);
    }
  };

  const handleTruncate = async () => {
    setLoading(true);
    try {
      const res = await fetchAPI(`/collection/${initialData?.name}/truncate`, { method: "DELETE" });
      if (res.success) {
        toast.success("Collection truncated successfully");
        setShowTruncateAlert(false);
      } else {
        toast.error(res.error?.message || "Failed to truncate collection");
      }
    } catch {
      toast.error("Network connection failure");
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async () => {
    setLoading(true);
    try {
      const res = await fetchAPI(`/collection/${initialData?.name}`, { method: "DELETE" });
      if (res.success) {
        toast.success("Collection deleted successfully");
        setShowDeleteAlert(false);
        onSuccess();
        onOpenChange(false);
      } else {
        toast.error(res.error?.message || "Failed to delete collection");
      }
    } catch {
      toast.error("Network connection failure");
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <Sheet open={open} onOpenChange={onOpenChange}>
        <SheetContent className="max-sm:w-full! max-sm:inset-0! md:max-w-xl! flex flex-col p-0 gap-0 h-full bg-background border-l shadow-2xl">
          <div tabIndex={0} className="sr-only" aria-hidden="true" />
          <SheetHeader className="p-4 border-b gap-0 relative">
            <div className="flex items-center justify-between">
              <div>
                <SheetTitle>{isEdit ? "Edit Collection" : "Create Collection"}</SheetTitle>
                <SheetDescription>
                  {isEdit ? "Modify your collection structure" : "Define your schema and field constraints"}
                </SheetDescription>
              </div>

              {isEdit && (
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="icon" className="h-8 w-8">
                      <MoreVertical className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" className="w-40">
                    <DropdownMenuGroup>
                      <DropdownMenuLabel>Danger Zone</DropdownMenuLabel>
                      <DropdownMenuItem variant="destructive" onClick={() => setShowTruncateAlert(true)}>
                        <Eraser className="mr-2 h-4 w-4" />
                        Truncate Data
                      </DropdownMenuItem>
                      <DropdownMenuItem variant="destructive" onClick={() => setShowDeleteAlert(true)}>
                        <Trash className="mr-2 h-4 w-4" />
                        Delete Collection
                      </DropdownMenuItem>
                    </DropdownMenuGroup>
                  </DropdownMenuContent>
                </DropdownMenu>
              )}
            </div>
          </SheetHeader>

          <ScrollArea className="flex-1 min-h-0 w-full">
            <div className="p-4 space-y-4">
              {/* General Settings */}
              <div className="space-y-2">
                <Label htmlFor="name" className="text-sm font-semibold mb-1 block">
                  Collection Name
                </Label>
                <Input
                  id="name"
                  placeholder="e.g. products, cloud_users"
                  className="font-medium h-12 px-4 shadow-none"
                  value={name}
                  onChange={(e) => setName(e.target.value.toLowerCase().replace(/[^a-z0-9_]/g, ""))}
                />
                <p className="text-xs text-muted-foreground">Lowercase alphanumeric characters and underscores only.</p>
              </div>

              {/* Schema Definition */}
              <div className="space-y-4">
                <div className="flex items-center justify-between pb-2">
                  <span className="text-sm font-semibold">Schema</span>

                  <Button
                    onClick={addField}
                    disabled={schemaLoading || types.length === 0}
                    variant="outline"
                    size="lg"
                    className="text-xs"
                  >
                    <Plus className="mr-1.5 h-3.5 w-3.5" />
                    Add Field
                  </Button>
                </div>

                <div className="space-y-2">
                  {fields.map((field, index) => {
                    const Config = getTypeConfig(field.type);
                    const isOpen = activeSettings[index];

                    return (
                      <div
                        key={index}
                        data-droppable-index={index}
                        draggable={field.type !== "id" && activeDragHandleIndex === index}
                        onDragStart={(e) => {
                          if (field.type === "id") return;
                          setDraggedIndex(index);
                          e.dataTransfer.setData("text/plain", index.toString());
                          // Set drag image, etc. if needed
                        }}
                        onDragEnter={(e) => {
                          e.preventDefault();
                          if (draggedIndex !== null && draggedIndex !== index && field.type !== "id") {
                            setDragOverIndex(index);
                          }
                        }}
                        onDragOver={(e) => {
                          e.preventDefault(); // Necessary to allow dropping
                        }}
                        onDragLeave={(e) => {
                          e.preventDefault();
                          // Sadece elementten tamamen yandaki bir alana çıkılınca temizlemek isterseniz:
                          // if (dragOverIndex === index) setDragOverIndex(null);
                        }}
                        onDrop={(e) => {
                          e.preventDefault();
                          if (draggedIndex !== null && draggedIndex !== index && field.type !== "id") {
                            moveField(draggedIndex, index);
                          }
                          setDraggedIndex(null);
                          setDragOverIndex(null);
                          setActiveDragHandleIndex(null);
                        }}
                        onDragEnd={() => {
                          setDraggedIndex(null);
                          setDragOverIndex(null);
                          setActiveDragHandleIndex(null);
                        }}
                        className={cn(
                          "group rounded-lg border transition-all duration-200 relative",
                          draggedIndex === index && "opacity-40 border-primary/50 scale-[0.98]",
                          dragOverIndex === index && draggedIndex !== index && "border-t-2 border-t-primary mt-2",
                        )}
                      >
                        <div className="flex items-center gap-3 p-2.5">
                          <div
                            onMouseEnter={() => field.type !== "id" && setActiveDragHandleIndex(index)}
                            onMouseLeave={() => setActiveDragHandleIndex(null)}
                            onTouchStart={(e) => {
                              if (field.type === "id") return;
                              setDraggedIndex(index);
                              setActiveDragHandleIndex(index);
                              document.body.style.overflow = "hidden"; // Prevent page scroll
                            }}
                            onTouchMove={(e) => {
                              if (draggedIndex === null) return;
                              const touch = e.touches[0];
                              const target = document.elementFromPoint(touch.clientX, touch.clientY);
                              const dropTarget = target?.closest("[data-droppable-index]");
                              if (dropTarget) {
                                const idx = parseInt(dropTarget.getAttribute("data-droppable-index") || "", 10);
                                if (!isNaN(idx) && idx !== draggedIndex) {
                                  setDragOverIndex(idx);
                                } else {
                                  setDragOverIndex(null);
                                }
                              } else {
                                setDragOverIndex(null);
                              }
                            }}
                            onTouchEnd={() => {
                              document.body.style.overflow = ""; // Restore scroll
                              if (draggedIndex !== null && dragOverIndex !== null && draggedIndex !== dragOverIndex) {
                                moveField(draggedIndex, dragOverIndex);
                              }
                              setDraggedIndex(null);
                              setDragOverIndex(null);
                              setActiveDragHandleIndex(null);
                            }}
                            className={cn(
                              "cursor-grab active:cursor-grabbing text-muted-foreground/30 hover:text-muted-foreground transition-colors group-data-[state=open]:rotate-90 p-1 -m-1 touch-none",
                              field.type === "id" && "cursor-not-allowed opacity-0",
                            )}
                          >
                            <GripVertical className="h-4 w-4" />
                          </div>

                          <div
                            className={cn(
                              "flex h-8 w-8 shrink-0 items-center justify-center rounded-lg border",
                              Config.color,
                            )}
                          >
                            <Config.icon className="h-5 w-5 opacity-90" />
                          </div>

                          <div className="flex-1 flex items-center gap-2">
                            <Input
                              placeholder="field_name"
                              className="px-2 border focus-visible:ring-0 shadow-none text-sm placeholder:text-muted-foreground/40"
                              value={field.name}
                              disabled={field.type === "id"}
                              onChange={(e) =>
                                updateField(index, {
                                  name: e.target.value.toLowerCase().replace(/[^a-z0-9_]/g, ""),
                                })
                              }
                            />
                          </div>

                          <div className="flex items-center gap-2">
                            <Select
                              value={field.type}
                              disabled={field.type === "id"}
                              onValueChange={(val) => {
                                if (!val) return;
                                const updates: any = { type: val as FieldType };
                                if (val === "relation") {
                                  updates.props = {
                                    ...field.props,
                                    is_nullable: false,
                                    is_unique: false,
                                    default: [],
                                  };
                                }
                                updateField(index, updates);
                              }}
                            >
                              <SelectTrigger className="h-8 bg-background text-xs capitalize">
                                <SelectValue />
                              </SelectTrigger>
                              <SelectContent className="p-1">
                                {types.map((t: any) => {
                                  const cfg = getTypeConfig(t.type, t.label);
                                  return (
                                    <SelectItem
                                      key={t.type}
                                      value={t.type}
                                      disabled={t.type === "id"}
                                      className="capitalize"
                                    >
                                      <div className="flex items-center gap-2">
                                        <cfg.icon className={cn("h-3.5 w-3.5", cfg.color)} />
                                        {cfg.label}
                                      </div>
                                    </SelectItem>
                                  );
                                })}
                              </SelectContent>
                            </Select>

                            <ButtonGroup className="border-l pl-2 ml-1">
                              <Button
                                variant={isOpen ? "secondary" : "ghost"}
                                size="icon"
                                className="h-8 w-8"
                                onClick={() => toggleSettings(index)}
                              >
                                {isOpen ? <ChevronDown className="h-4 w-4" /> : <Settings2 className="h-4 w-4" />}
                              </Button>
                              {field.type !== "id" && (
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="h-8 w-8 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
                                  onClick={() => removeField(index)}
                                >
                                  <Trash2 className="h-4 w-4" />
                                </Button>
                              )}
                            </ButtonGroup>
                          </div>
                        </div>

                        {isOpen && (
                          <div className="p-4 space-y-0 animate-in slide-in-from-top-2 duration-300 divide-y border-t bg-muted/5">
                            {field.type === "id" && (
                              <PropertyRow
                                label="ID Length"
                                description="Length of the auto-generated ID string."
                                disableTopPadding
                                disableBottomPadding
                              >
                                <Input
                                  type="number"
                                  min={3}
                                  max={120}
                                  className="h-8 w-16 text-center font-bold text-xs"
                                  value={field.props.id?.auto_len ?? ""}
                                  onChange={(e) =>
                                    updateField(index, {
                                      props: {
                                        ...field.props,
                                        id: {
                                          ...field.props.id,
                                          auto_len: e.target.value === "" ? undefined : parseInt(e.target.value),
                                        },
                                      },
                                    })
                                  }
                                />
                              </PropertyRow>
                            )}

                            {field.type === "text" && (
                              <>
                                <div className="grid grid-cols-2 space-x-4 border-b pb-4">
                                  <div className="space-y-1">
                                    <Label className="text-xs font-semibold text-muted-foreground mb-1 block">
                                      Min Length
                                    </Label>
                                    <Input
                                      type="number"
                                      className="h-8 text-xs bg-background"
                                      placeholder="None"
                                      value={field.props.text?.min ?? ""}
                                      onChange={(e) =>
                                        updateField(index, {
                                          props: {
                                            ...field.props,
                                            text: {
                                              ...field.props.text,
                                              min: e.target.value === "" ? undefined : parseInt(e.target.value),
                                            },
                                          },
                                        })
                                      }
                                    />
                                  </div>

                                  <div className="space-y-1">
                                    <Label className="text-xs font-semibold text-muted-foreground mb-1 block">
                                      Max Length
                                    </Label>
                                    <Input
                                      type="number"
                                      className="h-8 text-xs bg-background"
                                      placeholder="None"
                                      value={field.props.text?.max ?? ""}
                                      onChange={(e) =>
                                        updateField(index, {
                                          props: {
                                            ...field.props,
                                            text: {
                                              ...field.props.text,
                                              max: e.target.value === "" ? undefined : parseInt(e.target.value),
                                            },
                                          },
                                        })
                                      }
                                    />
                                  </div>
                                </div>

                                <PropertyRow
                                  label="Default Value"
                                  description="Default value for new records."
                                  bottomBorder
                                >
                                  <Input
                                    type="text"
                                    placeholder={field.props.is_unique ? "None (Unique)" : "None"}
                                    className="h-8 w-32 bg-background text-xs"
                                    value={field.props.is_unique ? "" : field.props.default || ""}
                                    disabled={field.props.is_unique}
                                    onChange={(e) =>
                                      updateField(index, { props: { ...field.props, default: e.target.value } })
                                    }
                                  />
                                </PropertyRow>

                                <PropertyRow label="Nullable" description="Can this field be empty?" bottomBorder>
                                  <Switch
                                    checked={field.props.is_nullable ?? false}
                                    onCheckedChange={(v) =>
                                      updateField(index, { props: { ...field.props, is_nullable: v } })
                                    }
                                  />
                                </PropertyRow>

                                <PropertyRow label="Unique" description="Prevent duplicate values?" bottomBorder>
                                  <Switch
                                    checked={field.props.is_unique ?? false}
                                    onCheckedChange={(v) =>
                                      updateField(index, {
                                        props: {
                                          ...field.props,
                                          is_unique: v,
                                          default: v ? "" : field.props.default,
                                        },
                                      })
                                    }
                                  />
                                </PropertyRow>

                                <PropertyRow
                                  label="Hidden"
                                  description="Hide from default API response?"
                                  disableBottomPadding
                                >
                                  <Switch
                                    checked={field.props.is_hidden ?? false}
                                    onCheckedChange={(v) =>
                                      updateField(index, { props: { ...field.props, is_hidden: v } })
                                    }
                                  />
                                </PropertyRow>
                              </>
                            )}

                            {field.type === "number" && (
                              <>
                                <div className="grid grid-cols-2 space-x-4 border-b pb-4">
                                  <div className="space-y-1">
                                    <Label className="text-xs font-semibold text-muted-foreground mb-1 block">
                                      Min Value
                                    </Label>
                                    <Input
                                      type="number"
                                      step="any"
                                      className="h-8 text-xs bg-background"
                                      placeholder="None"
                                      value={field.props.number?.min ?? ""}
                                      onChange={(e) => {
                                        const val = e.target.value === "" ? undefined : parseFloat(e.target.value);
                                        updateField(index, {
                                          props: {
                                            ...field.props,
                                            number: { ...field.props.number, min: val },
                                          },
                                        });
                                      }}
                                    />
                                  </div>
                                  <div className="space-y-1">
                                    <Label className="text-xs font-semibold text-muted-foreground mb-1 block">
                                      Max Value
                                    </Label>
                                    <Input
                                      type="number"
                                      step="any"
                                      className="h-8 text-xs bg-background"
                                      placeholder="None"
                                      value={field.props.number?.max ?? ""}
                                      onChange={(e) => {
                                        const val = e.target.value === "" ? undefined : parseFloat(e.target.value);
                                        updateField(index, {
                                          props: {
                                            ...field.props,
                                            number: { ...field.props.number, max: val },
                                          },
                                        });
                                      }}
                                    />
                                  </div>
                                </div>

                                <PropertyRow
                                  label="Default Value"
                                  description="Default value for new records."
                                  bottomBorder
                                >
                                  <Input
                                    type="number"
                                    placeholder={field.props.is_unique ? "None (Unique)" : "None"}
                                    className="h-8 w-32 bg-background text-xs"
                                    value={field.props.is_unique ? "" : field.props.default || ""}
                                    disabled={field.props.is_unique}
                                    onChange={(e) =>
                                      updateField(index, { props: { ...field.props, default: e.target.value } })
                                    }
                                  />
                                </PropertyRow>

                                <PropertyRow
                                  label="No Decimals"
                                  description="Store as an integer (no fractional part)."
                                  bottomBorder
                                >
                                  <Switch
                                    checked={field.props.number?.no_decimals ?? false}
                                    onCheckedChange={(v) =>
                                      updateField(index, {
                                        props: {
                                          ...field.props,
                                          number: { ...field.props.number, no_decimals: v },
                                        },
                                      })
                                    }
                                  />
                                </PropertyRow>

                                <PropertyRow label="No Zero" description="Do not allow zero values." bottomBorder>
                                  <Switch
                                    checked={field.props.number?.no_zero ?? false}
                                    onCheckedChange={(v) =>
                                      updateField(index, {
                                        props: {
                                          ...field.props,
                                          number: { ...field.props.number, no_zero: v },
                                        },
                                      })
                                    }
                                  />
                                </PropertyRow>

                                <PropertyRow label="Nullable" description="Can this field be empty?" bottomBorder>
                                  <Switch
                                    checked={field.props.is_nullable ?? false}
                                    onCheckedChange={(v) =>
                                      updateField(index, { props: { ...field.props, is_nullable: v } })
                                    }
                                  />
                                </PropertyRow>

                                <PropertyRow label="Unique" description="Prevent duplicate values?" bottomBorder>
                                  <Switch
                                    checked={field.props.is_unique ?? false}
                                    onCheckedChange={(v) =>
                                      updateField(index, {
                                        props: {
                                          ...field.props,
                                          is_unique: v,
                                          default: v ? "" : field.props.default,
                                        },
                                      })
                                    }
                                  />
                                </PropertyRow>

                                <PropertyRow
                                  label="Hidden"
                                  description="Hide from default API response?"
                                  disableBottomPadding
                                >
                                  <Switch
                                    checked={field.props.is_hidden ?? false}
                                    onCheckedChange={(v) =>
                                      updateField(index, { props: { ...field.props, is_hidden: v } })
                                    }
                                  />
                                </PropertyRow>
                              </>
                            )}

                            {field.type === "boolean" && (
                              <>
                                <PropertyRow
                                  label="Initial State"
                                  description="Default boolean value for new records."
                                  bottomBorder
                                  disableTopPadding
                                >
                                  <Switch
                                    checked={field.props.default_bool ?? false}
                                    onCheckedChange={(v) =>
                                      updateField(index, { props: { ...field.props, default_bool: v } })
                                    }
                                  />
                                </PropertyRow>

                                <PropertyRow
                                  label="Hidden"
                                  description="Hide from default API response?"
                                  disableBottomPadding
                                >
                                  <Switch
                                    checked={field.props.is_hidden ?? false}
                                    onCheckedChange={(v) =>
                                      updateField(index, { props: { ...field.props, is_hidden: v } })
                                    }
                                  />
                                </PropertyRow>
                              </>
                            )}

                            {field.type === "json" && (
                              <>
                                <PropertyRow
                                  label="Nullable"
                                  description="Can this field be empty?"
                                  disableTopPadding
                                  bottomBorder
                                >
                                  <Switch
                                    checked={field.props.is_nullable ?? false}
                                    onCheckedChange={(v) =>
                                      updateField(index, { props: { ...field.props, is_nullable: v } })
                                    }
                                  />
                                </PropertyRow>

                                <PropertyRow label="Unique" description="Prevent duplicate values?" bottomBorder>
                                  <Switch
                                    checked={field.props.is_unique ?? false}
                                    onCheckedChange={(v) =>
                                      updateField(index, {
                                        props: {
                                          ...field.props,
                                          is_unique: v,
                                          default: v ? "" : field.props.default,
                                        },
                                      })
                                    }
                                  />
                                </PropertyRow>

                                <PropertyRow
                                  label="Hidden"
                                  description="Hide from default API response?"
                                  disableBottomPadding
                                >
                                  <Switch
                                    checked={field.props.is_hidden ?? false}
                                    onCheckedChange={(v) =>
                                      updateField(index, { props: { ...field.props, is_hidden: v } })
                                    }
                                  />
                                </PropertyRow>
                              </>
                            )}

                            {field.type === "date" && (
                              <>
                                <PropertyRow
                                  label="Auto Timestamp"
                                  description="Automatically set time on entry lifecycle."
                                  disableTopPadding
                                  bottomBorder
                                >
                                  <Select
                                    value={field.props.auto_timestamp || "none"}
                                    onValueChange={(v) =>
                                      updateField(index, {
                                        props: {
                                          ...field.props,
                                          auto_timestamp: v === "none" ? undefined : (v as any),
                                        },
                                      })
                                    }
                                  >
                                    <SelectTrigger>
                                      <SelectValue className="capitalize">
                                        {field.props.auto_timestamp === "create"
                                          ? "Create"
                                          : field.props.auto_timestamp === "create_update"
                                            ? "Create/Update"
                                            : "None"}
                                      </SelectValue>
                                    </SelectTrigger>
                                    <SelectContent className="p-1">
                                      <SelectItem value="none">None</SelectItem>
                                      <SelectItem value="create">Create</SelectItem>
                                      <SelectItem value="create_update">Create/Update</SelectItem>
                                    </SelectContent>
                                  </Select>
                                </PropertyRow>

                                <PropertyRow label="Nullable" description="Can this field be empty?" bottomBorder>
                                  <Switch
                                    checked={field.props.is_nullable ?? false}
                                    onCheckedChange={(v) =>
                                      updateField(index, { props: { ...field.props, is_nullable: v } })
                                    }
                                  />
                                </PropertyRow>

                                <PropertyRow
                                  label="Hidden"
                                  description="Hide from default API response?"
                                  disableBottomPadding
                                >
                                  <Switch
                                    checked={field.props.is_hidden ?? false}
                                    onCheckedChange={(v) =>
                                      updateField(index, { props: { ...field.props, is_hidden: v } })
                                    }
                                  />
                                </PropertyRow>
                              </>
                            )}

                            {field.type === "relation" && (
                              <>
                                <PropertyRow
                                  label="Target Collection"
                                  description="Choose which collection to link with."
                                  disableTopPadding
                                  bottomBorder
                                >
                                  <Select
                                    value={field.props.relation?.collection || ""}
                                    onValueChange={(v) =>
                                      updateField(index, {
                                        props: {
                                          ...field.props,
                                          relation: {
                                            collection: v as string,
                                            relation_mode: field.props.relation?.relation_mode,
                                            cascade_delete: field.props.relation?.cascade_delete,
                                          },
                                        },
                                      })
                                    }
                                  >
                                    <SelectTrigger>
                                      <SelectValue placeholder="Select..." className="capitalize" />
                                    </SelectTrigger>
                                    <SelectContent className="p-1">
                                      {collectionsList.map((c) => (
                                        <SelectItem key={c.name} value={c.name}>
                                          {c.name}
                                        </SelectItem>
                                      ))}
                                    </SelectContent>
                                  </Select>
                                </PropertyRow>

                                <PropertyRow
                                  label="Relation Mode"
                                  description="Single for one record, Multiple for many."
                                  bottomBorder
                                >
                                  <Select
                                    value={field.props.relation?.relation_mode || "single"}
                                    onValueChange={(v) =>
                                      updateField(index, {
                                        props: {
                                          ...field.props,
                                          relation: {
                                            collection: field.props.relation?.collection,
                                            relation_mode: v as any,
                                            cascade_delete: field.props.relation?.cascade_delete,
                                          },
                                        },
                                      })
                                    }
                                  >
                                    <SelectTrigger>
                                      <SelectValue className="capitalize" />
                                    </SelectTrigger>
                                    <SelectContent className="p-1">
                                      <SelectItem value="single" className="text-xs">
                                        Single
                                      </SelectItem>
                                      <SelectItem value="multiple" className="text-xs">
                                        Multiple
                                      </SelectItem>
                                    </SelectContent>
                                  </Select>
                                </PropertyRow>

                                <PropertyRow
                                  label="Cascade Delete"
                                  description="Delete linked records when this record is deleted."
                                  bottomBorder
                                >
                                  <Switch
                                    checked={field.props.relation?.cascade_delete ?? false}
                                    onCheckedChange={(v) =>
                                      updateField(index, {
                                        props: {
                                          ...field.props,
                                          relation: {
                                            collection: field.props.relation?.collection,
                                            relation_mode: field.props.relation?.relation_mode,
                                            cascade_delete: v,
                                          },
                                        },
                                      })
                                    }
                                  />
                                </PropertyRow>

                                <PropertyRow
                                  label="Hidden"
                                  description="Hide from default API response?"
                                  disableBottomPadding
                                >
                                  <Switch
                                    checked={field.props.is_hidden ?? false}
                                    onCheckedChange={(v) =>
                                      updateField(index, { props: { ...field.props, is_hidden: v } })
                                    }
                                  />
                                </PropertyRow>
                              </>
                            )}
                          </div>
                        )}
                      </div>
                    );
                  })}
                </div>
              </div>
            </div>
          </ScrollArea>

          <SheetFooter className="flex flex-row p-4 border-t justify-end">
            <Button variant="outline" size="lg" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button variant="default" size="lg" onClick={() => handleSubmit()} disabled={loading}>
              {loading ? "Processing..." : isEdit ? "Save Changes" : "Finish & Create"}
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>

      {/* Truncate Alert */}
      <AlertDialog open={showTruncateAlert} onOpenChange={setShowTruncateAlert}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Truncate Collection?</AlertDialogTitle>
            <AlertDialogDescription>
              This will permanently delete all records in &quot;<strong>{initialData?.name}</strong>&quot;. This action
              cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction variant="destructive" onClick={handleTruncate}>
              Clear All Data
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Delete Alert */}
      <AlertDialog open={showDeleteAlert} onOpenChange={setShowDeleteAlert}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Collection?</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete &quot;<strong>{initialData?.name}</strong>&quot;? The table and all its
              data will be permanently removed.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction variant="destructive" onClick={handleDelete}>
              Delete Permanently
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Save Schema Changes Confirmation Alert */}
      <AlertDialog open={showSaveAlert} onOpenChange={setShowSaveAlert}>
        <AlertDialogContent className="max-w-md">
          <AlertDialogHeader>
            <AlertDialogTitle>Review Changes</AlertDialogTitle>
            <AlertDialogDescription>
              You are about to modify the schema for &quot;<strong>{initialData?.name}</strong>&quot;. Review the
              changes carefully as they may affect existing data.
            </AlertDialogDescription>
          </AlertDialogHeader>

          <ScrollArea className="max-h-64 py-2">
            <div className="space-y-2">
              {changes.map((change, i) => (
                <div
                  key={i}
                  className={cn(
                    "flex items-start gap-3 p-2 rounded-lg border text-xs",
                    change.type === "added" && "bg-emerald-500/5 border-emerald-500/20",
                    change.type === "removed" && "bg-destructive/5 border-destructive/20",
                    change.type === "modified" && "bg-amber-500/5 border-amber-500/20",
                  )}
                >
                  <div
                    className={cn(
                      "mt-0.5 h-2 w-2 rounded-full shrink-0",
                      change.type === "added" && "bg-emerald-500",
                      change.type === "removed" && "bg-destructive",
                      change.type === "modified" && "bg-amber-500",
                    )}
                  />
                  <div className="flex-1 space-y-1">
                    <div className="flex items-center justify-between">
                      <span className="font-semibold text-xs tracking-tight"> {change.type}</span>
                      <span className="text-xs opacity-60">#{change.name}</span>
                    </div>
                    <div className="text-foreground/80 font-medium text-xs">
                      {change.type === "added"
                        ? `Add new field "${change.name}"`
                        : change.type === "removed"
                          ? `Drop field "${change.name}"`
                          : `Update field "${change.name}"`}
                    </div>
                    {change.details && <div className="text-xs opacity-60 leading-tight">{change.details}</div>}
                  </div>
                </div>
              ))}
            </div>
          </ScrollArea>

          <AlertDialogFooter className="pt-2">
            <AlertDialogCancel>Discard Changes</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => {
                setShowSaveAlert(false);
                handleSubmit(true);
              }}
            >
              Apply Changes
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
