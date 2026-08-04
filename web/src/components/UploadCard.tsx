import { useState } from "react";
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

const UploadCard = () => {
  const [isCustom, setIsCustom] = useState(false);
  const [customExpiry, setCustomExpiry] = useState("");

  return (
    <div className="w-full max-w-xl p-6 bg-card border border-border rounded-none shadow-sm space-y-6">
      <Tabs defaultValue="file" className="w-full">
        <TabsList className="w-full grid grid-cols-2 rounded-none p-1 bg-muted/30 border border-border">
          <TabsTrigger 
            value="file" 
            className="rounded-none text-xs font-bold uppercase tracking-wider data-active:bg-primary data-active:text-primary-foreground data-active:shadow-sm transition-all"
          >
            File
          </TabsTrigger>
          <TabsTrigger 
            value="text" 
            className="rounded-none text-xs font-bold uppercase tracking-wider data-active:bg-primary data-active:text-primary-foreground data-active:shadow-sm transition-all"
          >
            Text
          </TabsTrigger>
        </TabsList>

        <TabsContent value="file" className="mt-4">
          <label
            htmlFor="file-upload"
            className="flex flex-col items-center justify-center p-8 border border-dashed border-border/80 bg-muted/5 hover:bg-muted/20 hover:border-foreground/40 transition-all cursor-pointer group"
          >
            <input
              type="file"
              id="file-upload"
              name="file"
              className="hidden"
            />

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
          </label>
        </TabsContent>

        <TabsContent value="text" className="mt-4">
          <textarea
            name="content"
            placeholder="Paste text notes, copied links, or code snippets here..."
            className="w-full h-32 p-4 text-xs font-mono bg-background border border-border/80 rounded-none text-foreground placeholder:text-muted-foreground/50 focus:outline-none focus:ring-1 focus:ring-ring resize-none"
          />
        </TabsContent>
      </Tabs>

      <div className="flex items-center justify-between gap-4 pt-2 border-t border-border/60">
        <div className="flex items-center gap-2.5 w-1/2">
          <label htmlFor="expires_in" className="text-xs font-semibold uppercase tracking-wider text-muted-foreground whitespace-nowrap">
            Expires In
          </label>

          {!isCustom ? (
            <Select defaultValue="7d" onValueChange={(val) => {
              if (val === "custom") {
                setIsCustom(true);
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
                value={customExpiry}
                onChange={(e) => setCustomExpiry(e.target.value)}
                placeholder="e.g. 30m, 3d"
                className="rounded-none text-xs h-9 w-full px-3 bg-background border border-border"
                autoFocus
              />
              <Button
                type="button"
                variant="outline"
                size="icon"
                onClick={() => {
                  setIsCustom(false);
                  setCustomExpiry("");
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
          <Checkbox id="burn_after_download" className="rounded-none h-4.5 w-4.5" />
          <label
            htmlFor="burn_after_download"
            className="text-xs font-semibold uppercase tracking-wider text-muted-foreground cursor-pointer select-none whitespace-nowrap"
          >
            Self Destruct
          </label>
        </div>
      </div>

      <Button
        type="submit"
        className="w-full rounded-none uppercase text-xs font-bold tracking-wider h-10"
      >
        Create Filedrop
      </Button>
    </div>
  );
};

export default UploadCard;
