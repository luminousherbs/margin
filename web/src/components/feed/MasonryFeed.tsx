import { useStore as useNanoStore, useStore } from "@nanostores/react";
import { useState } from "react";
import { $user } from "../../store/auth";
import { $feedLayout } from "../../store/feedLayout";
import { Tabs } from "../ui";
import LayoutToggle from "../ui/LayoutToggle";
import FeedItems from "./FeedItems";

interface MasonryFeedProps {
  motivation?: string;
  emptyMessage?: string;
  showTabs?: boolean;
  title?: string;
}

export default function MasonryFeed({
  motivation,
  emptyMessage = "No items found.",
  showTabs = false,
  title,
}: MasonryFeedProps) {
  const user = useStore($user);
  const layout = useNanoStore($feedLayout);
  const [activeTab, setActiveTab] = useState(user ? "my" : "global");

  const handleTabChange = (id: string) => {
    if (id === activeTab) return;
    setActiveTab(id);
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  const tabs = user
    ? [
        { id: "my", label: "My" },
        { id: "global", label: "Global" },
      ]
    : [{ id: "global", label: "Global" }];

  const creator = activeTab === "my" ? user?.did : undefined;
  const type = activeTab === "my" ? "my-feed" : "all";

  return (
    <div className="mx-auto max-w-2xl xl:max-w-none">
      {title && (
        <h1 className="text-3xl font-display font-bold text-surface-900 dark:text-white mb-6 text-center lg:text-left">
          {title}
        </h1>
      )}

      {showTabs && (
        <div className="sticky top-0 z-10 bg-white/95 dark:bg-surface-800/95 backdrop-blur-sm pb-4 mb-2 -mx-1 px-1 pt-1">
          <div className="flex items-center gap-3">
            <div className="flex-1">
              <Tabs
                tabs={tabs}
                activeTab={activeTab}
                onChange={handleTabChange}
              />
            </div>
            <LayoutToggle className="hidden sm:inline-flex" />
          </div>
        </div>
      )}

      {!showTabs && (
        <div className="flex justify-end mb-4">
          <LayoutToggle className="hidden sm:inline-flex" />
        </div>
      )}

      <FeedItems
        key={activeTab}
        type={type}
        motivation={motivation}
        emptyMessage={
          activeTab === "my"
            ? emptyMessage
            : `No ${motivation === "bookmarking" ? "bookmarks" : "highlights"} from the community yet.`
        }
        creator={creator}
        layout={layout}
      />
    </div>
  );
}
