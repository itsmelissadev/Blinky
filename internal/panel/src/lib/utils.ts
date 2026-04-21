import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatSize(bytes: number) {
  if (bytes === 0) return "0 Bytes";
  const k = 1024;
  const sizes = ["Bytes", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
}

export function formatDate(dateString: string | Date) {
  const date = new Date(dateString);
  return date.toLocaleString();
}

export type OSType = "windows" | "macos" | "linux" | "android" | "ios" | "unknown";

export function getOS(): OSType {
  if (typeof window === "undefined") return "unknown";

  const userAgent = navigator.userAgent;
  const platform = (navigator as any).userAgentData?.platform || navigator.platform;

  if (/Win/i.test(platform) || /Win/i.test(userAgent)) return "windows";
  if (/Mac/i.test(platform) || /Mac/i.test(userAgent)) return "macos";
  if (/Linux/i.test(platform) || /Linux/i.test(userAgent)) return "linux";
  if (/Android/i.test(userAgent)) return "android";
  if (/iPhone|iPad|iPod/i.test(userAgent)) return "ios";

  return "unknown";
}

export function getPostgresPathPlaceholder(): string {
  const os = getOS();
  switch (os) {
    case "windows":
      return "e.g. C:\\Program Files\\PostgreSQL\\18";
    case "macos":
      return "e.g. /Library/PostgreSQL/18";
    case "linux":
      return "e.g. /usr/lib/postgresql/18";
    default:
      return "e.g. /usr/lib/postgresql/18";
  }
}

export function getPostgresDataPlaceholder(): string {
  const os = getOS();
  switch (os) {
    case "windows":
      return "e.g. C:\\Program Files\\PostgreSQL\\18\\data";
    case "macos":
      return "e.g. /Library/PostgreSQL/18/data";
    case "linux":
      return "e.g. /var/lib/postgresql/18/data";
    default:
      return "e.g. /var/lib/postgresql/18/data";
  }
}

export function formatLabel(name: string): string {
  return name
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

export function getCookie(name: string): string {
  if (typeof document === "undefined") return "";
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) return parts.pop()?.split(";").shift() || "";
  return "";
}

export function setCookie(name: string, value: string, maxAgeInSeconds: number = 60 * 60 * 24 * 7) {
  if (typeof document === "undefined") return;
  document.cookie = `${name}=${value}; path=/; max-age=${maxAgeInSeconds}`;
}
