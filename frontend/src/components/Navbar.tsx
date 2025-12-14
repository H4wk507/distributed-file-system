import { useAuth } from "@/context/AuthContext";
import { useLogout } from "@/hooks/useAuth";
import { LogOut, Settings, User } from "lucide-react";
import { Link } from "react-router-dom";
import { Button } from "./ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "./ui/dropdown-menu";

export function Navbar() {
  const { user } = useAuth();
  const logout = useLogout();

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
