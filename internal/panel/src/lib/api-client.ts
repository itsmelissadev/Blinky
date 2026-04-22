import { getCookie } from "./utils";
import { toast } from "sonner";

const getApiUrl = () => {
  if (typeof window !== "undefined") {
    const port = (import.meta.env.ADMIN_PANEL_PORT as string) || "8080";
    return "";
  }

  const port = (import.meta.env.ADMIN_PANEL_PORT as string) || "8080";
  return `http://localhost:${port}`;
};

const API_URL = getApiUrl();

export interface APIResponse<T = any> {
  success: boolean;
  data?: T;
  meta?: {
    total: number;
    limit: number;
    offset: number;
  };
  error?: {
    number: number;
    code: string;
    message: string;
  };
}

export async function fetchAPI<T = any>(endpoint: string, options: RequestInit = {}): Promise<APIResponse<T>> {
  const normalizedEndpoint = endpoint.startsWith("/") ? endpoint : `/${endpoint}`;
  const url = `${API_URL}/_api${normalizedEndpoint}`;

  try {
    const response = await fetch(url, {
      credentials: "include",
      ...options,
      headers: {
        "Content-Type": "application/json",
        "X-CSRF-Token": getCookie("csrf_"),
        ...options.headers,
      },
    });

    const data: APIResponse<T> = await response.json();

    if (
      response.status === 401 &&
      !endpoint.includes("/login") &&
      !endpoint.includes("/initialized") &&
      !endpoint.includes("/me")
    ) {
      throw new Error("Unauthorized");
    }

    if (!data.success) {
      if (response.status >= 500) {
        const message = data.error?.message || "An unexpected error occurred";
        toast.error("System Error", { description: message });
        throw new Error(message);
      }
      return data;
    }

    return data;
  } catch (error: any) {
    if (error.message !== "Unauthorized") {
      logger.error("[API-CLIENT]", error.message);
    }
    throw error;
  }
}

export function getAPIUrl(endpoint: string) {
  const normalizedEndpoint = endpoint.startsWith("/") ? endpoint : `/${endpoint}`;
  return `${API_URL}/_api${normalizedEndpoint}`;
}

const logger = {
  error: (tag: string, message: string) => {
    console.error(`${tag} ${message}`);
  },
};
