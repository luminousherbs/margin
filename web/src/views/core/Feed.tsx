import { useStore } from "@nanostores/react";
import { clsx } from "clsx";
import {
  Bookmark,
  Highlighter,
  MessageSquareText,
  User,
  Users,
} from "lucide-react";
import { useState } from "react";
import FeedItems from "../../components/feed/FeedItems";
import { Button, Tabs } from "../../components/ui";
import LayoutToggle from "../../components/ui/LayoutToggle";
import { $user } from "../../store/auth";
import { $feedLayout } from "../../store/feedLayout";
import type { AnnotationItem, UserProfile } from "../../types";

interface FeedProps {
  initialType?: string;
  initialTag?: string;
  initialUser?: UserProfile | null;
  motivation?: string;
  showTabs?: boolean;
  emptyMessage?: string;
  initialItems?: AnnotationItem[];
  initialHasMore?: boolean;
}

export default function Feed({
  initialType = "all",
  initialTag,
  initialUser,
  motivation,
  showTabs = true,
  emptyMessage = "No items found.",
  initialItems,
  initialHasMore,
}: FeedProps) {
  const [tag, setTag] = useState<string | undefined>(
    initialTag ||
      (typeof window !== "undefined"
        ? new URLSearchParams(window.location.search).get("tag") || undefined
        : undefined),
  );
  const storeUser = useStore($user);
  const user = storeUser || initialUser || null;
  const layout = useStore($feedLayout);
  const [activeTab, setActiveTab] = useState(initialType);
  const [activeFilter, setActiveFilter] = useState<string | undefined>(
    motivation,
  );
  const [mineOnly, setMineOnly] = useState(false);

  const clearTag = () => {
    setTag(undefined);
    const url = new URL(window.location.href);
    url.searchParams.delete("tag");
    window.history.replaceState({}, "", url.toString());
  };

  const handleTabChange = (id: string) => {
    if (id === activeTab) return;
    setActiveTab(id);
    clearTag();
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  const handleFilterChange = (id: string) => {
    const next = id === "all" ? undefined : id;
    if (next === activeFilter) return;
    setActiveFilter(next);
    clearTag();
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  const tabs = [
    { id: "all", label: "Recent" },
    { id: "popular", label: "Popular" },
    { id: "atmosphereconf", label: "ATmosphereConf" },
    { id: "shelved", label: "Shelved" },
    { id: "margin", label: "Margin" },
    { id: "semble", label: "Semble" },
  ];

  const filters = [
    { id: "all", label: "All", icon: null },
    { id: "commenting", label: "Annotations", icon: MessageSquareText },
    { id: "highlighting", label: "Highlights", icon: Highlighter },
    { id: "bookmarking", label: "Bookmarks", icon: Bookmark },
  ];

  return (
    <div className="mx-auto max-w-2xl xl:max-w-none">
      {!user && (
        <div className="text-center py-10 px-6 mb-4 animate-fade-in">
          <h1 className="text-2xl font-display font-bold mb-2 tracking-tight text-surface-900 dark:text-white">
            Welcome to Margin
          </h1>
          <p className="text-surface-500 dark:text-surface-400 mb-4 max-w-md mx-auto">
            Annotate, highlight, and bookmark anything on the web.
          </p>
          <div className="flex gap-3 justify-center">
            <Button onClick={() => (window.location.href = "/login")}>
              Get Started
            </Button>
            <Button
              variant="secondary"
              onClick={() => window.open("/about", "_blank")}
            >
              Learn More
            </Button>
          </div>
        </div>
      )}

      <div className="sticky top-0 z-10 bg-white/90 dark:bg-surface-800/90 backdrop-blur-md pb-3 mb-2 -mx-1 px-1 pt-2 space-y-2">
        {showTabs && !tag && (
          <Tabs tabs={tabs} activeTab={activeTab} onChange={handleTabChange} />
        )}
        {tag && (
          <div className="flex items-center justify-between mb-2">
            <h2 className="text-xl font-bold flex items-center gap-2">
              <span className="text-surface-500 font-normal">
                Items with tag:
              </span>
              <span className="bg-primary-50 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400 px-2 py-0.5 rounded-lg">
                #{tag}
              </span>
            </h2>
            <button
              onClick={clearTag}
              className="text-sm text-surface-500 hover:text-surface-900 dark:hover:text-white"
            >
              Clear filter
            </button>
          </div>
        )}
        {showTabs && (
          <div className="flex items-center gap-1.5 flex-wrap">
            {filters.map((f) => {
              const isActive =
                f.id === "all" ? !activeFilter : activeFilter === f.id;
              return (
                <button
                  key={f.id}
                  onClick={() => handleFilterChange(f.id)}
                  className={clsx(
                    "inline-flex items-center gap-1.5 px-3 py-1 text-xs font-medium rounded-full border transition-all",
                    isActive
                      ? "bg-primary-600 dark:bg-primary-500 text-white border-transparent shadow-sm"
                      : "bg-white dark:bg-surface-900 text-surface-500 dark:text-surface-400 border-surface-200 dark:border-surface-700 hover:border-primary-300 dark:hover:border-primary-700 hover:text-primary-600 dark:hover:text-primary-400",
                  )}
                >
                  {f.icon && <f.icon size={12} />}
                  {f.label}
                </button>
              );
            })}
            <div className="ml-auto">
              <LayoutToggle className="hidden sm:inline-flex" />
            </div>
          </div>
        )}
        {!showTabs && user && (
          <div className="flex items-center gap-1.5">
            {[
              { id: "everyone", label: "Everyone", icon: Users },
              { id: "mine", label: "Mine", icon: User },
            ].map((f) => {
              const isActive = f.id === "mine" ? mineOnly : !mineOnly;
              return (
                <button
                  key={f.id}
                  onClick={() => setMineOnly(f.id === "mine")}
                  className={clsx(
                    "inline-flex items-center gap-1.5 px-3 py-1 text-xs font-medium rounded-full border transition-all",
                    isActive
                      ? "bg-primary-600 dark:bg-primary-500 text-white border-transparent shadow-sm"
                      : "bg-white dark:bg-surface-900 text-surface-500 dark:text-surface-400 border-surface-200 dark:border-surface-700 hover:border-primary-300 dark:hover:border-primary-700 hover:text-primary-600 dark:hover:text-primary-400",
                  )}
                >
                  <f.icon size={12} />
                  {f.label}
                </button>
              );
            })}
            <div className="ml-auto">
              <LayoutToggle className="hidden sm:inline-flex" />
            </div>
          </div>
        )}
      </div>

      <FeedItems
        key={`${activeTab}-${activeFilter || "all"}-${tag || ""}-${mineOnly ? "mine" : "all"}`}
        type={activeTab === "atmosphereconf" ? "all" : activeTab}
        motivation={activeFilter}
        creator={mineOnly && user ? user.did : undefined}
        emptyMessage={emptyMessage}
        layout={layout}
        tag={
          activeTab === "atmosphereconf" ? "atmosphereconf" : tag?.toLowerCase()
        }
        initialItems={
          activeTab === initialType && activeFilter === motivation && !mineOnly
            ? initialItems
            : undefined
        }
        initialHasMore={
          activeTab === initialType && activeFilter === motivation && !mineOnly
            ? initialHasMore
            : undefined
        }
      />
    </div>
  );
}
