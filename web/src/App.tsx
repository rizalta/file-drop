import { useState, useEffect } from "react";
import UploadCard from "./components/UploadCard";
import DropCard from "./components/DropCard";

const App = () => {
  const [currentPath, setCurrentPath] = useState(window.location.pathname);
  const [currentSearch, setCurrentSearch] = useState(window.location.search);

  useEffect(() => {
    const handlePopState = () => {
      setCurrentPath(window.location.pathname);
      setCurrentSearch(window.location.search);
    };
    window.addEventListener("popstate", handlePopState);
    return () => window.removeEventListener("popstate", handlePopState);
  }, []);

  const searchParams = new URLSearchParams(currentSearch);
  const isShareMode = searchParams.get("share") === "true";
  const dropId = currentPath.startsWith("/f/") ? currentPath.split("/f/")[1] : null;

  const navigateTo = (path: string) => {
    window.history.pushState({}, "", path);
    setCurrentPath(window.location.pathname);
    setCurrentSearch(window.location.search);
  };

  return (
    <main className="min-h-screen w-full flex flex-col p-4 justify-center items-center sm:p-6 lg:p-8 bg-background text-foreground">
      <div className="w-full max-w-xl flex flex-col items-center space-y-6">
        <header className="flex flex-col items-center text-center">
          <button
            type="button"
            onClick={() => navigateTo("/")}
            className="flex items-center gap-3 cursor-pointer hover:opacity-85 transition-opacity group focus:outline-none"
          >
            <img src="/logo.svg" alt="File Drop Logo" className="w-10 h-10 group-hover:scale-105 transition-transform" />
            <h1 className="text-3xl font-extrabold tracking-tight">File Drop</h1>
          </button>
          <p className="text-sm text-muted-foreground mt-1">Drop Your Files...</p>
        </header>

        {dropId ? (
          <DropCard
            id={dropId}
            mode={isShareMode ? "share" : "download"}
            onReset={() => navigateTo("/")}
          />
        ) : (
          <UploadCard
            onSuccess={(id) => navigateTo(`/f/${id}?share=true`)}
          />
        )}
      </div>
    </main>
  );
};

export default App;
