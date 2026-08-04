import UploadCard from "./components/UploadCard";

const App = () => {
  return (
    <main className="min-h-screen w-full flex flex-col p-4 justify-center items-center sm:p-6 lg:p-8 bg-background text-foreground">
      <div className="w-full max-w-xl flex flex-col items-center space-y-6">
        <header className="flex flex-col items-center text-center">
          <div className="flex items-center gap-3">
            <img src="/logo.svg" alt="File Drop Logo" className="w-10 h-10" />
            <h1 className="text-3xl font-extrabold tracking-tight">File Drop</h1>
          </div>
          <p className="text-sm text-muted-foreground">Drop Your Filess...</p>
        </header>
        <UploadCard />
      </div>
    </main>
  );
};

export default App;
