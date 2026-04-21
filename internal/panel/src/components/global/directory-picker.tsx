"use client";

import { useState, useCallback, useEffect } from "react";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
  SheetFooter,
  SheetClose,
} from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";
import { fetchAPI } from "@/lib/api-client";
import { toast } from "sonner";
import { Folder, Home, ChevronLeft, ChevronRight, Loader2 } from "lucide-react";
import { ScrollArea } from "@/components/ui/scroll-area";

interface DirectoryPickerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSelect: (path: string) => void;
  initialPath?: string;
}

export function DirectoryPickerSheet({ open, onOpenChange, onSelect, initialPath = "" }: DirectoryPickerProps) {
  const [currentPath, setCurrentPath] = useState(initialPath);
  const [dirs, setDirs] = useState<string[]>([]);
  const [parentPath, setParentPath] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  const browse = useCallback(async (path: string) => {
    setIsLoading(true);
    try {
      const res = await fetchAPI(`/system/files/browse?path=${encodeURIComponent(path)}`);
      if (res.success) {
        setCurrentPath(res.data.path);
        setDirs(res.data.dirs || []);
        setParentPath(res.data.parent || "");
      } else {
        toast.error("Failed to read directory");
      }
    } catch (error) {
      toast.error("An error occurred while browsing");
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (open) {
      browse(initialPath);
    }
  }, [open, initialPath, browse]);

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="max-sm:w-full! max-sm:inset-0! md:max-w-xl! flex flex-col p-0 gap-0 h-full bg-background border-l shadow-2xl">
        <div tabIndex={0} className="sr-only" aria-hidden="true" />
        <SheetHeader className="p-4 border-b gap-0 relative">
          <SheetTitle>Browse Server Directory</SheetTitle>
          <SheetDescription>Navigate through the server's file system to select a target path.</SheetDescription>
        </SheetHeader>

        <ScrollArea className="flex-1 min-h-0 w-full">
          <div className="p-4 space-y-4">
            {/* Current Path Bar */}
            <div className="space-y-2">
              <span className="text-sm font-semibold mb-1 block text-muted-foreground">Location</span>
              <div className="flex items-center gap-2 p-3 bg-muted/30 rounded-lg border text-xs font-mono break-all leading-relaxed shadow-sm">
                <Home className="w-3.5 h-3.5 text-primary shrink-0" />
                <span className="opacity-20">|</span>
                <span className="text-foreground">{currentPath || "Root"}</span>
              </div>
            </div>

            {/* List Header */}
            <div className="flex items-center justify-between pb-1 border-b mb-2">
              <span className="text-xs font-bold uppercase tracking-wider text-muted-foreground/60">
                Folders & Directories
              </span>
              {isLoading && <Loader2 className="h-3.5 w-3.5 animate-spin text-primary" />}
            </div>

            {/* Folder List Container */}
            <div className="space-y-1.5 min-h-[300px]">
              {(currentPath !== "" || (currentPath === "" && parentPath !== "")) && (
                <button
                  className="w-full flex items-center gap-2 p-2.5 hover:bg-muted/50 rounded-lg text-left text-sm group transition-all"
                  onClick={() => browse(parentPath)}
                >
                  <ChevronLeft className="w-4 h-4 text-muted-foreground group-hover:text-primary transition-colors" />
                  <span className="text-muted-foreground group-hover:text-foreground italic text-xs">
                    .. (Parent Directory)
                  </span>
                </button>
              )}

              {isLoading && dirs.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-32 text-muted-foreground opacity-50">
                  <Loader2 className="w-8 h-8 animate-spin mb-3 text-primary" />
                  <span className="text-sm font-medium">Scanning folders...</span>
                </div>
              ) : dirs.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-32 text-muted-foreground opacity-30">
                  <Folder className="w-12 h-12 mb-3 opacity-20" />
                  <span className="text-sm font-medium">This directory is empty</span>
                </div>
              ) : (
                dirs.map((dir) => (
                  <button
                    key={dir}
                    className="w-full flex items-center justify-between p-3.5 hover:bg-primary/5 rounded-lg group text-left border border-transparent hover:border-primary/20 transition-all shadow-sm bg-background"
                    onClick={() => {
                      let nextPath = "";
                      if (currentPath === "") {
                        nextPath = dir;
                      } else {
                        const separator = currentPath.endsWith("/") || currentPath.endsWith("\\") ? "" : "/";
                        nextPath = `${currentPath}${separator}${dir}`;
                      }
                      browse(nextPath);
                    }}
                  >
                    <div className="flex items-center gap-3.5 overflow-hidden">
                      <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg border bg-primary/5 text-primary group-hover:bg-primary group-hover:text-primary-foreground transition-all">
                        <Folder className="h-4 w-4" />
                      </div>
                      <span className="text-sm font-semibold truncate text-foreground group-hover:text-primary transition-colors">
                        {dir}
                      </span>
                    </div>
                    <ChevronRight className="w-4 h-4 text-muted-foreground opacity-0 group-hover:opacity-100 transition-all translate-x-1 group-hover:translate-x-0" />
                  </button>
                ))
              )}
            </div>
          </div>
        </ScrollArea>

        <SheetFooter className="p-4 border-t bg-muted/5">
          <div className="flex w-full justify-end gap-2">
            <SheetClose asChild>
              <Button variant="outline">Cancel</Button>
            </SheetClose>
            <Button
              onClick={() => {
                onSelect(currentPath);
                onOpenChange(false);
              }}
            >
              Select Directory
            </Button>
          </div>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  );
}
