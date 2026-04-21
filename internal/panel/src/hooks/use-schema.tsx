"use client";

import * as React from "react";
import { fetchAPI } from "@/lib/api-client";
import { useAuth } from "@/components/auth-guard";
import { toast } from "sonner";
import { Key, Type, Hash, ToggleLeft, Code as CodeIcon, Calendar, Database, LucideIcon, Link2 } from "lucide-react";

export type FieldType = "id" | "text" | "number" | "boolean" | "json" | "date" | "relation";

export interface FieldProps {
  id?: { auto_len?: number };
  is_nullable?: boolean;
  is_unique?: boolean;
  is_hidden?: boolean;
  default?: any;
  default_bool?: boolean;
  default_now?: boolean;
  auto_timestamp?: "create" | "create_update";
  text?: { min?: number; max?: number; regex?: string };
  number?: { min?: number; max?: number; no_zero?: boolean; no_decimals?: boolean };
  relation?: { collection?: string; relation_mode?: "single" | "multiple"; cascade_delete?: boolean };
}

export interface FieldDefinition {
  id: string;
  name: string;
  type: FieldType;
  props: FieldProps;
}

export interface FieldTypeInfo {
  type: FieldType;
  label: string;
}

export interface SchemaTypeMetadata {
  label: string;
  icon: LucideIcon;
  color: string;
}

const FIELD_VISUAL_MAP: Record<string, { icon: LucideIcon; color: string }> = {
  id: { icon: Key, color: "text-amber-500" },
  text: { icon: Type, color: "text-blue-500" },
  number: { icon: Hash, color: "text-emerald-500" },
  boolean: { icon: ToggleLeft, color: "text-rose-500" },
  json: { icon: CodeIcon, color: "text-purple-500" },
  date: { icon: Calendar, color: "text-orange-500" },
  relation: { icon: Link2, color: "text-indigo-500" },
};

interface SchemaContextType {
  types: FieldTypeInfo[];
  loading: boolean;
  getTypeConfig: (type: string, label?: string) => SchemaTypeMetadata;
}

const SchemaContext = React.createContext<SchemaContextType | undefined>(undefined);

export function SchemaProvider({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth();
  const [types, setTypes] = React.useState<FieldTypeInfo[]>([]);
  const [loading, setLoading] = React.useState(true);

  React.useEffect(() => {
    if (!isAuthenticated) {
      setLoading(false);
      return;
    }

    setLoading(true);
    fetchAPI("/collections/types")
      .then((json) => {
        if (json.success) setTypes(json.data);
      })
      .catch(() => toast.error("Failed to load schema definitions"))
      .finally(() => setLoading(false));
  }, [isAuthenticated]);

  const getTypeConfig = React.useCallback(
    (type: string, label?: string): SchemaTypeMetadata => {
      const meta = FIELD_VISUAL_MAP[type] || { icon: Database, color: "text-muted-foreground" };
      const typeInfo = types.find((t) => t.type === type);
      return {
        label: label || typeInfo?.label || type,
        icon: meta.icon,
        color: meta.color,
      };
    },
    [types],
  );

  return <SchemaContext.Provider value={{ types, loading, getTypeConfig }}>{children}</SchemaContext.Provider>;
}

export function useSchema() {
  const context = React.useContext(SchemaContext);
  if (context === undefined) {
    throw new Error("useSchema must be used within a SchemaProvider");
  }
  return context;
}
