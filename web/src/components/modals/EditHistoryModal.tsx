import React, { useState, useEffect } from "react";
import { X, Loader2, History } from "lucide-react";
import { useTranslation } from "react-i18next";
import { formatDistanceToNow } from "date-fns";
import type { AnnotationItem, EditHistoryItem } from "../../types";

interface EditHistoryModalProps {
  isOpen: boolean;
  onClose: () => void;
  item: AnnotationItem;
}

export default function EditHistoryModal({
  isOpen,
  onClose,
  item,
}: EditHistoryModalProps) {
  const { t } = useTranslation();
  const [history, setHistory] = useState<EditHistoryItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchHistory = async () => {
      if (!item.uri) return;

      try {
        setLoading(true);
        setError(null);
        const res = await fetch(
          `/api/notes/history?uri=${encodeURIComponent(item.uri)}`,
        );
        if (!res.ok) throw new Error("Failed to fetch history");
        const data = await res.json();
        setHistory(data);
      } catch (err) {
        console.error(err);
        setError(t("editHistory.failedLoad"));
      } finally {
        setLoading(false);
      }
    };

    if (isOpen && item.uri) {
      fetchHistory();
      document.body.style.overflow = "hidden";
    }
    return () => {
      document.body.style.overflow = "unset";
    };
  }, [isOpen, item.uri, t]);

  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fade-in"
      onClick={onClose}
    >
      <div
        className="w-full max-w-lg bg-white dark:bg-surface-900 rounded-3xl shadow-2xl overflow-hidden flex flex-col max-h-[80vh]"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="p-4 flex justify-between items-center border-b border-surface-100 dark:border-surface-800 shrink-0">
          <div className="flex items-center gap-2">
            <History className="text-surface-500" size={20} />
            <h2 className="text-xl font-display font-bold text-surface-900 dark:text-white">
              {t("editHistory.title")}
            </h2>
          </div>
          <button
            onClick={onClose}
            className="p-2 text-surface-400 hover:text-surface-900 dark:hover:text-white hover:bg-surface-50 dark:hover:bg-surface-800 rounded-full transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        <div className="p-0 overflow-y-auto flex-1 custom-scrollbar">
          {loading ? (
            <div className="flex justify-center p-8">
              <Loader2 className="animate-spin text-primary-500" size={32} />
            </div>
          ) : error ? (
            <div className="p-8 text-center text-red-500">{error}</div>
          ) : history.length === 0 ? (
            <div className="p-8 text-center text-surface-500">
              {t("editHistory.noHistory")}
            </div>
          ) : (
            <div className="divide-y divide-surface-100 dark:divide-surface-800">
              <div className="p-4 bg-primary-50/50 dark:bg-primary-900/10">
                <div className="flex justify-between items-start mb-2">
                  <span className="text-xs font-bold uppercase tracking-wider text-primary-600 dark:text-primary-400">
                    {t("editHistory.currentVersion")}
                  </span>
                  <span className="text-xs text-surface-400">
                    {item.editedAt
                      ? t("editHistory.editedAgo", {
                          time: formatDistanceToNow(new Date(item.editedAt)),
                        })
                      : t("editHistory.postedAgo", {
                          time: formatDistanceToNow(new Date(item.createdAt)),
                        })}
                  </span>
                </div>
                <div className="text-surface-900 dark:text-white whitespace-pre-wrap text-sm">
                  {item.text || item.body?.value}
                </div>
              </div>

              {history.map((edit, index) => (
                <div
                  key={edit.id || index}
                  className="p-4 hover:bg-surface-50 dark:hover:bg-surface-800/50 transition-colors"
                >
                  <div className="flex justify-between items-start mb-2">
                    <span className="text-xs font-medium text-surface-500">
                      {t("editHistory.previousVersion")}
                    </span>
                    <span
                      className="text-xs text-surface-400"
                      title={new Date(edit.editedAt).toLocaleString()}
                    >
                      {t("editHistory.timeAgo", {
                        time: formatDistanceToNow(new Date(edit.editedAt)),
                      })}
                    </span>
                  </div>
                  <div className="text-surface-600 dark:text-surface-300 whitespace-pre-wrap text-sm">
                    {edit.previousContent}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="p-4 border-t border-surface-100 dark:border-surface-800 bg-surface-50 dark:bg-surface-800/50 shrink-0">
          <button
            onClick={onClose}
            className="w-full py-2.5 bg-white dark:bg-surface-800 border border-surface-200 dark:border-surface-700 text-surface-700 dark:text-surface-200 font-medium rounded-xl hover:bg-surface-50 dark:hover:bg-surface-700 transition-colors"
          >
            {t("editHistory.close")}
          </button>
        </div>
      </div>
    </div>
  );
}
