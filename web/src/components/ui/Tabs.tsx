import { clsx } from "clsx";
import React from "react";

interface Tab {
  id: string;
  label: string;
  badge?: number;
}

interface TabsProps {
  tabs: Tab[];
  activeTab: string;
  onChange: (id: string) => void;
  className?: string;
}

export default function Tabs({
  tabs,
  activeTab,
  onChange,
  className,
}: TabsProps) {
  return (
    <div
      className={clsx(
        "flex max-w-full overflow-x-auto gap-1 bg-surface-100 dark:bg-surface-800 p-1 rounded-lg w-fit",
        className,
      )}
    >
      {tabs.map((tab) => (
        <button
          key={tab.id}
          onClick={() => onChange(tab.id)}
          className={clsx(
            "px-3 py-1.5 text-sm font-medium rounded-md transition-all relative",
            activeTab === tab.id
              ? "bg-white dark:bg-surface-700 text-surface-900 dark:text-white shadow-sm"
              : "text-surface-500 dark:text-surface-400 hover:text-surface-700 dark:hover:text-surface-200",
          )}
        >
          {tab.label}
          {tab.badge !== undefined && tab.badge > 0 && (
            <span className="ml-1.5 inline-flex items-center justify-center min-w-[1.25rem] h-5 px-1.5 text-[10px] font-bold rounded-full bg-primary-600 text-white">
              {tab.badge > 99 ? "99+" : tab.badge}
            </span>
          )}
        </button>
      ))}
    </div>
  );
}
