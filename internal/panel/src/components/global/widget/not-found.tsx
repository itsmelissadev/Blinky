import { SearchX, ArrowLeft } from "lucide-react";
import { Link } from "react-router-dom";
import { Button } from "@/components/ui/button";

interface NotFoundProps {
  title?: string;
  message?: string;
  backTo?: string;
  backText?: string;
}

export function NotFound({
  title = "Item Not Found",
  message = "The item you are looking for does not exist or has been deleted.",
  backTo = "/collections",
  backText = "Back to Collections",
}: NotFoundProps) {
  return (
    <div className="flex flex-col items-center justify-center p-8 text-center animate-in fade-in duration-500">
      <div className="flex size-12 items-center justify-center rounded-lg bg-muted text-muted-foreground mb-4">
        <SearchX className="size-6" />
      </div>

      <div className="space-y-1 mb-6">
        <h1 className="text-lg font-semibold tracking-tight">{title}</h1>
        <p className="text-sm text-muted-foreground max-w-xs mx-auto">{message}</p>
      </div>

      <Button variant="outline" size="sm" asChild>
        <Link to={backTo}>
          <ArrowLeft className="size-4" />
          {backText}
        </Link>
      </Button>
    </div>
  );
}
