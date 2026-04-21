import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { FolderOpen } from "lucide-react";
import { DirectoryPickerSheet } from "@/components/global/directory-picker";
import { getPostgresPathPlaceholder, getPostgresDataPlaceholder } from "@/lib/utils";
import { toast } from "sonner";
import { fetchAPI } from "@/lib/api-client";
import { useAuth } from "@/components/auth-guard";

export default function SetupPage() {
  const { isEnvExist, checkStatus } = useAuth();
  const [loading, setLoading] = useState(false);
  const [step, setStep] = useState(isEnvExist === false ? 1 : 2);
  const navigate = useNavigate();

  const [pickerOpen, setPickerOpen] = useState(false);
  const [pickerTarget, setPickerTarget] = useState<"folder" | "data" | null>(null);

  const [dbData, setDbData] = useState({
    POSTGRESQL_FOLDER_PATH: "",
    POSTGRESQL_DATA_PATH: "",
    POSTGRESQL_DB_HOST: "localhost",
    POSTGRESQL_DB_PORT: "5432",
    POSTGRESQL_DB_USER: "postgres",
    POSTGRESQL_DB_PASSWORD: "",
    POSTGRESQL_DB_NAME: "blinky_db",
  });

  const [adminData, setAdminData] = useState({
    nickname: "",
    username: "",
    email: "",
    password: "",
  });

  const handleTestConnection = async () => {
    setLoading(true);
    try {
      const res = await fetchAPI("/setup/test-db", {
        method: "POST",
        body: JSON.stringify({
          host: dbData.POSTGRESQL_DB_HOST,
          port: dbData.POSTGRESQL_DB_PORT,
          user: dbData.POSTGRESQL_DB_USER,
          password: dbData.POSTGRESQL_DB_PASSWORD,
          name: dbData.POSTGRESQL_DB_NAME,
        }),
      });

      if (res.success) {
        toast.success("Database connection verified!");
        return true;
      }
      return false;
    } catch (e: any) {
      toast.error(e.message || "Connection failed");
      return false;
    } finally {
      setLoading(false);
    }
  };

  const handleSaveEnv = async () => {
    const ok = await handleTestConnection();
    if (!ok) return;

    setLoading(true);
    try {
      const res = await fetchAPI("/setup/env", {
        method: "POST",
        body: JSON.stringify(dbData),
      });

      if (res.success) {
        toast.info("Config saved. Engine restarting...");
        setTimeout(() => {
          window.location.reload();
        }, 2000);
      }
    } catch (e: any) {
      toast.error(e.message);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateAdmin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      const res = await fetchAPI("/admins/user", {
        method: "POST",
        body: JSON.stringify(adminData),
      });

      if (res.success) {
        toast.success("System initialized successfully!");
        await checkStatus();
        navigate("/login");
      }
    } catch (e: any) {
      toast.error(e.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="container flex h-screen w-screen flex-col items-center justify-center">
      <Card className="w-full max-w-[400px]">
        <CardHeader className="space-y-1">
          <CardTitle className="text-2xl text-center">{step === 1 ? "Database Setup" : "Create Admin"}</CardTitle>
          <CardDescription className="text-center">
            {step === 1 ? "Connect Blinky to your PostgreSQL instance" : "Create the primary administrator account"}
          </CardDescription>
        </CardHeader>

        {step === 1 ? (
          <form
            className="flex flex-col gap-8"
            onSubmit={(e) => {
              e.preventDefault();
              handleSaveEnv();
            }}
          >
            <CardContent className="grid gap-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="grid gap-2">
                  <Label htmlFor="db-host">Host</Label>
                  <Input
                    id="db-host"
                    value={dbData.POSTGRESQL_DB_HOST}
                    onChange={(e) => setDbData({ ...dbData, POSTGRESQL_DB_HOST: e.target.value })}
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="db-port">Port</Label>
                  <Input
                    id="db-port"
                    value={dbData.POSTGRESQL_DB_PORT}
                    onChange={(e) => setDbData({ ...dbData, POSTGRESQL_DB_PORT: e.target.value })}
                  />
                </div>
              </div>
              <div className="grid gap-2">
                <Label htmlFor="db-name">Database Name</Label>
                <Input
                  id="db-name"
                  value={dbData.POSTGRESQL_DB_NAME}
                  onChange={(e) => setDbData({ ...dbData, POSTGRESQL_DB_NAME: e.target.value })}
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div className="grid gap-2">
                  <Label htmlFor="db-user">User</Label>
                  <Input
                    id="db-user"
                    value={dbData.POSTGRESQL_DB_USER}
                    onChange={(e) => setDbData({ ...dbData, POSTGRESQL_DB_USER: e.target.value })}
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="db-password">Password</Label>
                  <Input
                    id="db-password"
                    type="password"
                    value={dbData.POSTGRESQL_DB_PASSWORD}
                    onChange={(e) => setDbData({ ...dbData, POSTGRESQL_DB_PASSWORD: e.target.value })}
                  />
                </div>
              </div>
              <div className="grid gap-2">
                <Label htmlFor="pg-folder">PostgreSQL Folder Path</Label>
                <div className="flex gap-2">
                  <Input
                    id="pg-folder"
                    required
                    placeholder={getPostgresPathPlaceholder()}
                    value={dbData.POSTGRESQL_FOLDER_PATH}
                    onChange={(e) => setDbData({ ...dbData, POSTGRESQL_FOLDER_PATH: e.target.value })}
                  />
                  <Button
                    type="button"
                    variant="outline"
                    size="icon"
                    onClick={() => {
                      setPickerTarget("folder");
                      setPickerOpen(true);
                    }}
                  >
                    <FolderOpen className="h-4 w-4" />
                  </Button>
                </div>
              </div>
              <div className="grid gap-2">
                <Label htmlFor="pg-data">Data Path</Label>
                <div className="flex gap-2">
                  <Input
                    id="pg-data"
                    required
                    placeholder={getPostgresDataPlaceholder()}
                    value={dbData.POSTGRESQL_DATA_PATH}
                    onChange={(e) => setDbData({ ...dbData, POSTGRESQL_DATA_PATH: e.target.value })}
                  />
                  <Button
                    type="button"
                    variant="outline"
                    size="icon"
                    onClick={() => {
                      setPickerTarget("data");
                      setPickerOpen(true);
                    }}
                  >
                    <FolderOpen className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </CardContent>
            <CardFooter className="flex flex-col gap-2">
              <Button type="submit" className="w-full" disabled={loading}>
                {loading ? "Verifying..." : "Save & Continue"}
              </Button>
              <Button
                type="button"
                variant="ghost"
                className="w-full"
                onClick={handleTestConnection}
                disabled={loading}
              >
                Test Connection
              </Button>
            </CardFooter>
          </form>
        ) : (
          <form className="flex flex-col gap-8" onSubmit={handleCreateAdmin}>
            <CardContent className="grid gap-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="grid gap-2">
                  <Label htmlFor="admin-nickname">Nickname</Label>
                  <Input
                    id="admin-nickname"
                    required
                    value={adminData.nickname}
                    onChange={(e) => setAdminData({ ...adminData, nickname: e.target.value })}
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="admin-username">Username</Label>
                  <Input
                    id="admin-username"
                    required
                    value={adminData.username}
                    onChange={(e) => setAdminData({ ...adminData, username: e.target.value })}
                  />
                </div>
              </div>
              <div className="grid gap-2">
                <Label htmlFor="admin-email">Email</Label>
                <Input
                  id="admin-email"
                  type="email"
                  required
                  value={adminData.email}
                  onChange={(e) => setAdminData({ ...adminData, email: e.target.value })}
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="admin-password">Password</Label>
                <Input
                  id="admin-password"
                  type="password"
                  required
                  value={adminData.password}
                  onChange={(e) => setAdminData({ ...adminData, password: e.target.value })}
                />
              </div>
            </CardContent>
            <CardFooter>
              <Button className="w-full" type="submit" disabled={loading}>
                {loading ? "Processing..." : "Complete Setup"}
              </Button>
            </CardFooter>
          </form>
        )}
      </Card>

      {step === 1 && (
        <DirectoryPickerSheet
          open={pickerOpen}
          onOpenChange={setPickerOpen}
          initialPath={
            pickerTarget === "folder"
              ? dbData.POSTGRESQL_FOLDER_PATH
              : pickerTarget === "data"
                ? dbData.POSTGRESQL_DATA_PATH
                : ""
          }
          onSelect={(path) => {
            if (pickerTarget === "folder") {
              setDbData({ ...dbData, POSTGRESQL_FOLDER_PATH: path });
            } else if (pickerTarget === "data") {
              setDbData({ ...dbData, POSTGRESQL_DATA_PATH: path });
            }
          }}
        />
      )}
    </div>
  );
}
