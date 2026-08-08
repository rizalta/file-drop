import { useState, useEffect } from "react";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { getFileIcon } from "../lib/file-icon";

interface DropMetadata {
  filename: string;
  file_size: number;
  mime_type: string;
  is_text: boolean;
  text_content?: string;
  download_count: number;
  created_at: string;
}

interface DropCardProps {
  id: string;
  mode?: "share" | "download";
  onReset?: () => void;
}

const DropCard = ({ id, mode = "download", onReset }: DropCardProps) => {
  const [data, setData] = useState<DropMetadata | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [copiedLink, setCopiedLink] = useState(false);

  const [revealedText, setRevealedText] = useState<string | null>(null);
  const [revealing, setRevealing] = useState(false);
  const [copiedText, setCopiedText] = useState(false);

  const dropUrl = `${window.location.origin}/f/${id}`;

  useEffect(() => {
    const fetchMetadata = async () => {
      setLoading(true);
      setError("");

      try {
        const res = await fetch(`/api/f/${id}`);
        if (!res.ok) {
          if (res.status === 404) {
            throw new Error("Filedrop not found or expired.");
          }
          throw new Error("Failed to load drop details.");
        }

        const metadata: DropMetadata = await res.json();
        setData(metadata);
      } catch (err: any) {
        setError(err.message || "Something went wrong.");
      } finally {
        setLoading(false);
      }
    };

    if (id) {
      fetchMetadata();
    }
  }, [id]);

  const handleCopyLink = () => {
    navigator.clipboard.writeText(dropUrl);
    setCopiedLink(true);
    setTimeout(() => setCopiedLink(false), 2000);
  };

  const handleDownloadFile = () => {
    window.location.href = `/api/f/${id}?download=true`;
  };

  const handleRevealText = async () => {
    if (revealedText) {
      navigator.clipboard.writeText(revealedText);
      setCopiedText(true);
      setTimeout(() => setCopiedText(false), 2000);
      return;
    }

    setRevealing(true);
    try {
      const res = await fetch(`/api/f/${id}?download=true`);
      if (!res.ok) {
        throw new Error("Failed to reveal text snippet.");
      }
      const text = await res.text();
      setRevealedText(text);
      navigator.clipboard.writeText(text);
      setCopiedText(true);
      setTimeout(() => setCopiedText(false), 2000);
    } catch (err: any) {
      setError(err.message || "Failed to reveal text snippet.");
    } finally {
      setRevealing(false);
    }
  };

  if (loading) {
    return (
      <div className="w-full max-w-xl p-8 bg-card border border-border rounded-none shadow-sm flex flex-col items-center justify-center space-y-3 h-48">
        <svg className="animate-spin h-5 w-5 text-primary" viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
        </svg>
        <span className="text-xs font-mono text-muted-foreground uppercase tracking-wider">
          Fetching Filedrop...
        </span>
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="w-full max-w-xl p-6 bg-card border border-border rounded-none shadow-sm space-y-4 text-center">
        <div className="p-4 bg-destructive/10 border border-destructive/30 text-destructive text-xs font-mono">
          ⚠️ {error || "Filedrop unavailable."}
        </div>
        {onReset && (
          <Button
            type="button"
            variant="outline"
            onClick={onReset}
            className="w-full rounded-none uppercase text-xs font-bold tracking-wider h-10 border-border"
          >
            Create New Filedrop
          </Button>
        )}
      </div>
    );
  }

  return (
    <div className="w-full max-w-xl p-6 bg-card border border-border rounded-none shadow-sm space-y-6">
      <div className="flex items-center justify-between border-b border-border/60 pb-3">
        <div className="flex items-center gap-2">
          <span className="h-2 w-2 bg-primary rounded-full animate-pulse" />
          <span className="text-xs font-bold uppercase tracking-wider text-muted-foreground">
            {mode === "share" ? "Filedrop Created" : "Ready to Access"}
          </span>
        </div>

        <div className="flex items-center gap-3 text-[11px] font-mono text-muted-foreground/80">
          <span>Downloads: {data.download_count}</span>
        </div>
      </div>

      {!data.is_text ? (
        <div className="flex items-center h-36 w-full border border-border bg-muted/10 overflow-hidden">
          <div className="w-1/4 h-full bg-muted/20 border-r border-border flex items-center justify-center text-foreground shrink-0">
            {getFileIcon(data.filename)}
          </div>
          <div className="w-3/4 h-full flex flex-col justify-center px-5">
            <p className="text-sm font-bold text-foreground tracking-tight truncate" title={data.filename}>
              {data.filename}
            </p>
            <p className="text-xs text-muted-foreground font-mono mt-1">
              {(data.file_size / 1024).toFixed(1)} KB
            </p>
          </div>
        </div>
      ) : (
        <div className="w-full h-36 p-4 text-xs font-mono bg-background border border-border/80 rounded-none text-foreground overflow-y-auto whitespace-pre-wrap flex flex-col justify-center items-center text-center">
          {revealedText !== null ? (
            <p className="text-left w-full h-full font-mono text-xs">{revealedText}</p>
          ) : (
            <div className="space-y-1.5 text-muted-foreground">
              <p className="font-semibold text-foreground text-xs uppercase tracking-wider">🔒 Protected Text Snippet</p>
              <p className="text-[11px]">Click reveal below to access and copy text content.</p>
            </div>
          )}
        </div>
      )}

      {mode === "share" ? (
        <div className="space-y-4">
          <div className="space-y-1.5">
            <label className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
              Shareable Link
            </label>
            <div className="flex items-center gap-1.5">
              <Input
                readOnly
                value={dropUrl}
                className="rounded-none text-xs font-mono h-10 bg-background border-border"
              />
              <Button
                type="button"
                onClick={handleCopyLink}
                className="rounded-none uppercase text-xs font-bold tracking-wider h-10 px-5 shrink-0"
              >
                {copiedLink ? "Copied!" : "Copy"}
              </Button>
            </div>
          </div>

          {onReset && (
            <Button
              type="button"
              variant="outline"
              onClick={onReset}
              className="w-full rounded-none uppercase text-xs font-bold tracking-wider h-10 border-border hover:bg-muted/30"
            >
              Create Another Filedrop
            </Button>
          )}
        </div>
      ) : (
        <div className="space-y-3">
          {data.is_text ? (
            <Button
              type="button"
              disabled={revealing}
              onClick={handleRevealText}
              className="w-full rounded-none uppercase text-xs font-bold tracking-wider h-10 disabled:opacity-50"
            >
              {revealing
                ? "Revealing Text..."
                : copiedText
                ? "Copied to Clipboard!"
                : revealedText !== null
                ? "Copy Text to Clipboard"
                : "Reveal & Copy Text"}
            </Button>
          ) : (
            <Button
              type="button"
              onClick={handleDownloadFile}
              className="w-full rounded-none uppercase text-xs font-bold tracking-wider h-10"
            >
              Download File
            </Button>
          )}
        </div>
      )}
    </div>
  );
};

export default DropCard;
