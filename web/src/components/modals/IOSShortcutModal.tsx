import React from "react";
import { Button } from "../ui";
import {
  X,
  ExternalLink,
  Key,
  Share,
  Bookmark,
  PenTool,
  MoreHorizontal,
} from "lucide-react";
import { AppleIcon } from "../common/Icons";

interface IOSShortcutModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export default function IOSShortcutModal({
  isOpen,
  onClose,
}: IOSShortcutModalProps) {
  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm animate-fade-in overflow-y-auto"
      onClick={onClose}
    >
      <div
        className="bg-white dark:bg-surface-900 rounded-xl shadow-2xl max-w-lg w-full animate-scale-in ring-1 ring-surface-200 dark:ring-surface-700 overflow-hidden my-auto"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between px-6 py-4 border-b border-surface-200 dark:border-surface-800">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-lg bg-surface-100 dark:bg-surface-800 flex items-center justify-center text-surface-900 dark:text-white">
              <AppleIcon size={16} />
            </div>
            <h2 className="text-lg font-semibold text-surface-900 dark:text-white">
              Save from iOS Safari
            </h2>
          </div>
          <button
            onClick={onClose}
            className="p-2 text-surface-400 dark:text-surface-500 hover:bg-surface-100 dark:hover:bg-surface-800 rounded-lg transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        <div className="px-6 py-5 max-h-[70vh] overflow-y-auto space-y-6">
          <div className="rounded-xl overflow-hidden bg-surface-100 dark:bg-surface-800 ring-1 ring-surface-200 dark:ring-surface-700 aspect-[9/16] max-h-80 mx-auto flex items-center justify-center relative">
            <video
              src="/shortcut_walkthrough.mp4"
              autoPlay
              muted
              loop
              playsInline
              controls
              className="w-full h-full object-contain"
            />
          </div>

          <div className="space-y-4">
            <h3 className="text-sm font-semibold text-surface-900 dark:text-white uppercase tracking-wider">
              How to use the shortcut
            </h3>

            <div className="space-y-3">
              <div className="flex gap-3 text-sm">
                <div className="w-6 h-6 rounded-full bg-primary-100 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400 flex items-center justify-center flex-shrink-0 font-medium text-xs mt-0.5">
                  1
                </div>
                <div>
                  <p className="text-surface-900 dark:text-white font-medium">
                    Install the shortcut
                  </p>
                  <a
                    href="https://www.icloud.com/shortcuts/1e33ebf52f55431fae1e187cfe9738c3"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-1.5 mt-1.5 px-3 py-1.5 bg-surface-100 dark:bg-surface-800 hover:bg-surface-200 dark:hover:bg-surface-700 text-surface-900 dark:text-white rounded-lg text-xs font-medium transition-colors"
                  >
                    <ExternalLink size={14} /> Get iOS Shortcut
                  </a>
                </div>
              </div>

              <div className="flex gap-3 text-sm">
                <div className="w-6 h-6 rounded-full bg-primary-100 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400 flex items-center justify-center flex-shrink-0 font-medium text-xs mt-0.5">
                  2
                </div>
                <div>
                  <p className="text-surface-900 dark:text-white font-medium flex items-center gap-1.5">
                    Generate an API Key{" "}
                    <Key size={14} className="text-surface-400" />
                  </p>
                  <p className="text-surface-600 dark:text-surface-400 mt-0.5">
                    Create a new key on this settings page and copy it.
                  </p>
                </div>
              </div>

              <div className="flex gap-3 text-sm">
                <div className="w-6 h-6 rounded-full bg-primary-100 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400 flex items-center justify-center flex-shrink-0 font-medium text-xs mt-0.5">
                  3
                </div>
                <div>
                  <p className="text-surface-900 dark:text-white font-medium">
                    Configure the shortcut
                  </p>
                  <p className="text-surface-600 dark:text-surface-400 mt-0.5">
                    In the Shortcuts app, click the{" "}
                    <MoreHorizontal size={14} className="inline mx-0.5" /> menu
                    on the <strong>Save to Margin</strong> shortcut, and paste
                    your API key in the Text action right below the setup
                    comment.
                  </p>
                </div>
              </div>

              <div className="flex gap-3 text-sm">
                <div className="w-6 h-6 rounded-full bg-primary-100 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400 flex items-center justify-center flex-shrink-0 font-medium text-xs mt-0.5">
                  4
                </div>
                <div>
                  <p className="text-surface-900 dark:text-white font-medium flex items-center gap-1.5">
                    To Bookmark a page{" "}
                    <Bookmark size={14} className="text-surface-400" />
                  </p>
                  <p className="text-surface-600 dark:text-surface-400 mt-0.5">
                    Don't select any text. Click the{" "}
                    <MoreHorizontal size={14} className="inline mx-0.5" /> menu
                    in Safari, press{" "}
                    <Share size={12} className="inline mx-0.5" />{" "}
                    <strong>Share</strong>, and select{" "}
                    <strong>Save to Margin</strong>.
                  </p>
                </div>
              </div>

              <div className="flex gap-3 text-sm">
                <div className="w-6 h-6 rounded-full bg-primary-100 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400 flex items-center justify-center flex-shrink-0 font-medium text-xs mt-0.5">
                  5
                </div>
                <div>
                  <p className="text-surface-900 dark:text-white font-medium flex items-center gap-1.5">
                    To Highlight text{" "}
                    <PenTool size={14} className="text-surface-400" />
                  </p>
                  <p className="text-surface-600 dark:text-surface-400 mt-0.5">
                    Select text on the page, click the{" "}
                    <MoreHorizontal size={14} className="inline mx-0.5" /> menu,
                    press <strong>Share</strong>, and select{" "}
                    <strong>Save to Margin</strong>. Leave the Note field empty.
                  </p>
                </div>
              </div>

              <div className="flex gap-3 text-sm">
                <div className="w-6 h-6 rounded-full bg-primary-100 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400 flex items-center justify-center flex-shrink-0 font-medium text-xs mt-0.5">
                  6
                </div>
                <div>
                  <p className="text-surface-900 dark:text-white font-medium">
                    To Add an Annotation
                  </p>
                  <p className="text-surface-600 dark:text-surface-400 mt-0.5">
                    Select text, share to <strong>Save to Margin</strong> (via
                    the <MoreHorizontal size={14} className="inline mx-0.5" />{" "}
                    menu), enter your custom note in the Note field, and press
                    Done!
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div className="px-6 py-4 bg-surface-50 dark:bg-surface-800/50 border-t border-surface-200 dark:border-surface-800">
          <Button
            onClick={onClose}
            variant="primary"
            className="w-full justify-center"
          >
            Got it
          </Button>
        </div>
      </div>
    </div>
  );
}
