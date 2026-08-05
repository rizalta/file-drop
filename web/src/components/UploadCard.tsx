import React, { useState, useEffect } from "react";
import { Tabs, TabsList, TabsContent, TabsTrigger } from "./ui/tabs";
import { Button } from "./ui/button";
import { Checkbox } from "./ui/checkbox";
import { Input } from "./ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "./ui/select";
import { getFileIcon } from "../lib/file-icon";

const UploadCard = () => {
  const [activeTab, setActiveTab] = useState<"file" | "text">("file");
  const [file, setFile] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [expiresIn, setExpiresIn] = useState("7d");
  const [burnAfterDownload, setBurnAfterDownload] = useState(false);
  const [isCustomExpiry, setIsCustomExpiry] = useState(false);
  const [customExpiresIn, setCustomExpiresIn] = useState("");
  const [content, setContent] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [isDragging, setIsDragging] = useState(false);

  useEffect(() => {
    if (file && file.type.startsWith("image/")) {
      const url = URL.createObjectURL(file);
      setPreviewUrl(url);
      return () => URL.revokeObjectURL(url);
    }
    setPreviewUrl(null);
  }, [file]);

  const handleDragEnter = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(true);
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    e.dataTransfer.dropEffect = "copy";
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.currentTarget.contains(e.relatedTarget as Node)) return;
    setIsDragging(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);

    const droppedFile = e.dataTransfer.files?.[0];
    if (droppedFile) {
      setTimeout(() => {
        setFile(droppedFile);
      }, 0);
    }
  };

  const handleUpload = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (activeTab === "file" && !file) {
      setError("Please select a file to drop.");
      return;
    }
    if (activeTab === "text" && !content.trim()) {
      setError("Please enter text content to drop.");
      return;
    }
    if (isCustomExpiry && !customExpiresIn.trim()) {
      setError("Please enter a custom expiration duration (e.g. 30m, 3d).");
      return;
    }

    setLoading(true);

    const formData = new FormData();
    const finalExpiresIn = isCustomExpiry ? customExpiresIn.trim() : expiresIn;

    formData.append("burn_after_download", burnAfterDownload ? "true" : "false");
    formData.append("expires_in", finalExpiresIn);

    if (activeTab === "file" && file) {
      formData.append("file", file);
    } else if (activeTab === "text") {
      formData.append("content", content);
    }

    try {
      const res = await fetch("/api/upload", {
        method: "POST",
        body: formData,
      });

      const data = await res.json();

      if (!res.ok) {
        setError(data.error || "Upload failed");
        return;
      }

      console.log("Uploaded successfully:", data);
    } catch (err) {
      setError("Something went wrong. Please check connection.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleUpload} className="w-full max-w-xl p-6 bg-card border border-border rounded-none shadow-sm space-y-6">
      <Tabs defaultValue="file" value={activeTab} onValueChange={(val) => setActiveTab(val as "file" | "text")} className="w-full">
        <TabsList className="w-full grid grid-cols-2 rounded-none">
          <TabsTrigger value="file" className="rounded-none text-xs font-semibold uppercase tracking-wider">File</TabsTrigger>
          <TabsTrigger value="text" className="rounded-none text-xs font-semibold uppercase tracking-wider">Text</TabsTrigger>
        </TabsList>

        <TabsContent value="file" className="mt-4">
          {!file ? (
            <label
              htmlFor="file-upload"
              className={`flex flex-col items-center justify-center h-36 border border-dashed transition-all cursor-pointer group ${isDragging
                ? "border-primary bg-primary/10"
                : "border-border/80 bg-muted/5 hover:bg-muted/20 hover:border-foreground/40"}`}
              onDragEnter={handleDragEnter}
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
              onDrop={handleDrop}
            >
              <input
                type="file"
                id="file-upload"
                name="file"
                className="hidden"
                onChange={(e) => { setFile(e.target.files?.[0] ?? null) }}
              />

              <div className="pointer-events-none flex flex-col items-center justify-center">
                <div className="mb-2 text-muted-foreground group-hover:text-foreground transition-colors">
                  <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M12 3v12" />
                    <path d="m17 8-5-5-5 5" />
                    <path d="M2 15v4a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-4" />
                  </svg>
                </div>

                <p className="text-xs font-medium text-foreground tracking-tight">
                  Drop file or <span className="underline underline-offset-4 decoration-border group-hover:decoration-foreground">browse</span>
                </p>
                <p className="text-[11px] text-muted-foreground/70 mt-1">
                  Supports any binary file
                </p>
              </div>
            </label>
          ) : (
            <div className="flex items-center h-36 w-full border border-border bg-muted/10 overflow-hidden">
              <div className="w-1/4 h-full bg-muted/20 border-r border-border flex items-center justify-center text-foreground shrink-0 overflow-hidden">
                {previewUrl ? (
                  <img
                    src={previewUrl}
                    alt="File preview"
                    className="w-full h-full object-cover"
                  />
                ) : (
                  getFileIcon(file.name)
                )}
              </div>

              <div className="w-3/4 h-full flex items-center justify-between px-5 py-3">
                <div className="flex flex-col truncate pr-2">
                  <p className="text-sm font-bold text-foreground tracking-tight truncate" title={file.name}>
                    {file.name}
                  </p>
                  <p className="text-xs text-muted-foreground font-mono mt-1">
                    {(file.size / 1024).toFixed(1)} KB
                  </p>
                </div>

                <button
                  type="button"
                  onClick={() => setFile(null)}
                  className="text-muted-foreground hover:text-foreground p-2 transition-colors cursor-pointer shrink-0 hover:bg-muted/30 border border-transparent hover:border-border"
                  title="Remove file"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <line x1="18" y1="6" x2="6" y2="18"></line>
                    <line x1="6" y1="6" x2="18" y2="18"></line>
                  </svg>
                </button>
              </div>
            </div>
          )}
        </TabsContent>

        <TabsContent value="text" className="mt-4">
          <textarea
            name="content"
            placeholder="Paste text notes, copied links, or code snippets here..."
            className="w-full h-32 p-4 text-xs font-mono bg-background border border-border/80 rounded-none text-foreground placeholder:text-muted-foreground/50 focus:outline-none focus:ring-1 focus:ring-ring resize-none"
            value={content}
            onChange={(e) => setContent(e.target.value)}
          />
        </TabsContent>
      </Tabs>

      <div className="flex items-center justify-between gap-4 pt-2 border-t border-border/60">
        <div className="flex items-center gap-2.5 w-1/2">
          <label htmlFor="expires_in" className="text-xs font-semibold uppercase tracking-wider text-muted-foreground whitespace-nowrap">
            Expires In
          </label>

          {!isCustomExpiry ? (
            <Select defaultValue="7d" onValueChange={(val) => {
              if (val === "custom") {
                setIsCustomExpiry(true);
              } else if (val) {
                setExpiresIn(val);
              } else {
                setExpiresIn("7d");
              }
            }}>
              <SelectTrigger id="expires_in" size="sm" className="rounded-none text-xs h-9 w-full px-3 bg-background border border-border">
                <SelectValue placeholder="7 Days" />
              </SelectTrigger>
              <SelectContent className="rounded-none">
                <SelectItem value="10m" label="10 Minutes" className="rounded-none text-xs">10 Minutes</SelectItem>
                <SelectItem value="1h" label="1 Hour" className="rounded-none text-xs">1 Hour</SelectItem>
                <SelectItem value="1d" label="1 Day" className="rounded-none text-xs">1 Day</SelectItem>
                <SelectItem value="7d" label="7 Days (Default)" className="rounded-none text-xs">7 Days (Default)</SelectItem>
                <SelectItem value="custom" label="Custom..." className="rounded-none text-xs">Custom...</SelectItem>
              </SelectContent>
            </Select>
          ) : (
            <div className="flex items-center gap-1.5 w-full h-9">
              <Input
                type="text"
                value={customExpiresIn}
                onChange={(e) => setCustomExpiresIn(e.target.value)}
                placeholder="e.g. 30m, 3d"
                className="rounded-none text-xs h-9 w-full px-3 bg-background border border-border"
                autoFocus
              />
              <Button
                type="button"
                variant="outline"
                size="icon"
                onClick={() => {
                  setIsCustomExpiry(false);
                  setCustomExpiresIn("");
                }}
                className="h-9 w-9 rounded-none shrink-0"
                title="Reset to select"
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <line x1="18" y1="6" x2="6" y2="18"></line>
                  <line x1="6" y1="6" x2="18" y2="18"></line>
                </svg>
              </Button>
            </div>
          )}
        </div>

        <div className="flex items-center gap-2.5 w-1/2 justify-center">
          <Checkbox id="burn_after_download" className="rounded-none h-4.5 w-4.5" checked={burnAfterDownload}
            onCheckedChange={setBurnAfterDownload} />
          <label
            htmlFor="burn_after_download"
            className="text-xs font-semibold uppercase tracking-wider text-muted-foreground cursor-pointer select-none whitespace-nowrap"
          >
            Burn After Download
          </label>
        </div>
      </div>

      {error && (
        <div className="p-3 bg-destructive/10 border border-destructive/30 text-destructive text-xs font-mono flex items-center justify-between">
          <span>⚠️ {error}</span>
          <button
            type="button"
            onClick={() => setError("")}
            className="ml-2 text-destructive hover:opacity-75 cursor-pointer font-bold"
          >
            ✕
          </button>
        </div>
      )}

      <Button
        type="submit"
        disabled={loading}
        className="w-full rounded-none uppercase text-xs font-bold tracking-wider h-10 disabled:opacity-50"
      >
        {loading ? (
          <span className="flex items-center gap-2">
            <svg className="animate-spin h-3.5 w-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
            </svg>
            Creating Filedrop...
          </span>
        ) : (
          "Create Filedrop"
        )}
      </Button>
    </form>
  );
};

export default UploadCard;
