import { BrowserRouter, Routes, Route, Navigate, useLocation } from "react-router-dom";
import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/app-sidebar";
import { TooltipProvider } from "@/components/ui/tooltip";
import { ThemeProvider } from "@/components/theme-provider";
import { ThemeToggle } from "@/components/theme-toggle";
import { Toaster } from "@/components/ui/sonner";
import { SchemaProvider } from "@/hooks/use-schema";
import { Separator } from "@/components/ui/separator";
import { AuthProvider, useAuth } from "@/components/auth-guard";

// Pages
import DashboardPage from "./app/page";
import AdminsPage from "./app/admins/page";
import CollectionsPage from "./app/collections/page";
import CollectionPreviewPage from "./app/collections/[name]/page";
import NotFoundPage from "./app/not-found-page";
import LoginPage from "./app/login/page";
import BackupPage from "./app/settings/backup/page";
import EnvironmentsPage from "./app/settings/environments/page";
import PostgresSettingsPage from "./app/settings/postgresql/page";
import ServerSettingsPage from "./app/settings/server/page";
import SetupPage from "./app/initialize/page";
import SQLQueryPage from "./app/sql-query/page";

function AuthenticatedLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-background">
      <SidebarProvider>
        <AppSidebar />
        <div className="flex flex-1 flex-col min-w-0 overflow-hidden">
          <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4 justify-between bg-background/50 backdrop-blur-md sticky top-0 z-10">
            <div className="flex items-center gap-2">
              <SidebarTrigger className="-ml-1" />
              <Separator orientation="vertical" className="mr-2 h-4" />
            </div>
            <div className="flex items-center gap-4">
              <ThemeToggle />
            </div>
          </header>
          <main className="flex-1 overflow-y-auto p-4 md:p-8">
            <div className="max-w-7xl mx-auto">{children}</div>
          </main>
        </div>
      </SidebarProvider>
    </div>
  );
}

function PageRouter() {
  const { isInitialized, isAuthenticated, isLoading } = useAuth();
  const location = useLocation();

  if (isLoading) {
    return (
      <div className="h-screen w-full flex items-center justify-center bg-zinc-950">
        <div className="w-8 h-8 border-2 border-indigo-500/20 border-t-indigo-500 rounded-full animate-spin" />
      </div>
    );
  }

  const isPublicPage = location.pathname === "/login" || location.pathname === "/setup";
  if (isAuthenticated && !isPublicPage) {
    return (
      <AuthenticatedLayout>
        <Routes>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/admins" element={<AdminsPage />} />
          <Route path="/collections" element={<CollectionsPage />} />
          <Route path="/collections/:name" element={<CollectionPreviewPage />} />
          <Route path="/settings/backup" element={<BackupPage />} />
          <Route path="/settings/environments" element={<EnvironmentsPage />} />
          <Route path="/settings/postgresql" element={<PostgresSettingsPage />} />
          <Route path="/settings/server" element={<ServerSettingsPage />} />
          <Route path="/sql-query" element={<SQLQueryPage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </AuthenticatedLayout>
    );
  }

  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/setup" element={<SetupPage />} />
      <Route path="*" element={<Navigate to={isInitialized === false ? "/setup" : "/login"} replace />} />
    </Routes>
  );
}

export default function App() {
  return (
    <ThemeProvider attribute="class" defaultTheme="system" enableSystem disableTransitionOnChange>
      <TooltipProvider>
        <BrowserRouter>
          <AuthProvider>
            <SchemaProvider>
              <PageRouter />
              <Toaster position="top-right" expand={true} richColors />
            </SchemaProvider>
          </AuthProvider>
        </BrowserRouter>
      </TooltipProvider>
    </ThemeProvider>
  );
}
