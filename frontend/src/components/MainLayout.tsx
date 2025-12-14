import type { ReactNode } from "react";
import { Navbar } from "./Navbar";

interface MainLayoutProps {
  children: ReactNode;
}

export function MainLayout({ children }: MainLayoutProps) {
  return (
    <div className="min-h-screen flex flex-col bg-background">
      <Navbar />
      <main className="flex-1 w-full">
        <div className="max-w-5xl px-4 sm:px-6 py-6 mx-auto w-full">
          {children}
        </div>
      </main>
    </div>
  );
}
