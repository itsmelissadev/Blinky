"use client";

import * as React from "react";
import { cn } from "@/lib/utils";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import { CalendarIcon, Type, Hash, ToggleLeft, Braces, Link2 } from "lucide-react";
import { format } from "date-fns";
import { Calendar } from "@/components/ui/calendar";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { X } from "lucide-react";

const typeConfig: Record<string, { icon: any; label: string; color: string }> = {
  text: { icon: Type, label: "Text", color: "text-blue-500" },
  number: { icon: Hash, label: "Number", color: "text-emerald-500" },
  boolean: { icon: ToggleLeft, label: "Boolean", color: "text-amber-500" },
  date: { icon: CalendarIcon, label: "Date", color: "text-violet-500" },
  json: { icon: Braces, label: "JSON", color: "text-orange-500" },
  relation: { icon: Link2, label: "Relation", color: "text-rose-500" },
};

interface BaseFieldProps {
  label: string;
  type: string;
  className?: string;
  children: React.ReactNode;
  hint?: string;
}

function FieldRow({ label, type, hint, className, children }: BaseFieldProps) {
  const config = typeConfig[type] || typeConfig.text;
  const Icon = config.icon;

  return (
    <div className={cn("space-y-2 pb-4 border-b last:border-b-0 last:pb-0", className)}>
      <div className="flex items-center gap-2">
        <div className={cn(
          "flex h-6 w-6 shrink-0 items-center justify-center rounded-md border",
          config.color,
        )}>
          <Icon className="h-3.5 w-3.5" />
        </div>
        <div className="flex-1 flex items-center justify-between gap-2">
          <Label className="text-sm font-semibold leading-none">{label}</Label>
          <span className="text-xs font-medium text-muted-foreground/60">
            {config.label}
          </span>
        </div>
      </div>
      {hint && <p className="text-xs text-muted-foreground pl-8">{hint}</p>}
      <div className="pl-8">{children}</div>
    </div>
  );
}

export function TextField({ label, type, register, ...props }: any) {
  return (
    <FieldRow label={label} type={type}>
      <Input
        type="text"
        {...register}
        {...props}
        className="font-medium"
      />
    </FieldRow>
  );
}

export function NumberField({ label, type, register, ...props }: any) {
  return (
    <FieldRow label={label} type={type}>
      <Input
        type="number"
        {...register}
        {...props}
        className="font-medium"
      />
    </FieldRow>
  );
}

export function BooleanField({ label, type, value, onCheckedChange }: any) {
  const config = typeConfig.boolean;
  const Icon = config.icon;

  return (
    <div className="flex items-center justify-between pb-4 border-b last:border-b-0 last:pb-0">
      <div className="flex items-center gap-2">
        <div className={cn(
          "flex h-6 w-6 shrink-0 items-center justify-center rounded-md border",
          config.color,
        )}>
          <Icon className="h-3.5 w-3.5" />
        </div>
        <div className="space-y-0.5">
          <Label className="text-sm font-semibold leading-none">{label}</Label>
          <p className="text-xs text-muted-foreground leading-none">
            {value ? "Enabled" : "Disabled"}
          </p>
        </div>
      </div>
      <Switch checked={value} onCheckedChange={onCheckedChange} />
    </div>
  );
}

export function DateField({ label, type, value, onSelect, getValues, setValue, name }: any) {
  return (
    <FieldRow label={label} type={type}>
      <Popover>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            className={cn(
              "w-full justify-start text-left font-medium",
              !value && "text-muted-foreground",
            )}
          >
            <CalendarIcon className="mr-2 h-4 w-4 opacity-50" />
            {value ? format(new Date(value), "PPP HH:mm") : "Pick a date & time"}
          </Button>
        </PopoverTrigger>
        <PopoverContent className="max-w-max p-0 flex flex-col shadow-xl border-muted" align="start">
          <Calendar
            className="w-full"
            mode="single"
            selected={value ? new Date(value) : undefined}
            onSelect={onSelect}
            initialFocus
          />
          <div className="p-3 border-t flex items-center gap-3">
            <div className="flex flex-col gap-1 flex-1">
              <span className="text-xs font-semibold text-muted-foreground px-1">Hour</span>
              <Input
                type="number"
                min={0}
                max={23}
                className="text-center text-xs"
                value={value ? new Date(value).getHours() : 12}
                onChange={(e) => {
                  const currentVal = getValues(name);
                  const d = currentVal ? new Date(currentVal) : new Date();
                  d.setHours(parseInt(e.target.value) || 0);
                  setValue(name, d.toISOString());
                }}
              />
            </div>
            <span className="text-muted-foreground mt-6">:</span>
            <div className="flex flex-col gap-1 flex-1">
              <span className="text-xs font-semibold text-muted-foreground px-1">Min</span>
              <Input
                type="number"
                min={0}
                max={59}
                className="text-center text-xs"
                value={value ? new Date(value).getMinutes() : 0}
                onChange={(e) => {
                  const currentVal = getValues(name);
                  const d = currentVal ? new Date(currentVal) : new Date();
                  d.setMinutes(parseInt(e.target.value) || 0);
                  setValue(name, d.toISOString());
                }}
              />
            </div>
          </div>
        </PopoverContent>
      </Popover>
    </FieldRow>
  );
}

export function JsonField({ label, type, register, ...props }: any) {
  return (
    <FieldRow label={label} type={type} hint="Valid JSON required">
      <Textarea
        {...register}
        {...props}
        className="font-mono text-sm! min-h-[100px] px-3 py-2"
        placeholder="{}"
      />
    </FieldRow>
  );
}

export function RelationField({ label, type, value, onChange, options, multiple, targetCollection }: any) {
  const selectedIDs = Array.isArray(value) ? value : value ? [value] : [];

  const handleSelect = (id: string) => {
    if (multiple) {
      if (selectedIDs.includes(id)) {
        onChange(selectedIDs.filter((i: string) => i !== id));
      } else {
        onChange([...selectedIDs, id]);
      }
    } else {
      onChange(id);
    }
  };

  return (
    <FieldRow label={label} type={type} hint={`Link to ${targetCollection}`}>
      <div className="space-y-2">
        {multiple && selectedIDs.length > 0 && (
          <div className="flex flex-wrap gap-1.5">
            {selectedIDs.map((id: string) => (
              <Badge
                key={id}
                variant="secondary"
                className="pl-2 pr-1 h-6 gap-1 text-xs font-medium"
              >
                {id}
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-4 w-4 p-0 hover:bg-transparent"
                  onClick={() => onChange(selectedIDs.filter((i: string) => i !== id))}
                >
                  <X className="h-3 w-3" />
                </Button>
              </Badge>
            ))}
          </div>
        )}

        <Select onValueChange={handleSelect} value={!multiple ? value || "" : ""}>
          <SelectTrigger className="text-sm font-medium py-4 px-3">
            <SelectValue placeholder={multiple ? "Add related records..." : (selectedIDs[0] || "Select a record...")} />
          </SelectTrigger>
          <SelectContent className="max-h-[200px]">
            {options.length === 0 && <div className="p-4 text-center text-xs text-muted-foreground">No records found.</div>}
            {options.map((opt: any) => (
              <SelectItem key={opt.id} value={opt.id} className="text-xs font-medium">
                <div className="flex flex-col">
                  <span className="font-semibold">{opt.id}</span>
                  {Object.keys(opt).filter(k => k !== 'id').slice(0, 1).map(k => (
                    <span key={k} className="text-xs text-muted-foreground/60 truncate max-w-[200px]">
                      {k}: {String(opt[k])}
                    </span>
                  ))}
                </div>
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
    </FieldRow>
  );
}
