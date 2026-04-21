import { Loader2 } from "lucide-react";

interface PageLoaderProps {
  title?: string;
  description?: string;
  minHeight?: string;
}

export function PageLoader({
  title = "Loading...",
  description = "Please wait while we fetch the latest data.",
  minHeight = "400px",
}: PageLoaderProps) {
  return (
    <div
      className="flex flex-col items-center justify-center w-full gap-4 animate-in fade-in duration-500"
      style={{ minHeight }}
    >
      <Loader2 className="size-8 text-primary/60 animate-spin" />

      <div className="flex flex-col items-center gap-1 text-center">
        <h3 className="text-sm font-semibold text-foreground">{title}</h3>
        <p className="text-xs text-muted-foreground">{description}</p>
      </div>
    </div>
  );
}
