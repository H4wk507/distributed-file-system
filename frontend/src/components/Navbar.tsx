import { useAuth } from "@/context/AuthContext";
import { useLogout } from "@/hooks/useAuth";
import { Activity, Files, LogOut, Server, Settings, User } from "lucide-react";
import { Link, useLocation } from "react-router-dom";
import { Button } from "./ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "./ui/dropdown-menu";

const navLinks = [
  { href: "/", label: "pliki", icon: Files },
  { href: "/nodes", label: "węzły", icon: Server },
  { href: "/monitoring", label: "monitoring", icon: Activity },
];

export function Navbar() {
  const { user } = useAuth();
  const logout = useLogout();
  const location = useLocation();

  return (
    <header className="border-b border-border bg-card sticky top-0 z-50">
      <div className="px-4 sm:px-6">
        <div className="flex items-center justify-between h-12">
          {/* Logo - asymetryczny, surowy */}
          <Link
            to="/"
            className="flex items-center gap-2 hover:text-primary transition-colors"
          >
            <span className="text-primary font-mono text-sm">[RSP]</span>
            <span className="hidden sm:inline text-xs text-muted-foreground font-mono">
              rozproszony system plików
            </span>
          </Link>

          {/* Nawigacja - styl terminala */}
          <nav className="hidden sm:flex items-center">
            {navLinks.map((link, idx) => {
              const isActive = location.pathname === link.href;
              return (
                <Link
                  key={link.href}
                  to={link.href}
                  className={`flex items-center gap-1.5 px-3 py-1.5 text-xs font-mono uppercase tracking-wide transition-colors border-l border-border ${
                    isActive
                      ? "text-primary bg-primary/10"
                      : "text-muted-foreground hover:text-foreground hover:bg-muted"
                  } ${idx === navLinks.length - 1 ? "border-r" : ""}`}
                >
                  <link.icon className="w-3.5 h-3.5" />
                  {link.label}
                </Link>
              );
            })}
          </nav>

          {/* User menu - minimalny */}
          <div className="flex items-center gap-3">
            <span className="hidden md:inline text-xs text-muted-foreground font-mono">
              user: {user?.username}
            </span>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" className="gap-1.5 px-2 h-8">
                  <User className="w-4 h-4" />
                  <span className="sm:hidden text-xs">{user?.username}</span>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-48">
                <DropdownMenuLabel>
                  <div className="flex flex-col font-mono">
                    <span className="text-xs">{user?.username}</span>
                    <span className="text-xs text-muted-foreground">
                      {user?.email}
                    </span>
                  </div>
                </DropdownMenuLabel>
                <DropdownMenuSeparator />
                <div className="sm:hidden">
                  {navLinks.map((link) => (
                    <DropdownMenuItem key={link.href} asChild>
                      <Link
                        to={link.href}
                        className="flex items-center gap-2 text-xs"
                      >
                        <link.icon className="w-3.5 h-3.5" />
                        {link.label}
                      </Link>
                    </DropdownMenuItem>
                  ))}
                  <DropdownMenuSeparator />
                </div>

                <DropdownMenuItem className="text-xs">
                  <Settings className="w-3.5 h-3.5 mr-2" />
                  ustawienia
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  onClick={logout}
                  className="text-destructive focus:text-destructive text-xs"
                >
                  <LogOut className="w-3.5 h-3.5 mr-2" />
                  wyloguj
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
      </div>
    </header>
  );
}
