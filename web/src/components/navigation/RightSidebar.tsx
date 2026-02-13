import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Search } from "lucide-react";
import { getTrendingTags, type Tag } from "../../api/client";

export default function RightSidebar() {
  const navigate = useNavigate();
  const [tags, setTags] = useState<Tag[]>([]);
  const [browser] = useState<"chrome" | "firefox" | "edge" | "other">(() => {
    if (typeof navigator === "undefined") return "other";
    const ua = navigator.userAgent;
    if (/Edg\//i.test(ua)) return "edge";
    if (/Firefox/i.test(ua)) return "firefox";
    if (/Chrome/i.test(ua)) return "chrome";
    return "other";
  });
  const [searchQuery, setSearchQuery] = useState("");

  const handleSearch = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && searchQuery.trim()) {
      navigate(`/url/${encodeURIComponent(searchQuery.trim())}`);
    }
  };

  useEffect(() => {
    getTrendingTags(10).then(setTags);
  }, []);

  const extensionLink =
    browser === "firefox"
      ? "https://addons.mozilla.org/en-US/firefox/addon/margin/"
      : browser === "edge"
        ? "https://microsoftedge.microsoft.com/addons/detail/margin/nfjnmllpdgcdnhmmggjihjbidmeadddn"
        : "https://chromewebstore.google.com/detail/margin/cgpmbiiagnehkikhcbnhiagfomajncpa";

  return (
    <aside className="hidden xl:block w-[320px] shrink-0 sticky top-0 h-screen overflow-y-auto px-6 py-6">
      <div className="space-y-5">
        <div className="relative">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <Search
              className="text-surface-400 dark:text-surface-500"
              size={15}
            />
          </div>
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            onKeyDown={handleSearch}
            placeholder="Search..."
            className="w-full bg-surface-100 dark:bg-surface-800/80 rounded-lg pl-9 pr-4 py-2 text-sm text-surface-900 dark:text-white placeholder:text-surface-400 dark:placeholder:text-surface-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 focus:bg-white dark:focus:bg-surface-800 transition-all border border-surface-200/60 dark:border-surface-700/60"
          />
        </div>

        <div className="rounded-xl p-4 bg-gradient-to-br from-primary-50 to-primary-100/50 dark:from-primary-950/30 dark:to-primary-900/10 border border-primary-200/40 dark:border-primary-800/30">
          <h3 className="font-semibold text-sm mb-1 text-surface-900 dark:text-white">
            Get the Extension
          </h3>
          <p className="text-surface-500 dark:text-surface-400 text-xs mb-3 leading-relaxed">
            Highlight, annotate, and bookmark from any page.
          </p>
          <a
            href={extensionLink}
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center justify-center w-full px-4 py-2 bg-primary-600 hover:bg-primary-700 dark:bg-primary-500 dark:hover:bg-primary-400 text-white dark:text-white rounded-lg transition-colors text-sm font-medium"
          >
            Download for{" "}
            {browser === "firefox"
              ? "Firefox"
              : browser === "edge"
                ? "Edge"
                : "Chrome"}
          </a>
        </div>

        <div>
          <h3 className="font-semibold text-sm px-1 mb-3 text-surface-900 dark:text-white tracking-tight">
            Trending
          </h3>
          {tags.length > 0 ? (
            <div className="flex flex-col">
              {tags.map((t) => (
                <a
                  key={t.tag}
                  href={`/home?tag=${encodeURIComponent(t.tag)}`}
                  className="px-2 py-2.5 hover:bg-surface-100 dark:hover:bg-surface-800 rounded-lg transition-colors group"
                >
                  <div className="font-semibold text-sm text-surface-900 dark:text-white group-hover:text-primary-600 dark:group-hover:text-primary-400 transition-colors">
                    #{t.tag}
                  </div>
                  <div className="text-xs text-surface-400 dark:text-surface-500 mt-0.5">
                    {t.count} {t.count === 1 ? "post" : "posts"}
                  </div>
                </a>
              ))}
            </div>
          ) : (
            <div className="px-2">
              <p className="text-sm text-surface-400 dark:text-surface-500">
                Nothing trending right now.
              </p>
            </div>
          )}
        </div>

        <div className="px-1 pt-2">
          <div className="flex flex-wrap gap-x-3 gap-y-1 text-[12px] text-surface-400 dark:text-surface-500 leading-relaxed">
            <a
              href="/about"
              className="hover:underline hover:text-surface-600 dark:hover:text-surface-300"
            >
              About
            </a>
            <a
              href="/privacy"
              className="hover:underline hover:text-surface-600 dark:hover:text-surface-300"
            >
              Privacy
            </a>
            <a
              href="/terms"
              className="hover:underline hover:text-surface-600 dark:hover:text-surface-300"
            >
              Terms
            </a>
            <a
              href="https://github.com/margin-at/margin"
              target="_blank"
              rel="noreferrer"
              className="hover:underline hover:text-surface-600 dark:hover:text-surface-300"
            >
              GitHub
            </a>
            <a
              href="https://tangled.org/margin.at/margin"
              target="_blank"
              rel="noreferrer"
              className="hover:underline hover:text-surface-600 dark:hover:text-surface-300"
            >
              Tangled
            </a>
            <span>© 2026 Margin</span>
          </div>
        </div>
      </div>
    </aside>
  );
}
