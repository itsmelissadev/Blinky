import React, { useEffect, useState, createContext, useContext } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { fetchAPI } from "@/lib/api-client";

interface AuthContextType {
  isEnvExist: boolean | null;
  isInitialized: boolean | null;
  isAuthenticated: boolean | null;
  isLoading: boolean;
  user: any | null;
  checkStatus: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType>({
  isEnvExist: null,
  isInitialized: null,
  isAuthenticated: null,
  isLoading: true,
  user: null,
  checkStatus: async () => {},
});

export const useAuth = () => useContext(AuthContext);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [isEnvExist, setIsEnvExist] = useState<boolean | null>(null);
  const [isInitialized, setIsInitialized] = useState<boolean | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState<boolean | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [user, setUser] = useState<any | null>(null);

  const navigate = useNavigate();
  const location = useLocation();

  const checkStatus = async () => {
    try {
      const setupStatus = await fetchAPI("/setup/status");
      const envOk = setupStatus.data?.is_env_exist;
      setIsEnvExist(envOk);

      if (!envOk) {
        setIsInitialized(false);
        setIsAuthenticated(false);
        return;
      }

      const initData = await fetchAPI("/admins/initialized");
      const initialized = initData.data?.success || false;
      setIsInitialized(initialized);

      const meData = await fetchAPI("/admins/me").catch(() => ({ success: false, data: undefined }));
      if (meData.success && meData.data) {
        setIsAuthenticated(true);
        setUser(meData.data);
      } else {
        setIsAuthenticated(false);
        setUser(null);
      }
    } catch (error) {
      console.error("Auth check failed", error);
      setIsAuthenticated(false);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    checkStatus();
  }, []);

  useEffect(() => {
    if (isLoading) return;

    const path = location.pathname;

    if (isEnvExist === false || isInitialized === false) {
      if (path !== "/setup") {
        navigate("/setup", { replace: true });
      }
      return;
    }

    if (isInitialized === true) {
      if (isAuthenticated === true) {
        if (path === "/login" || path === "/setup") {
          navigate("/", { replace: true });
        }
      } else if (isAuthenticated === false) {
        if (path !== "/login") {
          navigate("/login", { replace: true });
        }
      }
    }
  }, [isEnvExist, isInitialized, isAuthenticated, isLoading, location.pathname]);

  if (isLoading) {
    return (
      <div className="h-screen w-full flex items-center justify-center bg-zinc-950">
        <div className="flex flex-col items-center gap-4">
          <div className="w-12 h-12 border-4 border-indigo-500/20 border-t-indigo-500 rounded-full animate-spin" />
          <p className="text-sm text-zinc-500 animate-pulse font-medium tracking-widest uppercase">
            Loading Dashboard...
          </p>
        </div>
      </div>
    );
  }

  return (
    <AuthContext.Provider value={{ isEnvExist, isInitialized, isAuthenticated, isLoading, user, checkStatus }}>
      {children}
    </AuthContext.Provider>
  );
}
