"use client";

import { Rocket } from "lucide-react";

export default function Home() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] gap-8 animate-in fade-in duration-700">
      <div className="text-center space-y-4 max-w-lg">
        <div className="inline-flex items-center justify-center size-20 rounded-full bg-indigo-500/10 mb-2">
          <Rocket className="size-10 text-indigo-500" />
        </div>
        <h1 className="text-4xl font-bold tracking-tight">Coming Soon</h1>
        <p className="text-muted-foreground leading-relaxed">
          Dashboard features are currently under development. Please use the sidebar to manage collections and system
          settings.
        </p>
      </div>
    </div>
  );
}
