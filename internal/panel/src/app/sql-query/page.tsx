"use client";

import { useState } from "react";
import Editor from "@monaco-editor/react";
import { useTheme } from "next-themes";
import { Play, Database, Table, AlertCircle, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { fetchAPI } from "@/lib/api-client";
import { toast } from "sonner";

export default function SQLQueryPage() {
  const { resolvedTheme } = useTheme();
  const [query, setQuery] = useState("SELECT * FROM _collections;");
  const [results, setResults] = useState<{ columns: string[]; rows: any[] } | null>(null);
  const [isExecuting, setIsExecuting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [editorHeight, setEditorHeight] = useState(200);

  const handleEditorMount = (editor: any) => {
    const updateHeight = () => {
      const contentHeight = Math.min(Math.max(editor.getContentHeight(), 200), 600);
      setEditorHeight(contentHeight);
    };
    editor.onDidContentSizeChange(updateHeight);
    updateHeight();
  };

  const handleExecute = async () => {
    if (!query.trim()) return;

    setIsExecuting(true);
    setError(null);
    setResults(null);
    try {
      const res = await fetchAPI("/system/sql", {
        method: "POST",
        body: JSON.stringify({ query }),
      });

      if (res.success) {
        setResults(res.data);
        toast.success("Query executed successfully");
      } else {
        setError(res.error?.message || "Failed to execute query");
      }
    } catch (err: any) {
      setError(err.message || "An unexpected error occurred");
    } finally {
      setIsExecuting(false);
    }
  };

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <h1 className="text-2xl font-bold tracking-tight">SQL Query Runner</h1>
        <p className="text-sm text-muted-foreground">Execute raw SQL commands directly against the engine.</p>
      </div>

      <Card className="p-0 gap-0 overflow-hidden border">
        <CardHeader className="flex flex-row items-center justify-between p-4! border-b">
          <div className="flex items-center gap-2">
            <Database className="h-4 w-4" />
            <CardTitle className="text-base font-bold">SQL Editor</CardTitle>
          </div>
          <Button size="sm" onClick={handleExecute} disabled={isExecuting} className="gap-2">
            {isExecuting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Play className="h-4 w-4" />}
            Run Query
          </Button>
        </CardHeader>
        <CardContent className="p-0">
          <div className="font-mono p-0 m-0 border-none transition-[height] duration-200">
            <Editor
              height={editorHeight}
              language="sql"
              theme={resolvedTheme === "dark" ? "vs-dark" : "light"}
              value={query}
              onChange={(value) => setQuery(value || "")}
              onMount={handleEditorMount}
              options={{
                minimap: { enabled: false },
                fontSize: 14,
                padding: { top: 16, bottom: 16 },
                scrollBeyondLastLine: false,
                wordWrap: "on",
                lineNumbers: "on",
                renderLineHighlight: "all",
                roundedSelection: false,
                hideCursorInOverviewRuler: true,
                scrollbar: {
                  vertical: "hidden",
                  handleMouseWheel: true,
                },
                automaticLayout: true,
              }}
            />
          </div>
        </CardContent>
      </Card>

      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle className="font-bold">SQL Execution Error</AlertTitle>
          <AlertDescription className="font-mono text-xs mt-2 whitespace-pre-wrap leading-relaxed">
            {error}
          </AlertDescription>
        </Alert>
      )}

      {results && (
        <Card className="p-0 gap-0">
          <CardHeader className="flex flex-row items-center justify-between p-4! border-b">
            <div className="flex items-center gap-2">
              <Table className="h-4 w-4" />
              <CardTitle className="text-base font-bold">Results</CardTitle>
              <span className="text-xs text-muted-foreground">({results.rows.length} rows)</span>
            </div>
          </CardHeader>
          <CardContent className="p-0 pt-0">
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b bg-muted/30">
                    {results.columns.map((col) => (
                      <th key={col} className="px-4 py-2 text-left font-medium text-muted-foreground tracking-wider">
                        {col}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {results.rows.length === 0 ? (
                    <tr>
                      <td colSpan={results.columns.length} className="px-4 py-8 text-center text-muted-foreground">
                        No rows returned.
                      </td>
                    </tr>
                  ) : (
                    results.rows.map((row, i) => (
                      <tr key={i} className="hover:bg-muted/50 transition-colors">
                        {results.columns.map((col) => (
                          <td key={col} className="px-4 py-2 font-mono whitespace-nowrap">
                            {row[col] === null ? (
                              <span className="opacity-30 italic">NULL</span>
                            ) : typeof row[col] === "object" ? (
                              JSON.stringify(row[col])
                            ) : (
                              String(row[col])
                            )}
                          </td>
                        ))}
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
