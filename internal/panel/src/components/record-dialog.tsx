"use client";

import * as React from "react";
import { useForm, useWatch, type Control } from "react-hook-form";
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription, SheetFooter } from "@/components/ui/sheet";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Button } from "@/components/ui/button";
import { fetchAPI } from "@/lib/api-client";
import { toast } from "sonner";
import { TextField, NumberField, BooleanField, DateField, JsonField, RelationField } from "@/components/ui/form-fields";

interface RecordDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  collectionName: string;
  schema: any[];
  onSuccess?: () => void;
  recordToEdit?: any | null;
}

export function RecordDialog({
  open,
  onOpenChange,
  collectionName,
  schema,
  onSuccess,
  recordToEdit,
}: RecordDialogProps) {
  const { register, handleSubmit, reset, setValue, control, getValues } = useForm();
  const [loading, setLoading] = React.useState(false);

  // Reset form when dialog opens
  React.useEffect(() => {
    if (open) {
      if (recordToEdit) {
        // Prepare relations to be strings (or arrays of strings) if they come as full objects
        const formattedData = { ...recordToEdit };
        schema?.forEach((field) => {
          if (field.type === "relation" && formattedData[field.name]) {
            if (Array.isArray(formattedData[field.name])) {
              formattedData[field.name] = formattedData[field.name].map((r: any) => (typeof r === "object" ? r.id : r));
            } else if (typeof formattedData[field.name] === "object") {
              formattedData[field.name] = formattedData[field.name].id;
            }
          }
          if (
            field.type === "json" &&
            typeof formattedData[field.name] === "object" &&
            formattedData[field.name] !== null
          ) {
            formattedData[field.name] = JSON.stringify(formattedData[field.name], null, 2);
          }
        });
        reset(formattedData);
      } else {
        reset({});
        // Set defaults for booleans
        schema?.forEach((field) => {
          if (field.type === "boolean") {
            setValue(field.name, false);
          }
        });
      }
    }
  }, [open, reset, schema, setValue, recordToEdit]);

  const onSubmit = async (data: any) => {
    setLoading(true);
    try {
      // Validate and Clean data based on schema props
      const payload: any = {};
      for (const field of schema) {
        const props = field.props || {};

        // Skip fields that are handled by the server or marked as hidden
        if (field.name === "id") continue;
        if (field.type === "date" && (props.auto_timestamp === "create" || props.auto_timestamp === "create_update"))
          continue;

        let val = data[field.name];

        // 2. Required Check (is_nullable) - Always skip for 'id' as it's server generated
        if (props.is_nullable === false && field.name !== "id") {
          if (val === undefined || val === null || val === "" || (field.type === "bool" && val === undefined)) {
            toast.error(`Field "${field.name}" is required`);
            setLoading(false);
            return;
          }
        }

        // 2. Type Specific Validation and Cleaning
        if (field.type === "number") {
          const num = val === "" || val === undefined ? null : Number(val);
          if (num !== null) {
            if (props.number?.min !== undefined && num < props.number.min) {
              toast.error(`"${field.name}" must be at least ${props.number.min}`);
              setLoading(false);
              return;
            }
            if (props.number?.max !== undefined && num > props.number.max) {
              toast.error(`"${field.name}" must be at most ${props.number.max}`);
              setLoading(false);
              return;
            }
          }
          val = num;
        } else if (field.type === "text") {
          if (val) {
            if (props.text?.min !== undefined && val.length < props.text.min) {
              toast.error(`"${field.name}" too short (min ${props.text.min})`);
              setLoading(false);
              return;
            }
            if (props.text?.max !== undefined && val.length > props.text.max) {
              toast.error(`"${field.name}" too long (max ${props.text.max})`);
              setLoading(false);
              return;
            }
          }
        } else if (field.type === "json") {
          if (val) {
            try {
              val = typeof val === "string" ? JSON.parse(val) : val;
            } catch {
              toast.error(`"${field.name}" has invalid JSON`);
              setLoading(false);
              return;
            }
          } else {
            val = null;
          }
        }

        payload[field.name] = val;
      }

      const url = recordToEdit
        ? `/collection/${collectionName}/records/${recordToEdit.id}`
        : `/collection/${collectionName}/records`;

      const method = recordToEdit ? "PATCH" : "POST";

      const res = await fetchAPI(url, {
        method: method,
        body: JSON.stringify(payload),
      });

      if (res.success) {
        toast.success(recordToEdit ? "Record updated successfully" : "Record created successfully");
        onOpenChange(false);
        if (onSuccess) onSuccess();
      } else {
        toast.error(res.error?.message || `Failed to ${recordToEdit ? "update" : "create"} record`);
      }
    } catch (err: any) {
      toast.error(err.message || "An error occurred");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="max-sm:w-full! max-sm:inset-0! md:max-w-xl! flex flex-col p-0 gap-0 h-full bg-background border-l shadow-2xl">
        <div tabIndex={0} className="sr-only" aria-hidden="true" />
        <SheetHeader className="p-4 border-b gap-0 relative">
          <SheetTitle>{recordToEdit ? "Edit Record" : "New Record"}</SheetTitle>
          <SheetDescription>
            {recordToEdit ? "Update the entry in the " : "Add a new entry to the "} <strong>{collectionName}</strong>{" "}
            collection.
          </SheetDescription>
        </SheetHeader>

        <form onSubmit={handleSubmit(onSubmit)} className="flex-1 flex flex-col min-h-0">
          <ScrollArea className="flex-1 min-h-0 w-full">
            <div className="p-4 space-y-4">
              {schema
                ?.filter((f) => {
                  if (f.name === "id") return false;
                  if (f.props?.is_hidden) return false;
                  if (
                    f.type === "date" &&
                    (f.props?.auto_timestamp === "create" || f.props?.auto_timestamp === "create_update")
                  )
                    return false;
                  return true;
                })
                .map((field) => (
                  <FormField
                    key={field.id}
                    field={field}
                    control={control}
                    register={register}
                    setValue={setValue}
                    getValues={getValues}
                  />
                ))}
            </div>
          </ScrollArea>

          <SheetFooter className="flex flex-row p-4 border-t justify-end">
            <Button type="button" variant="outline" size="lg" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" size="lg" disabled={loading}>
              {loading ? "Processing..." : recordToEdit ? "Save Changes" : "Create Record"}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  );
}

function FormField({
  field,
  control,
  register,
  setValue,
  getValues,
}: {
  field: any;
  control: Control<any>;
  register: any;
  setValue: any;
  getValues: any;
}) {
  const value = useWatch({
    control,
    name: field.name,
  });

  if (field.type === "boolean") {
    return (
      <BooleanField
        label={field.name}
        type={field.type}
        value={value}
        onCheckedChange={(checked: boolean) => setValue(field.name, checked)}
      />
    );
  }

  if (field.type === "date") {
    return (
      <DateField
        label={field.name}
        type={field.type}
        value={value}
        onSelect={(date: Date) => {
          if (!date) return;
          const currentVal = getValues(field.name);
          const current = currentVal ? new Date(currentVal) : new Date();
          date.setHours(current.getHours());
          date.setMinutes(current.getMinutes());
          setValue(field.name, date.toISOString());
        }}
        getValues={getValues}
        setValue={setValue}
        name={field.name}
      />
    );
  }

  if (field.type === "number") {
    return (
      <NumberField
        label={field.name}
        type={field.type}
        placeholder={`Enter ${field.name}...`}
        register={register(field.name)}
        min={field.props?.number?.min}
        max={field.props?.number?.max}
        required={!field.props?.is_nullable}
      />
    );
  }

  if (field.type === "json") {
    return (
      <JsonField
        label={field.name}
        type={field.type}
        placeholder="{}"
        register={register(field.name)}
        required={!field.props?.is_nullable}
      />
    );
  }

  if (field.type === "relation") {
    return <RelationFieldSelector field={field} value={value} setValue={setValue} />;
  }

  return (
    <TextField
      label={field.name}
      type={field.type}
      placeholder={`Enter ${field.name}...`}
      register={register(field.name)}
      minLength={field.props?.text?.min}
      maxLength={field.props?.text?.max}
      required={!field.props?.is_nullable}
    />
  );
}

function RelationFieldSelector({ field, value, setValue }: { field: any; value: any; setValue: any }) {
  const [options, setOptions] = React.useState<any[]>([]);
  const target = field.props?.relation?.collection;

  React.useEffect(() => {
    if (target) {
      fetchAPI(`/collection/${target}/records`).then((res) => {
        if (res.success) setOptions(res.data || []);
      });
    }
  }, [target]);

  return (
    <RelationField
      label={field.name}
      type={field.type}
      value={value}
      onChange={(v: any) => setValue(field.name, v)}
      options={options}
      multiple={field.props?.relation?.relation_mode === "multiple"}
      targetCollection={target}
    />
  );
}
