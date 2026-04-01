import { useStore } from "@nanostores/react";
import {
  Bell,
  Bookmark,
  Folder,
  Highlighter,
  Home,
  LogOut,
  MessageSquareText,
  MoreHorizontal,
  PenSquare,
  Search,
  Settings,
  User,
  X,
} from "lucide-react";
import React, { useEffect, useState } from "react";
import { getUnreadNotificationCount } from "../../api/client";
import { $user, logout } from "../../store/auth";
import { AppleIcon } from "../common/Icons";

interface MobileNavProps {
  currentPath?: string;
  onNavigate?: (path: string) => void;
}

export default function MobileNav({
  currentPath: initialPath,
  onNavigate,
}: MobileNavProps) {
  const user = useStore($user);
  const [currentPath, setCurrentPath] = useState(initialPath || "/");
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [unreadCount, setUnreadCount] = useState(0);

  const isAuthenticated = !!user;

  const isActive = (path: string) => {
    if (path === "/") return currentPath === "/";
    return currentPath.startsWith(path);
  };

  useEffect(() => {
    if (isAuthenticated) {
      getUnreadNotificationCount()
        .then((count) => setUnreadCount(count || 0))
        .catch(() => {});
    }
  }, [isAuthenticated]);

  const closeMenu = () => setIsMenuOpen(false);

  return (
    <>
      {isMenuOpen && (
        <div
          className="fixed inset-0 bg-black/50 z-40 md:hidden"
          onClick={closeMenu}
        />
      )}

      {isMenuOpen && (
        <div className="fixed bottom-16 left-0 right-0 bg-white dark:bg-surface-900 rounded-t-2xl shadow-2xl z-50 md:hidden animate-slide-up">
          <div className="p-4 space-y-1">
            {isAuthenticated && user ? (
              <>
                <a
                  href={`/profile/${user.did}`}
                  className="flex items-center gap-3 p-3 rounded-xl hover:bg-surface-100 dark:hover:bg-surface-800 transition-colors"
                  onClick={(e) => {
                    if (onNavigate) {
                      e.preventDefault();
                      onNavigate(`/profile/${user.did}`);
                    }
                    closeMenu();
                  }}
                >
                  {user.avatar ? (
                    <img
                      src={user.avatar}
                      alt=""
                      className="w-10 h-10 rounded-full object-cover"
                    />
                  ) : (
                    <div className="w-10 h-10 rounded-full bg-surface-200 dark:bg-surface-700 flex items-center justify-center">
                      <User size={18} className="text-surface-500" />
                    </div>
                  )}
                  <div className="flex flex-col">
                    <span className="font-semibold text-surface-900 dark:text-white">
                      {user.displayName || user.handle}
                    </span>
                    <span className="text-sm text-surface-500">
                      @{user.handle}
                    </span>
                  </div>
                </a>

                <div className="h-px bg-surface-200 dark:bg-surface-700 my-2" />

                {[
                  {
                    href: "/annotations",
                    icon: MessageSquareText,
                    label: "Annotations",
                  },
                  {
                    href: "/highlights",
                    icon: Highlighter,
                    label: "Highlights",
                  },
                  { href: "/bookmarks", icon: Bookmark, label: "Bookmarks" },
                  { href: "/collections", icon: Folder, label: "Collections" },
                  { href: "/settings", icon: Settings, label: "Settings" },
                ].map(({ href, icon: Icon, label }) => (
                  <a
                    key={href}
                    href={href}
                    className="flex items-center gap-3 p-3 rounded-xl hover:bg-surface-100 dark:hover:bg-surface-800 transition-colors text-surface-700 dark:text-surface-200"
                    onClick={(e) => {
                      if (onNavigate) {
                        e.preventDefault();
                        onNavigate(href);
                      }
                      closeMenu();
                    }}
                  >
                    <Icon size={20} />
                    <span>{label}</span>
                  </a>
                ))}

                <div className="h-px bg-surface-200 dark:bg-surface-700 my-2" />

                <a
                  href="https://www.icloud.com/shortcuts/1e33ebf52f55431fae1e187cfe9738c3"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center gap-3 p-3 rounded-xl hover:bg-surface-100 dark:hover:bg-surface-800 transition-colors text-surface-700 dark:text-surface-200"
                  onClick={closeMenu}
                >
                  <AppleIcon size={20} />
                  <span>iOS Shortcut</span>
                </a>

                <div className="h-px bg-surface-200 dark:bg-surface-700 my-2" />

                <button
                  className="flex items-center gap-3 p-3 rounded-xl hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors text-red-600 w-full"
                  onClick={() => {
                    logout();
                    closeMenu();
                  }}
                >
                  <LogOut size={20} />
                  <span>Log Out</span>
                </button>
              </>
            ) : (
              <>
                <a
                  href="/login"
                  className="flex items-center gap-3 p-3 rounded-xl hover:bg-surface-100 dark:hover:bg-surface-800 transition-colors text-surface-700 dark:text-surface-200"
                  onClick={closeMenu}
                >
                  <User size={20} />
                  <span>Sign In</span>
                </a>
                {[
                  { href: "/collections", icon: Folder, label: "Collections" },
                  { href: "/settings", icon: Settings, label: "Settings" },
                ].map(({ href, icon: Icon, label }) => (
                  <a
                    key={href}
                    href={href}
                    className="flex items-center gap-3 p-3 rounded-xl hover:bg-surface-100 dark:hover:bg-surface-800 transition-colors text-surface-700 dark:text-surface-200"
                    onClick={(e) => {
                      if (onNavigate) {
                        e.preventDefault();
                        onNavigate(href);
                      }
                      closeMenu();
                    }}
                  >
                    <Icon size={20} />
                    <span>{label}</span>
                  </a>
                ))}

                <div className="h-px bg-surface-200 dark:bg-surface-700 my-2" />

                <a
                  href="https://www.icloud.com/shortcuts/1e33ebf52f55431fae1e187cfe9738c3"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center gap-3 p-3 rounded-xl hover:bg-surface-100 dark:hover:bg-surface-800 transition-colors text-surface-700 dark:text-surface-200"
                  onClick={closeMenu}
                >
                  <AppleIcon size={20} />
                  <span>iOS Shortcut</span>
                </a>
              </>
            )}
          </div>
        </div>
      )}

      <nav className="fixed bottom-0 left-0 right-0 h-14 bg-white/90 dark:bg-surface-900/90 backdrop-blur-md border-t border-surface-200 dark:border-surface-700 flex items-center justify-around px-2 z-50 md:hidden safe-area-bottom">
        <a
          href="/home"
          className={`flex flex-col items-center justify-center w-14 h-14 rounded-xl transition-colors ${
            isActive("/home")
              ? "text-primary-600"
              : "text-surface-500 hover:text-surface-700"
          }`}
          onClick={(e) => {
            if (onNavigate) {
              e.preventDefault();
              onNavigate("/home");
            }
            setCurrentPath("/home");
            closeMenu();
          }}
        >
          <Home size={24} strokeWidth={1.5} />
        </a>

        <a
          href="/search"
          className={`flex flex-col items-center justify-center w-14 h-14 rounded-xl transition-colors ${
            isActive("/search")
              ? "text-primary-600"
              : "text-surface-500 hover:text-surface-700"
          }`}
          onClick={(e) => {
            if (onNavigate) {
              e.preventDefault();
              onNavigate("/search");
            }
            setCurrentPath("/search");
            closeMenu();
          }}
        >
          <Search size={24} strokeWidth={1.5} />
        </a>

        {isAuthenticated ? (
          <>
            <a
              href="/new"
              className="flex items-center justify-center w-12 h-12 rounded-full bg-primary-600 text-white shadow-lg hover:bg-primary-500 transition-colors -mt-4"
              onClick={(e) => {
                if (onNavigate) {
                  e.preventDefault();
                  onNavigate("/new");
                }
                setCurrentPath("/new");
                closeMenu();
              }}
            >
              <PenSquare size={20} strokeWidth={2} />
            </a>

            <a
              href="/notifications"
              className={`relative flex flex-col items-center justify-center w-14 h-14 rounded-xl transition-colors ${
                isActive("/notifications")
                  ? "text-primary-600"
                  : "text-surface-500 hover:text-surface-700"
              }`}
              onClick={(e) => {
                if (onNavigate) {
                  e.preventDefault();
                  onNavigate("/notifications");
                }
                setCurrentPath("/notifications");
                closeMenu();
              }}
            >
              <Bell size={24} strokeWidth={1.5} />
              {unreadCount > 0 && (
                <span className="absolute top-2 right-2 w-2 h-2 bg-red-500 rounded-full" />
              )}
            </a>
          </>
        ) : (
          <a
            href="/login"
            className="flex items-center justify-center w-12 h-12 rounded-full bg-primary-600 text-white shadow-lg hover:bg-primary-500 transition-colors -mt-4"
            onClick={closeMenu}
          >
            <User size={20} strokeWidth={2} />
          </a>
        )}

        <button
          className={`flex flex-col items-center justify-center w-14 h-14 rounded-xl transition-colors ${
            isMenuOpen
              ? "text-primary-600"
              : "text-surface-500 hover:text-surface-700"
          }`}
          onClick={() => setIsMenuOpen(!isMenuOpen)}
        >
          {isMenuOpen ? (
            <X size={24} strokeWidth={1.5} />
          ) : (
            <MoreHorizontal size={24} strokeWidth={1.5} />
          )}
        </button>
      </nav>
    </>
  );
}
