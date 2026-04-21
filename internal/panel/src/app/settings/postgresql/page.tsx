"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Loader2, Save, Send, Folder, RefreshCw, Database, FileText } from "lucide-react";
import { toast } from "sonner";
import { fetchAPI } from "@/lib/api-client";
import { PageLoader } from "@/components/global/widget/loader";
import { DirectoryPickerSheet } from "@/components/global/directory-picker";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import { getPostgresPathPlaceholder, formatLabel } from "@/lib/utils";
import { POSTGRES_CONF_TEMPLATE, type ConfSection, type ConfField } from "@/lib/postgresql-template";

interface ConfigFieldRowProps {
  field: ConfField;
  sIndex: number;
  fIndex: number;
  onUpdate: (sIndex: number, fIndex: number, value: string) => void;
  onToggle: (sIndex: number, fIndex: number, enabled: boolean) => void;
  onPickPath: (sIndex: number, fIndex: number) => void;
}

const ConfigFieldRow = ({ field, sIndex, fIndex, onUpdate, onToggle, onPickPath }: ConfigFieldRowProps) => {
  const [localValue, setLocalValue] = useState(field.value);

  // Sync when field.value changes from outside (e.g. save)
  useEffect(() => {
    setLocalValue(field.value);
  }, [field.value]);

  return (
    <div className="group flex flex-col px-4 md:flex-row md:items-center justify-between gap-4 py-4 border-b border-border/40 last:border-0 hover:bg-muted/5 transition-colors px-1">
      <div className="flex items-start gap-4 flex-1">
        <div className="pt-1">
          <Switch checked={!field.is_commented} onCheckedChange={(checked) => onToggle(sIndex, fIndex, checked)} />
        </div>
        <div className="flex flex-col gap-1.5 flex-1">
          <div className="flex items-center gap-2">
            <span
              className={`text-sm font-bold tracking-tight transition-colors ${
                field.is_commented ? "text-muted-foreground/50" : "text-foreground"
              }`}
            >
              {formatLabel(field.name)}
            </span>
            {field.is_commented && (
              <Badge
                variant="secondary"
                className="text-[9px] h-4 px-1.5 font-bold uppercase tracking-tighter opacity-40"
              >
                Default
              </Badge>
            )}
          </div>
          <p className="text-xs text-muted-foreground leading-relaxed max-w-xl">{field.description}</p>
        </div>
      </div>
      <div
        className={`w-full md:w-64 transition-opacity duration-200 ${
          field.is_commented ? "opacity-30 pointer-events-none" : "opacity-100"
        }`}
      >
        {field.type === "select" ? (
          <Select value={localValue} onValueChange={(val) => onUpdate(sIndex, fIndex, val)}>
            <SelectTrigger className="h-10 text-sm font-mono border-none bg-muted/40 focus:bg-background transition-all">
              <SelectValue placeholder="Select option" />
            </SelectTrigger>
            <SelectContent>
              {field.options?.map((opt) => (
                <SelectItem key={opt} value={opt} className="text-xs font-mono">
                  {opt}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        ) : field.type === "toggle" ? (
          <div className="flex justify-start md:justify-end items-center h-10 px-2">
            <Switch
              checked={localValue === "on"}
              onCheckedChange={(checked) => {
                const val = checked ? "on" : "off";
                setLocalValue(val);
                onUpdate(sIndex, fIndex, val);
              }}
            />
          </div>
        ) : field.type === "path" ? (
          <div className="relative group/path">
            <Input
              value={localValue}
              onChange={(e) => setLocalValue(e.target.value)}
              onBlur={() => {
                if (localValue !== field.value) {
                  onUpdate(sIndex, fIndex, localValue);
                }
              }}
              className="h-10 pr-10 text-xs font-mono border-none bg-muted/40 focus:bg-background transition-all"
            />
            <button
              onClick={() => onPickPath(sIndex, fIndex)}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
            >
              <Folder className="h-4 w-4" />
            </button>
          </div>
        ) : (
          <Input
            type={field.type === "number" ? "number" : "text"}
            value={localValue}
            onChange={(e) => setLocalValue(e.target.value)}
            onBlur={() => {
              if (localValue !== field.value) {
                onUpdate(sIndex, fIndex, localValue);
              }
            }}
            className="h-10 text-sm font-mono border-none bg-muted/40 focus-visible:bg-background transition-all"
          />
        )}
      </div>
    </div>
  );
};

export default function PostgresSettingsPage() {
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [isTesting, setIsTesting] = useState(false);
  const [isRestarting, setIsRestarting] = useState(false);
  const [countdown, setCountdown] = useState(5);
  const [pickerTarget, setPickerTarget] = useState<"root" | "data" | { s: number; f: number } | null>(null);
  const [activeTab, setActiveTab] = useState("server");

  // Server Config
  const [config, setConfig] = useState({
    host: "",
    port: "",
    user: "",
    password: "",
    database: "",
    postgresPath: "",
    postgresDataPath: "",
  });

  // Conf File
  const [confPath, setConfPath] = useState("");
  const [sections, setSections] = useState<ConfSection[]>([]);

  useEffect(() => {
    const fetchConfig = async () => {
      try {
        const res = await fetchAPI("/settings/postgresql");
        if (res.success) {
          setConfig(res.data);
        }
      } catch (error) {
        toast.error("Failed to load PostgreSQL configuration");
      } finally {
        setIsLoading(false);
      }
    };
    fetchConfig();
  }, []);

  const fetchConfFile = async () => {
    try {
      const res = await fetchAPI("/settings/postgresql/conf");
      if (res.success) {
        setConfPath(res.data.path);

        // Merge API data with the client-side template
        const apiFields = new Map();
        res.data.sections.forEach((section: ConfSection) => {
          section.fields.forEach((field) => {
            apiFields.set(field.name, field);
          });
        });

        const mergedSections = POSTGRES_CONF_TEMPLATE.map((tSection) => ({
          ...tSection,
          fields: tSection.fields.map((tField) => {
            const apiField = apiFields.get(tField.name);
            if (apiField) {
              return {
                ...tField,
                value: apiField.value,
                is_commented: apiField.is_commented,
              };
            }
            return tField;
          }),
        }));

        setSections(mergedSections);
      } else {
        toast.error(res.error?.message || "Failed to load postgresql.conf");
      }
    } catch (error) {
      toast.error("An error occurred while loading config file");
    }
  };

  useEffect(() => {
    if (activeTab === "config" && sections.length === 0) {
      fetchConfFile();
    }
  }, [activeTab]);

  useEffect(() => {
    let timer: any;
    if (isRestarting && countdown > 0) {
      timer = setTimeout(() => setCountdown(countdown - 1), 1000);
    } else if (isRestarting && countdown === 0) {
      window.location.reload();
    }
    return () => clearTimeout(timer);
  }, [isRestarting, countdown]);

  const handleUpdateConfField = (sectionIndex: number, fieldIndex: number, newValue: string) => {
    setSections((prev) => {
      const updated = [...prev];
      updated[sectionIndex].fields[fieldIndex].value = newValue;
      updated[sectionIndex].fields[fieldIndex].is_commented = false;
      return updated;
    });
  };

  const handleToggleConfStatus = (sectionIndex: number, fieldIndex: number, enabled: boolean) => {
    setSections((prev) => {
      const updated = [...prev];
      updated[sectionIndex].fields[fieldIndex].is_commented = !enabled;
      return updated;
    });
  };

  const handleSaveConf = async () => {
    setIsSaving(true);
    try {
      const res = await fetchAPI("/settings/postgresql/conf", {
        method: "PATCH",
        body: JSON.stringify({ path: confPath, sections: sections }),
      });
      if (res.success) {
        toast.success("postgresql.conf updated! Restarting engine...");
        try {
          await fetchAPI("/system/engine/restart", { method: "POST" });
          setIsRestarting(true);
        } catch (e) {
          setIsRestarting(true);
        }
      }
    } catch (e) {
      toast.error("Failed to save configuration");
    } finally {
      setIsSaving(false);
    }
  };

  const handleTestConnection = async () => {
    setIsTesting(true);
    try {
      const res = await fetchAPI("/settings/postgresql/test", {
        method: "POST",
        body: JSON.stringify(config),
      });
      if (res.success) {
        toast.success("Connection successful!");
      } else {
        toast.error(res.error?.message || "Connection failed");
      }
    } catch (error) {
      toast.error("An error occurred while testing connection");
    } finally {
      setIsTesting(false);
    }
  };

  const handleSave = async () => {
    if (!config.postgresPath || !config.postgresDataPath) {
      toast.error("Required Fields Missing", {
        description: "PostgreSQL Root Path and Data Path are mandatory.",
      });
      return;
    }

    setIsSaving(true);
    try {
      const res = await fetchAPI("/settings/postgresql", {
        method: "PATCH",
        body: JSON.stringify(config),
      });
      if (res.success) {
        toast.success("Settings saved! Initiating engine restart...");
        try {
          await fetchAPI("/system/engine/restart", { method: "POST" });
          setIsRestarting(true);
        } catch (e) {
          setIsRestarting(true);
        }
      }
    } catch (error) {
      toast.error("An error occurred while saving settings");
    } finally {
      setIsSaving(false);
    }
  };

  if (isLoading) {
    return (
      <PageLoader
        title="Loading Database Config"
        description="Retrieving secure connection parameters from the engine..."
      />
    );
  }

  if (isRestarting) {
    return (
      <div className="flex flex-col items-center justify-center py-32">
        <div className="mb-8">
          <RefreshCw className="h-12 w-12 text-primary animate-spin" />
        </div>
        <div className="text-center space-y-2 mb-8">
          <h2 className="text-2xl font-bold tracking-tight">Engine Restarting</h2>
          <p className="text-muted-foreground max-w-sm mx-auto">
            Please wait while Blinky applies your custom PostgreSQL configurations and refreshes the system.
          </p>
        </div>
        <div className="flex items-center gap-2 px-4 py-2 bg-muted rounded-full border">
          <span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Reconnecting in</span>
          <div className="h-6 w-6 rounded-full bg-primary flex items-center justify-center text-[10px] font-bold text-primary-foreground">
            {countdown}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-8 max-w-6xl">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-6">
        <div className="flex flex-col gap-1">
          <h1 className="text-2xl font-bold tracking-tight">PostgreSQL Configuration</h1>
          <p className="text-sm text-muted-foreground">Manage your core database connection and binary tool paths.</p>
        </div>
        <Tabs value={activeTab} onValueChange={setActiveTab} className="w-auto">
          <TabsList className="bg-muted/50 border">
            <TabsTrigger value="server" className="gap-2 px-4">
              <Database className="h-4 w-4" />
              Server
            </TabsTrigger>
            <TabsTrigger value="config" className="gap-2 px-4">
              <FileText className="h-4 w-4" />
              Config
            </TabsTrigger>
          </TabsList>
        </Tabs>
      </div>

      <Tabs value={activeTab} className="w-full">
        <TabsContent value="server" className="mt-0 space-y-8 animate-in fade-in slide-in-from-bottom-1 duration-300">
          <div className="grid gap-6 max-w-4xl">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="space-y-2">
                <Label htmlFor="host">Host</Label>
                <Input id="host" value={config.host} onChange={(e) => setConfig({ ...config, host: e.target.value })} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="port">Port</Label>
                <Input id="port" value={config.port} onChange={(e) => setConfig({ ...config, port: e.target.value })} />
              </div>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="space-y-2">
                <Label htmlFor="user">Database User</Label>
                <Input id="user" value={config.user} onChange={(e) => setConfig({ ...config, user: e.target.value })} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="password">Password</Label>
                <Input
                  id="password"
                  type="password"
                  value={config.password}
                  onChange={(e) => setConfig({ ...config, password: e.target.value })}
                />
              </div>
            </div>
            <div className="space-y-2">
              <Label htmlFor="database">Database Name</Label>
              <Input
                id="database"
                value={config.database}
                onChange={(e) => setConfig({ ...config, database: e.target.value })}
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 pt-6 border-t font-mono">
              <div className="space-y-3">
                <div className="space-y-1">
                  <Label className="text-base">PostgreSQL Root Path</Label>
                  <p className="text-xs text-muted-foreground">
                    Installation directory where binary tools (/bin) reside.
                  </p>
                </div>
                <div className="relative">
                  <Input
                    value={config.postgresPath}
                    onChange={(e) => setConfig({ ...config, postgresPath: e.target.value })}
                    placeholder={getPostgresPathPlaceholder()}
                    className="pr-10 text-xs h-9"
                  />
                  <button
                    onClick={() => setPickerTarget("root")}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                  >
                    <Folder className="h-3.5 w-3.5" />
                  </button>
                </div>
              </div>

              <div className="space-y-3">
                <div className="space-y-1">
                  <Label className="text-base">PostgreSQL Data Path</Label>
                  <p className="text-xs text-muted-foreground">
                    Variable directory where configuration and databases are stored.
                  </p>
                </div>
                <div className="relative">
                  <Input
                    value={config.postgresDataPath}
                    onChange={(e) => setConfig({ ...config, postgresDataPath: e.target.value })}
                    placeholder={getPostgresPathPlaceholder()}
                    className="pr-10 text-xs h-9"
                  />
                  <button
                    onClick={() => setPickerTarget("data")}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                  >
                    <Folder className="h-3.5 w-3.5" />
                  </button>
                </div>
              </div>
            </div>

            <div className="flex flex-col sm:flex-row gap-3 pt-6 border-t">
              <Button variant="outline" onClick={handleTestConnection} disabled={isTesting}>
                {isTesting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Send className="mr-2 h-4 w-4" />}
                Test Connection
              </Button>
              <Button onClick={handleSave} disabled={isSaving}>
                {isSaving ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Save className="mr-2 h-4 w-4" />}
                Save Configuration
              </Button>
            </div>
          </div>
        </TabsContent>

        <TabsContent value="config" className="mt-0 space-y-8 animate-in fade-in slide-in-from-bottom-1 duration-300">
          <div className="flex flex-col gap-8 pb-24">
            <div className="fixed bottom-8 right-8 z-50">
              <Button
                onClick={handleSaveConf}
                disabled={isSaving}
                size="lg"
                className="h-14 px-8 rounded-full gap-3 hover:scale-105 active:scale-95 transition-all bg-primary hover:bg-primary/90 text-primary-foreground group"
              >
                {isSaving ? (
                  <Loader2 className="h-5 w-5 animate-spin" />
                ) : (
                  <RefreshCw className="h-5 w-5 group-hover:rotate-180 transition-transform duration-500" />
                )}
                <div className="flex flex-col items-start leading-none gap-1">
                  <span className="text-xs font-bold uppercase tracking-widest opacity-70">Apply Changes</span>
                  <span className="text-sm font-black">Save & Restart</span>
                </div>
              </Button>
            </div>

            <Accordion type="multiple">
              {sections.map((section, sIndex) => (
                <AccordionItem key={sIndex} value={`item-${sIndex}`}>
                  <AccordionTrigger className="hover:no-underline group">
                    <div className="flex items-center gap-4 w-full">
                      <h3 className="text-sm font-bold tracking-[0.2em] group-data-[state=open]:text-primary transition-colors">
                        {section.title}
                      </h3>
                      <div className="h-px flex-1 bg-linear-to-r from-border/50 via-border/5 to-transparent group-data-[state=open]:via-primary/20 transition-all" />
                    </div>
                  </AccordionTrigger>
                  <AccordionContent className="border-none">
                    <div className="grid gap-0">
                      {section.fields.map((field, fIndex) => (
                        <ConfigFieldRow
                          key={`${sIndex}-${fIndex}-${field.name}`}
                          field={field}
                          sIndex={sIndex}
                          fIndex={fIndex}
                          onUpdate={handleUpdateConfField}
                          onToggle={handleToggleConfStatus}
                          onPickPath={(s, f) => setPickerTarget({ s, f })}
                        />
                      ))}
                    </div>
                  </AccordionContent>
                </AccordionItem>
              ))}
            </Accordion>
          </div>
        </TabsContent>
      </Tabs>

      <DirectoryPickerSheet
        open={pickerTarget !== null}
        onOpenChange={(open) => !open && setPickerTarget(null)}
        onSelect={(path) => {
          if (pickerTarget === "root") setConfig({ ...config, postgresPath: path });
          else if (pickerTarget === "data") setConfig({ ...config, postgresDataPath: path });
          else if (typeof pickerTarget === "object" && pickerTarget !== null) {
            handleUpdateConfField(pickerTarget.s, pickerTarget.f, path);
          }
          setPickerTarget(null);
        }}
        initialPath={
          pickerTarget === "root"
            ? config.postgresPath
            : pickerTarget === "data"
              ? config.postgresDataPath
              : typeof pickerTarget === "object" && pickerTarget !== null
                ? sections[pickerTarget.s].fields[pickerTarget.f].value
                : undefined
        }
      />
    </div>
  );
}
