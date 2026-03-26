import React, { useEffect, useState } from "react";
import {
  Home,
  Bookmark,
  Settings,
  LogOut,
  Bell,
  Sun,
  Moon,
  Monitor,
  Folder,
  LogIn,
  PenSquare,
  MessageSquareText,
  Highlighter,
  Compass,
} from "lucide-react";
import { useStore } from "@nanostores/react";
import { $user, logout } from "../../store/auth";
import { $theme, cycleTheme } from "../../store/theme";
import { getUnreadNotificationCount } from "../../api/client";
import { Avatar, CountBadge } from "../ui";
import type { UserProfile } from "../../types";

interface SidebarProps {
  initialUser?: UserProfile | null;
  currentPath?: string;
}

export default function Sidebar({
  initialUser,
  currentPath: initialPath,
}: SidebarProps) {
  const storeUser = useStore($user);
  const user = storeUser || initialUser || null;
  const theme = useStore($theme);
  const [currentPath, setCurrentPath] = useState(initialPath || "/");
  const [unreadCount, setUnreadCount] = useState(0);

  useEffect(() => {
    if (initialUser && !storeUser) {
      $user.set(initialUser);
    }
  }, [initialUser, storeUser]);

  useEffect(() => {
    const handler = () => setCurrentPath(window.location.pathname);
    document.addEventListener("astro:page-load", handler);
    return () => document.removeEventListener("astro:page-load", handler);
  }, []);

  const handleNav = (href: string) => () => {
    setCurrentPath(href);
  };

  useEffect(() => {
    if (!user) return;

    const checkNotifications = async () => {
      const count = await getUnreadNotificationCount();
      setUnreadCount(count);
    };

    checkNotifications();
    const interval = setInterval(checkNotifications, 30000);
    return () => clearInterval(interval);
  }, [user]);

  const publicNavItems = [
    { icon: Home, label: "Feed", href: "/home", badge: undefined },
    { icon: Compass, label: "Discover", href: "/discover", badge: undefined },
    {
      icon: MessageSquareText,
      label: "Annotations",
      href: "/annotations",
      badge: undefined,
    },
    {
      icon: Highlighter,
      label: "Highlights",
      href: "/highlights",
      badge: undefined,
    },
    {
      icon: Bookmark,
      label: "Bookmarks",
      href: "/bookmarks",
      badge: undefined,
    },
  ];

  const authNavItems = [
    { icon: Home, label: "Feed", href: "/home" },
    { icon: Compass, label: "Discover", href: "/discover" },
    {
      icon: Bell,
      label: "Activity",
      href: "/notifications",
      badge: unreadCount,
    },
    { icon: MessageSquareText, label: "Annotations", href: "/annotations" },
    { icon: Highlighter, label: "Highlights", href: "/highlights" },
    { icon: Bookmark, label: "Bookmarks", href: "/bookmarks" },
    { icon: Folder, label: "Collections", href: "/collections" },
  ];

  const navItems = user ? authNavItems : publicNavItems;

  return (
    <aside className="sticky top-0 h-screen hidden md:flex flex-col justify-between py-6 px-2 lg:px-4 z-50 w-[68px] lg:w-[260px] transition-all duration-200">
      <div className="flex flex-col gap-6">
        <a
          href="/home"
          className="px-3 hover:opacity-80 transition-opacity w-fit flex items-center gap-2.5"
        >
          <img src="/logo.svg" alt="Margin" className="w-8 h-8" />
        </a>

        <nav className="flex flex-col gap-0.5">
          {navItems.map((item) => {
            const isActive =
              currentPath === item.href ||
              (item.href !== "/home" && currentPath.startsWith(item.href));
            return (
              <a
                key={item.href}
                href={item.href}
                title={item.label}
                onClick={handleNav(item.href)}
                data-astro-prefetch="viewport"
                className={`flex items-center justify-center lg:justify-start gap-3 px-0 lg:px-3 py-2.5 rounded-lg transition-all duration-150 text-[14px] group ${
                  isActive
                    ? "font-semibold text-primary-700 dark:text-primary-300 bg-primary-50 dark:bg-primary-950/40"
                    : "font-medium text-surface-600 dark:text-surface-400 hover:bg-surface-100 dark:hover:bg-surface-800 hover:text-surface-900 dark:hover:text-white"
                }`}
              >
                <item.icon
                  size={20}
                  className={`transition-colors ${isActive ? "text-primary-600 dark:text-primary-400" : ""}`}
                  strokeWidth={isActive ? 2.25 : 1.75}
                />
                <span className="flex-1 hidden lg:inline">{item.label}</span>
                {(item.badge ?? 0) > 0 && (
                  <CountBadge count={item.badge ?? 0} />
                )}
              </a>
            );
          })}

          {user && (
            <a
              href="/new"
              title="New annotation"
              className="flex items-center justify-center lg:justify-start gap-3 px-0 lg:px-3 py-2.5 mt-2 rounded-lg bg-primary-600 dark:bg-primary-500 text-white hover:bg-primary-700 dark:hover:bg-primary-400 transition-colors text-[14px] font-semibold"
            >
              <PenSquare size={20} strokeWidth={1.75} />
              <span className="hidden lg:inline">New</span>
            </a>
          )}
        </nav>
      </div>

      <div className="space-y-1">
        <button
          onClick={cycleTheme}
          title={
            theme === "light" ? "Light" : theme === "dark" ? "Dark" : "System"
          }
          className="flex items-center justify-center lg:justify-start gap-3 px-0 lg:px-3 py-2.5 rounded-lg hover:bg-surface-100 dark:hover:bg-surface-800 text-[13px] font-medium text-surface-500 dark:text-surface-400 w-full transition-colors"
        >
          {theme === "light" ? (
            <Sun size={18} />
          ) : theme === "dark" ? (
            <Moon size={18} />
          ) : (
            <Monitor size={18} />
          )}
          <span className="hidden lg:inline">
            {theme === "light" ? "Light" : theme === "dark" ? "Dark" : "System"}
          </span>
        </button>

        {user ? (
          <>
            <a
              href="/settings"
              title="Settings"
              className="flex items-center justify-center lg:justify-start gap-3 px-0 lg:px-3 py-2.5 rounded-lg hover:bg-surface-100 dark:hover:bg-surface-800 text-[13px] font-medium text-surface-500 dark:text-surface-400 transition-colors"
            >
              <Settings size={18} />
              <span className="hidden lg:inline">Settings</span>
            </a>

            <div className="h-px bg-surface-200/60 dark:bg-surface-800/60 my-2" />

            <a
              href={`/profile/${user.did}`}
              title={user.displayName || user.handle}
              className="flex items-center justify-center lg:justify-start gap-2.5 p-2 rounded-lg hover:bg-surface-100 dark:hover:bg-surface-800 transition-colors w-full"
            >
              <Avatar did={user.did} avatar={user.avatar} size="sm" />
              <div className="flex-1 min-w-0 hidden lg:block">
                <p className="font-medium text-surface-900 dark:text-white truncate text-[13px]">
                  {user.displayName || user.handle}
                </p>
                <p className="text-[11px] text-surface-500 dark:text-surface-400 truncate">
                  @{user.handle}
                </p>
              </div>
            </a>

            <button
              onClick={logout}
              title="Log out"
              className="flex items-center justify-center lg:justify-start gap-3 px-0 lg:px-3 py-2 rounded-lg hover:bg-red-50 dark:hover:bg-red-900/20 text-[13px] font-medium text-surface-400 dark:text-surface-500 hover:text-red-600 dark:hover:text-red-400 w-full text-left transition-colors"
            >
              <LogOut size={16} />
              <span className="hidden lg:inline">Log out</span>
            </button>
          </>
        ) : (
          <>
            <div className="h-px bg-surface-200/60 dark:bg-surface-800/60 my-2" />

            <a
              href="/login"
              title="Sign in"
              className="flex items-center justify-center lg:justify-start gap-3 px-0 lg:px-3 py-2.5 rounded-lg bg-primary-50 dark:bg-primary-950/40 text-primary-700 dark:text-primary-300 hover:bg-primary-100 dark:hover:bg-primary-950/60 text-[13px] font-semibold transition-colors"
            >
              <LogIn size={18} />
              <span className="hidden lg:inline">Sign in</span>
            </a>
          </>
        )}
      </div>
    </aside>
  );
}
