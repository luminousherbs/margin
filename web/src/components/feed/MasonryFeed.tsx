import { useStore as useNanoStore, useStore } from "@nanostores/react";
import { Loader2 } from "lucide-react";
import React, { useEffect, useState } from "react";
import { getFeed } from "../../api/client";
import { $user } from "../../store/auth";
import { $feedLayout } from "../../store/feedLayout";
import type { AnnotationItem } from "../../types";
import Card from "../common/Card";
import { EmptyState, Tabs } from "../ui";
import LayoutToggle from "../ui/LayoutToggle";

interface MasonryFeedProps {
  motivation?: string;
  emptyMessage?: string;
  showTabs?: boolean;
  title?: string;
}

function MasonryContent({
  tab,
  motivation,
  emptyMessage,
  userDid,
  layout,
}: {
  tab: string;
  motivation?: string;
  emptyMessage: string;
  userDid?: string;
  layout: "list" | "mosaic";
}) {
  const [items, setItems] = useState<AnnotationItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;

    const params: { type?: string; motivation?: string; creator?: string } = {
      motivation,
    };

    if (tab === "my" && userDid) {
      params.creator = userDid;
      params.type = "my-feed";
    } else {
      params.type = "all";
    }

    getFeed(params)
      .then((data) => {
        if (cancelled) return;
        setItems(data.items);
        setLoading(false);
      })
      .catch((e) => {
        if (cancelled) return;
        console.error(e);
        setItems([]);
        setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [tab, motivation, userDid]);

  const handleDelete = (uri: string) => {
    setItems((prev) => prev.filter((i) => i.uri !== uri));
  };

  if (loading) {
    return (
      <div className="flex justify-center py-20">
        <Loader2
          className="animate-spin text-primary-600 dark:text-primary-400"
          size={32}
        />
      </div>
    );
  }

  if (items.length === 0) {
    return (
      <EmptyState
        message={
          tab === "my"
            ? emptyMessage
            : `No ${motivation === "bookmarking" ? "bookmarks" : "highlights"} from the community yet.`
        }
      />
    );
  }

  if (layout === "list") {
    return (
      <div className="space-y-3 animate-fade-in">
        {items.map((item) => (
          <Card
            key={item.uri || item.cid}
            item={item}
            onDelete={handleDelete}
          />
        ))}
      </div>
    );
  }

  return (
    <div className="columns-1 sm:columns-2 md:columns-3 xl:columns-4 gap-4 animate-fade-in">
      {items.map((item) => (
        <div key={item.uri || item.cid} className="break-inside-avoid mb-4">
          <Card item={item} onDelete={handleDelete} layout="mosaic" />
        </div>
      ))}
    </div>
  );
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

      <MasonryContent
        key={activeTab}
        tab={activeTab}
        motivation={motivation}
        emptyMessage={emptyMessage}
        userDid={user?.did}
        layout={layout}
      />
    </div>
  );
}
