"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Loader2,
  Save,
  RefreshCw,
  Server,
  AlertTriangle,
  ShieldCheck,
  AlertCircle,
  Lock,
  Shield,
  Eye,
  EyeOff,
} from "lucide-react";
import { toast } from "sonner";
import { fetchAPI } from "@/lib/api-client";
import { Switch } from "@/components/ui/switch";
import { PageLoader } from "@/components/global/widget/loader";
import { InputPassword } from "@/components/ui/input-password";

export default function ServerSettingsPage() {
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [isRestarting, setIsRestarting] = useState(false);
  const [countdown, setCountdown] = useState(5);
  const [isSshVerified, setIsSshVerified] = useState(false);

  const [config, setConfig] = useState({
    publicApiHost: "",
    publicApiPort: "",
    adminPanelHost: "",
    adminPanelPort: "",
    adminSshEnabled: false,
    publicSshEnabled: false,
    sshPort: "",
    sshUser: "",
    sshPassword: "",
  });

  useEffect(() => {
    const fetchConfig = async () => {
      try {
        const res = await fetchAPI("/settings/server");
        if (res.success) {
          setConfig(res.data);
          if (res.data.sshPort && res.data.sshUser && res.data.sshPassword) {
            setIsSshVerified(true);
          }
        }
      } catch (error) {
        toast.error("Failed to load server configuration");
      } finally {
        setIsLoading(false);
      }
    };
    fetchConfig();
  }, []);

  useEffect(() => {
    let timer: any;
    if (isRestarting && countdown > 0) {
      timer = setTimeout(() => setCountdown(countdown - 1), 1000);
    } else if (isRestarting && countdown === 0) {
      window.location.reload();
    }
    return () => clearTimeout(timer);
  }, [isRestarting, countdown]);

  const handleSave = async () => {
    const missingFields = [];
    if (!config.adminSshEnabled && !config.adminPanelHost) missingFields.push("Admin Panel Host");
    if (!config.publicSshEnabled && !config.publicApiHost) missingFields.push("Public API Host");
    if (!config.adminPanelPort) missingFields.push("Admin Panel Port");
    if (!config.publicApiPort) missingFields.push("Public API Port");

    if (missingFields.length > 0) {
      toast.error("Required Fields Missing", {
        description: `Please fill: ${missingFields.join(", ")}`,
      });
      return;
    }

    setIsSaving(true);
    try {
      const res = await fetchAPI("/settings/server", {
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
        title="Loading Server Config"
        description="Retrieving server and network parameters from the engine..."
      />
    );
  }

  if (isRestarting) {
    return (
      <div className="flex flex-col items-center justify-center py-32 animate-in fade-in zoom-in duration-500">
        <div className="mb-8 relative">
          <div className="absolute inset-0 bg-primary/20 rounded-full blur-xl animate-pulse" />
          <RefreshCw className="relative h-12 w-12 text-primary animate-spin" />
        </div>
        <div className="text-center space-y-2 mb-8">
          <h2 className="text-2xl font-bold tracking-tight">Engine Restarting</h2>
          <p className="text-muted-foreground max-w-sm mx-auto">
            Please wait while Blinky applies your custom server configurations and refreshes the system network
            bindings.
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

  const renderHostWarning = (host: string, isAdmin: boolean, isSsh?: boolean) => {
    if (!host) return null;

    if (isSsh) {
      return (
        <div className="flex items-center gap-2 mt-2 text-emerald-500 bg-emerald-500/10 px-2 py-1.5 rounded-md border border-emerald-500/20 w-fit">
          <ShieldCheck className="h-4 w-4 shrink-0" />
          <span className="text-xs font-bold leading-none tracking-tight">SAFE: Secured via SSH Tunneling</span>
        </div>
      );
    }

    if (host === "0.0.0.0") {
      return (
        <div className="flex items-center gap-2 mt-2 text-destructive bg-destructive/10 px-2 py-1.5 rounded-md border border-destructive/20 w-fit">
          <AlertTriangle className="h-4 w-4 shrink-0" />
          <span className="text-xs font-semibold">
            CRITICAL: {isAdmin ? "Admin panel" : "This API"} is exposed to everyone on the network.
          </span>
        </div>
      );
    } else if (host === "localhost" || host === "127.0.0.1") {
      return (
        <div className="flex items-center gap-2 mt-2 text-emerald-500 bg-emerald-500/10 px-2 py-1.5 rounded-md border border-emerald-500/20 w-fit">
          <ShieldCheck className="h-4 w-4 shrink-0" />
          <span className="text-xs font-semibold">SAFE: Restricted to local machine only.</span>
        </div>
      );
    }

    return (
      <div className="flex items-center gap-2 mt-2 text-amber-500 bg-amber-500/10 px-2 py-1.5 rounded-md border border-amber-500/20 w-fit">
        <AlertCircle className="h-4 w-4 shrink-0" />
        <span className="text-xs font-semibold">WARNING: Bound to a custom IP. Ensure firewall rules are applied.</span>
      </div>
    );
  };

  return (
    <div className="flex flex-col gap-8 max-w-4xl animate-in fade-in slide-in-from-bottom-2 duration-500">
      <div className="flex flex-col gap-1 border-b pb-6">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-primary/10 rounded-lg">
            <Server className="h-5 w-5 text-primary" />
          </div>
          <h1 className="text-2xl font-bold tracking-tight">Server Configuration</h1>
        </div>
        <p className="text-sm text-muted-foreground mt-2">
          Manage your system's network bindings for the Public API and Admin operations. Saving changes will restart the
          engine.
        </p>
      </div>

      <div className="grid gap-8">
        <div className="bg-card border rounded-lg overflow-hidden shadow-sm">
          <div className="p-4 bg-muted/50 border-b flex items-center gap-2">
            <div className="h-2 w-2 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.6)]" />
            <h3 className="font-semibold text-sm tracking-wide">Admin Panel Bindings</h3>
          </div>
          <div className="p-6 grid grid-cols-1 sm:grid-cols-2 gap-6">
            <div className="space-y-3">
              <div className="space-y-1">
                <Label htmlFor="adminPanelHost" className="text-sm font-semibold">
                  Admin Panel Host
                </Label>
                <p className="text-xs text-muted-foreground">Hostname or IP for admin operations.</p>
              </div>
              <Input
                id="adminPanelHost"
                value={config.adminSshEnabled ? "127.0.0.1" : config.adminPanelHost}
                required
                disabled={config.adminSshEnabled}
                onChange={(e) => setConfig({ ...config, adminPanelHost: e.target.value })}
                placeholder="localhost"
                className="font-mono text-sm shadow-sm"
              />
              <div
                className={`mt-3 flex items-center justify-between px-3 py-2.5 rounded-md border transition-all duration-300 ${!isSshVerified ? "opacity-30 grayscale pointer-events-none" : "bg-muted/20 border-border/50 shadow-sm"}`}
              >
                <div className="flex items-center gap-2.5">
                  <Shield
                    className={`h-4 w-4 ${config.adminSshEnabled ? "text-emerald-500" : "text-muted-foreground/60"}`}
                  />
                  <Label htmlFor="adminSsh" className="text-sm font-medium cursor-pointer text-foreground/80">
                    Use SSH Tunnel
                  </Label>
                </div>
                <div className="flex items-center gap-3">
                  <Switch
                    id="adminSsh"
                    checked={config.adminSshEnabled}
                    disabled={!isSshVerified}
                    onCheckedChange={(val) => setConfig({ ...config, adminSshEnabled: val })}
                  />
                </div>
              </div>
              {renderHostWarning(
                config.adminSshEnabled ? "127.0.0.1" : config.adminPanelHost,
                true,
                config.adminSshEnabled,
              )}
            </div>
            <div className="space-y-3">
              <div className="space-y-1">
                <Label htmlFor="adminPanelPort" className="text-sm font-semibold">
                  Admin Panel Port
                </Label>
                <p className="text-xs text-muted-foreground">Network port for the admin backend.</p>
              </div>
              <Input
                id="adminPanelPort"
                value={config.adminPanelPort}
                required
                onChange={(e) => setConfig({ ...config, adminPanelPort: e.target.value })}
                placeholder="8080"
                className="font-mono text-sm"
              />
            </div>
          </div>
        </div>

        <div className="bg-card border rounded-lg overflow-hidden shadow-sm">
          <div className="p-4 bg-muted/50 border-b flex items-center gap-2">
            <div className="h-2 w-2 rounded-full bg-blue-500 shadow-[0_0_8px_rgba(59,130,246,0.6)]" />
            <h3 className="font-semibold text-sm tracking-wide">Public API Bindings</h3>
          </div>
          <div className="p-6 grid grid-cols-1 sm:grid-cols-2 gap-6">
            <div className="space-y-3">
              <div className="space-y-1">
                <Label htmlFor="publicApiHost" className="text-sm font-semibold">
                  Public API Host
                </Label>
                <p className="text-xs text-muted-foreground">Hostname or IP for public API endpoints.</p>
              </div>
              <Input
                id="publicApiHost"
                value={config.publicSshEnabled ? "127.0.0.1" : config.publicApiHost}
                required
                disabled={config.publicSshEnabled}
                onChange={(e) => setConfig({ ...config, publicApiHost: e.target.value })}
                placeholder="localhost"
                className="font-mono text-sm shadow-sm"
              />
              <div
                className={`mt-3 flex items-center justify-between px-3 py-2.5 rounded-md border transition-all duration-300 ${!isSshVerified ? "opacity-30 grayscale pointer-events-none" : "bg-muted/20 border-border/50 shadow-sm"}`}
              >
                <div className="flex items-center gap-2.5">
                  <Shield
                    className={`h-4 w-4 ${config.publicSshEnabled ? "text-emerald-500" : "text-muted-foreground/60"}`}
                  />
                  <Label htmlFor="publicSsh" className="text-sm font-medium cursor-pointer text-foreground/80">
                    Use SSH Tunnel
                  </Label>
                </div>
                <div className="flex items-center gap-3">
                  <Switch
                    id="publicSsh"
                    checked={config.publicSshEnabled}
                    disabled={!isSshVerified}
                    onCheckedChange={(val) => setConfig({ ...config, publicSshEnabled: val })}
                  />
                </div>
              </div>
              {renderHostWarning(
                config.publicSshEnabled ? "127.0.0.1" : config.publicApiHost,
                false,
                config.publicSshEnabled,
              )}
            </div>
            <div className="space-y-3">
              <div className="space-y-1">
                <Label htmlFor="publicApiPort" className="text-sm font-semibold">
                  Public API Port
                </Label>
                <p className="text-xs text-muted-foreground">Network port serving public data.</p>
              </div>
              <Input
                id="publicApiPort"
                value={config.publicApiPort}
                required
                onChange={(e) => setConfig({ ...config, publicApiPort: e.target.value })}
                placeholder="8090"
                className="font-mono text-sm"
              />
            </div>
          </div>
        </div>

        <div className="bg-card border rounded-lg overflow-hidden shadow-sm">
          <div className="p-4 bg-muted/50 border-b flex items-center gap-2">
            <div
              className={`h-2 w-2 rounded-full ${config.adminSshEnabled || config.publicSshEnabled ? "bg-amber-500 shadow-[0_0_8px_rgba(245,158,11,0.6)]" : "bg-muted-foreground/30"}`}
            />
            <h3 className="font-semibold text-sm tracking-wide">SSH Configuration & Gateway</h3>
          </div>
          <div className="p-6 grid grid-cols-1 sm:grid-cols-2 gap-6 transition-all duration-300">
            <div className="space-y-3">
              <div className="space-y-1">
                <Label htmlFor="sshPort" className="text-sm font-semibold">
                  SSH Port
                </Label>
                <p className="text-xs text-muted-foreground">Secure port for tunnel access.</p>
              </div>
              <InputPassword
                id="sshPort"
                value={config.sshPort}
                onChange={(e) => setConfig({ ...config, sshPort: e.target.value })}
                placeholder="2222"
                className="font-mono text-sm"
              />
            </div>
            <div className="space-y-3">
              <div className="space-y-1">
                <Label htmlFor="sshUser" className="text-sm font-semibold">
                  SSH Username
                </Label>
                <p className="text-xs text-muted-foreground">System user for tunnel login.</p>
              </div>
              <InputPassword
                id="sshUser"
                value={config.sshUser}
                onChange={(e) => setConfig({ ...config, sshUser: e.target.value })}
                placeholder="admin"
                className="font-mono text-sm"
              />
            </div>
            <div className="sm:col-span-2 space-y-4">
              <div className="space-y-3">
                <div className="space-y-1">
                  <Label htmlFor="sshPass" className="text-sm font-semibold">
                    SSH Password
                  </Label>
                  <p className="text-xs text-muted-foreground">Encryption key for the tunnel gateway.</p>
                </div>
                <InputPassword
                  id="sshPass"
                  value={config.sshPassword}
                  onChange={(e) => setConfig({ ...config, sshPassword: e.target.value })}
                  placeholder="••••••••"
                  className="font-mono text-sm"
                />
              </div>

              <div className="flex flex-col gap-4 pt-2">
                <Button
                  variant={isSshVerified ? "outline" : "default"}
                  size="sm"
                  className="w-full sm:w-auto h-9"
                  disabled={isSaving}
                  onClick={async () => {
                    if (isSshVerified) {
                      setIsSshVerified(false);
                      return;
                    }

                    if (config.sshPort && config.sshUser && config.sshPassword) {
                      try {
                        const res = await fetchAPI("/settings/server/ssh/test", {
                          method: "POST",
                          body: JSON.stringify({ sshPort: config.sshPort }),
                        });

                        if (res.success) {
                          setIsSshVerified(true);
                          toast.success("Connection Verified", {
                            description: "Port is available and credentials are ready.",
                          });
                        } else {
                          toast.error("Verification Failed", {
                            description:
                              typeof res.error === "string"
                                ? res.error
                                : res.error?.message || "Port might be in use or restricted.",
                          });
                        }
                      } catch (err) {
                        toast.error("Network error during verification");
                      }
                    } else {
                      toast.error("Incomplete Credentials", {
                        description: "Please fill SSH port, user, and password first.",
                      });
                    }
                  }}
                >
                  {isSshVerified ? (
                    <>
                      <ShieldCheck className="mr-2 h-4 w-4 text-emerald-500" />
                      Configuration Unlocked (Click to Modify)
                    </>
                  ) : (
                    <>
                      <Shield className="mr-2 h-4 w-4" />
                      Verify & Unlock SSH
                    </>
                  )}
                </Button>

                {(config.adminSshEnabled || config.publicSshEnabled) && (
                  <div className="p-3 bg-orange-500/10 border border-orange-500/20 rounded-md">
                    <p className="text-sm text-orange-600 dark:text-orange-400 font-medium">
                      * Security Alert: Services will only be accessible via the defined SSH tunnel.
                    </p>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
        <div className="flex justify-end pt-4">
          <Button
            onClick={handleSave}
            disabled={isSaving}
            size="lg"
            className="w-full sm:w-auto min-w-[200px] shadow-md hover:shadow-lg transition-all"
          >
            {isSaving ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Save className="mr-2 h-4 w-4" />}
            Save & Restart Engine
          </Button>
        </div>
      </div>
    </div>
  );
}
