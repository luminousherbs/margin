import React, { useState } from "react";
import { Button } from "../ui";
import { ExternalLink, Shield } from "lucide-react";
import { addSkippedHostname } from "../../store/preferences";
import { useTranslation } from "react-i18next";

interface ExternalLinkModalProps {
  isOpen: boolean;
  onClose: () => void;
  url: string | null;
}

export default function ExternalLinkModal({
  isOpen,
  onClose,
  url,
}: ExternalLinkModalProps) {
  const { t } = useTranslation();
  const [dontAskAgain, setDontAskAgain] = useState(false);

  if (!isOpen || !url) return null;

  const displayUrl = url.split("#:~:text=")[0];

  const handleContinue = () => {
    if (dontAskAgain && url) {
      try {
        const hostname = new URL(url).hostname;
        addSkippedHostname(hostname);
      } catch (e) {
        console.error("Invalid URL", e);
      }
    }
    window.open(url, "_blank", "noopener,noreferrer");
    onClose();
  };

  const hostname = (() => {
    try {
      return new URL(url).hostname;
    } catch {
      return "this site";
    }
  })();

  return (
    <div
      className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm animate-fade-in"
      onClick={onClose}
    >
      <div
        className="bg-white dark:bg-surface-900 rounded-xl shadow-2xl max-w-md w-full animate-scale-in ring-1 ring-surface-200 dark:ring-surface-700 overflow-hidden"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="px-6 pt-6 pb-4">
          <div className="flex items-start gap-3">
            <div className="w-9 h-9 bg-primary-100 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400 rounded-lg flex items-center justify-center flex-shrink-0 mt-0.5">
              <Shield size={18} />
            </div>
            <div className="min-w-0">
              <h2 className="text-base font-semibold text-surface-900 dark:text-white">
                {t("externalLink.title")}
              </h2>
              <p className="text-sm text-surface-500 dark:text-surface-400 mt-1">
                {t("externalLink.message")}
              </p>
            </div>
          </div>

          <div className="mt-4 flex items-center gap-2 bg-surface-50 dark:bg-surface-800/60 border border-surface-200 dark:border-surface-700 rounded-lg px-3 py-2.5">
            <ExternalLink
              size={14}
              className="text-surface-400 dark:text-surface-500 flex-shrink-0"
            />
            <span className="text-sm text-surface-700 dark:text-surface-300 break-all line-clamp-2">
              {displayUrl}
            </span>
          </div>
        </div>

        <div className="px-6 pb-5 pt-2 flex flex-col gap-3">
          <label className="flex items-center gap-2 cursor-pointer select-none group">
            <input
              type="checkbox"
              checked={dontAskAgain}
              onChange={(e) => setDontAskAgain(e.target.checked)}
              className="rounded border-surface-300 dark:border-surface-600 text-primary-600 focus:ring-primary-500 w-3.5 h-3.5 cursor-pointer"
            />
            <span className="text-xs text-surface-500 dark:text-surface-400 group-hover:text-surface-600 dark:group-hover:text-surface-300 transition-colors">
              {t("externalLink.alwaysAllow", { hostname })}
            </span>
          </label>

          <div className="flex gap-2">
            <Button
              onClick={onClose}
              variant="ghost"
              className="flex-1 justify-center"
            >
              {t("externalLink.cancel")}
            </Button>
            <Button
              onClick={handleContinue}
              variant="primary"
              className="flex-1 justify-center"
              icon={<ExternalLink size={14} />}
            >
              {t("externalLink.open")}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
