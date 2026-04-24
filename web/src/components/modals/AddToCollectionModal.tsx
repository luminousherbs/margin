import React, { useState, useEffect, useCallback } from "react";
import { useTranslation } from "react-i18next";
import {
  X,
  Plus,
  Check,
  Loader2,
  ChevronRight,
  FolderPlus,
} from "lucide-react";
import CollectionIcon from "../common/CollectionIcon";
import { ICON_MAP } from "../common/iconMap";
import { Theme } from "emoji-picker-react";
const EmojiPicker = React.lazy(() => import("emoji-picker-react"));
import { useStore } from "@nanostores/react";
import { $user } from "../../store/auth";
import { $theme } from "../../store/theme";
import { analytics } from "../../lib/analytics";
import {
  getCollections,
  addCollectionItem,
  createCollection,
  getCollectionsContaining,
  type Collection,
} from "../../api/client";

interface AddToCollectionModalProps {
  isOpen: boolean;
  onClose: () => void;
  annotationUri: string;
}

export default function AddToCollectionModal({
  isOpen,
  onClose,
  annotationUri,
}: AddToCollectionModalProps) {
  const { t } = useTranslation();
  const user = useStore($user);
  const theme = useStore($theme);
  const [collections, setCollections] = useState<Collection[]>([]);
  const [loading, setLoading] = useState(true);
  const [addingTo, setAddingTo] = useState<string | null>(null);
  const [addedTo, setAddedTo] = useState<Set<string>>(new Set());
  const [error, setError] = useState<string | null>(null);

  const [showNewForm, setShowNewForm] = useState(false);
  const [newName, setNewName] = useState("");
  const [newDescription, setNewDescription] = useState("");
  const [newIcon, setNewIcon] = useState("");
  const [activeTab, setActiveTab] = useState<"icon" | "emoji">("icon");
  const [creating, setCreating] = useState(false);

  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = "hidden";
    }
    return () => {
      document.body.style.overflow = "unset";
    };
  }, [isOpen]);

  const loadCollections = useCallback(async () => {
    if (!user) return;
    try {
      setLoading(true);
      const data = await getCollections(user.did);
      setCollections(data);
    } catch (err) {
      console.error(err);
      setError(t("addToCollection.failedLoad"));
    } finally {
      setLoading(false);
    }
  }, [user, t]);

  useEffect(() => {
    if (isOpen && user) {
      loadCollections();
      setError(null);
      getCollectionsContaining(annotationUri).then((uris) => {
        setAddedTo(new Set(uris));
      });
    }
  }, [isOpen, user, loadCollections, annotationUri]);

  const handleAdd = async (collectionUri: string) => {
    if (addedTo.has(collectionUri)) return;

    try {
      setAddingTo(collectionUri);
      await addCollectionItem(collectionUri, annotationUri);
      setAddedTo((prev) => new Set([...prev, collectionUri]));
      analytics.capture("item_added_to_collection");
    } catch (err) {
      console.error(err);
      setError(t("addToCollection.failedAdd"));
    } finally {
      setAddingTo(null);
    }
  };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newName.trim()) return;
    try {
      setCreating(true);
      const iconValue = newIcon
        ? ICON_MAP[newIcon]
          ? `icon:${newIcon}`
          : newIcon
        : undefined;
      const newCollection = await createCollection(
        newName.trim(),
        newDescription.trim() || undefined,
        iconValue,
      );
      if (newCollection) {
        setCollections((prev) => [newCollection, ...prev]);
        setNewName("");
        setNewDescription("");
        setNewIcon("");
        setShowNewForm(false);
      }
    } catch (err) {
      console.error(err);
      setError(t("addToCollection.failedCreate"));
    } finally {
      setCreating(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fade-in"
      onClick={onClose}
    >
      <div
        className="w-full max-w-md bg-white dark:bg-surface-900 rounded-3xl shadow-2xl overflow-hidden"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="p-4 flex justify-between items-center border-b border-surface-100 dark:border-surface-800">
          <h2 className="text-xl font-display font-bold text-surface-900 dark:text-white">
            {t("addToCollection.title")}
          </h2>
          <button
            onClick={onClose}
            className="p-2 text-surface-400 hover:text-surface-900 dark:hover:text-white hover:bg-surface-50 dark:hover:bg-surface-800 rounded-full transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        <div className="px-6 pb-6 pt-4">
          {loading ? (
            <div className="text-center py-10">
              <Loader2
                size={32}
                className="animate-spin text-primary-600 dark:text-primary-400 mx-auto mb-3"
              />
              <p className="text-surface-500 dark:text-surface-400 font-medium">
                {t("addToCollection.loading")}
              </p>
            </div>
          ) : showNewForm ? (
            <form onSubmit={handleCreate} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-surface-700 dark:text-surface-300 mb-1">
                  {t("addToCollection.collectionNameLabel")}
                </label>
                <input
                  type="text"
                  className="w-full px-4 py-3 bg-surface-50 dark:bg-surface-800 border border-surface-200 dark:border-surface-700 rounded-xl focus:border-primary-500 dark:focus:border-primary-400 focus:ring-4 focus:ring-primary-500/10 outline-none transition-all text-surface-900 dark:text-white placeholder-surface-400 dark:placeholder-surface-500"
                  value={newName}
                  onChange={(e) => setNewName(e.target.value)}
                  placeholder={t("addToCollection.namePlaceholder")}
                  autoFocus
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-surface-700 dark:text-surface-300 mb-1">
                  {t("addToCollection.descriptionLabel")}
                </label>
                <textarea
                  className="w-full px-4 py-3 bg-surface-50 dark:bg-surface-800 border border-surface-200 dark:border-surface-700 rounded-xl focus:border-primary-500 dark:focus:border-primary-400 focus:ring-4 focus:ring-primary-500/10 outline-none transition-all text-surface-900 dark:text-white placeholder-surface-400 dark:placeholder-surface-500 resize-none"
                  value={newDescription}
                  onChange={(e) => setNewDescription(e.target.value)}
                  placeholder={t("addToCollection.descriptionPlaceholder")}
                  rows={2}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-surface-700 dark:text-surface-300 mb-2">
                  {t("addToCollection.iconLabel")}
                </label>

                <div className="flex gap-2 mb-3 bg-surface-100 dark:bg-surface-800 p-1 rounded-xl">
                  <button
                    type="button"
                    onClick={() => setActiveTab("icon")}
                    className={`flex-1 py-1.5 text-sm font-medium rounded-lg transition-colors ${
                      activeTab === "icon"
                        ? "bg-white dark:bg-surface-700 text-surface-900 dark:text-white shadow-sm"
                        : "text-surface-600 dark:text-surface-400 hover:text-surface-900 dark:hover:text-surface-200"
                    }`}
                  >
                    {t("addToCollection.iconsTab")}
                  </button>
                  <button
                    type="button"
                    onClick={() => setActiveTab("emoji")}
                    className={`flex-1 py-1.5 text-sm font-medium rounded-lg transition-colors ${
                      activeTab === "emoji"
                        ? "bg-white dark:bg-surface-700 text-surface-900 dark:text-white shadow-sm"
                        : "text-surface-600 dark:text-surface-400 hover:text-surface-900 dark:hover:text-surface-200"
                    }`}
                  >
                    {t("addToCollection.emojisTab")}
                  </button>
                </div>

                {activeTab === "icon" ? (
                  <div className="grid grid-cols-8 gap-1.5 max-h-60 overflow-y-auto p-2 bg-surface-50 dark:bg-surface-800 rounded-xl border border-surface-200 dark:border-surface-700 custom-scrollbar">
                    {Object.keys(ICON_MAP).map((iconName) => {
                      const isSelected = newIcon === iconName;
                      return (
                        <button
                          key={iconName}
                          type="button"
                          onClick={() => setNewIcon(isSelected ? "" : iconName)}
                          className={`w-8 h-8 flex items-center justify-center rounded-lg transition-all ${
                            isSelected
                              ? "bg-primary-600 text-white"
                              : "hover:bg-surface-200 dark:hover:bg-surface-700 text-surface-600 dark:text-surface-400"
                          }`}
                          title={iconName}
                        >
                          <CollectionIcon icon={`icon:${iconName}`} size={16} />
                        </button>
                      );
                    })}
                  </div>
                ) : (
                  <div className="w-full bg-surface-50 dark:bg-surface-800 rounded-xl border border-surface-200 dark:border-surface-700 overflow-hidden">
                    <React.Suspense
                      fallback={
                        <div className="flex items-center justify-center h-[300px]">
                          <Loader2
                            className="animate-spin text-surface-400"
                            size={24}
                          />
                        </div>
                      }
                    >
                      <EmojiPicker
                        className="custom-emoji-picker"
                        onEmojiClick={(emojiData) =>
                          setNewIcon(emojiData.emoji)
                        }
                        autoFocusSearch={false}
                        width="100%"
                        height={300}
                        previewConfig={{ showPreview: false }}
                        skinTonesDisabled
                        lazyLoadEmojis
                        theme={
                          theme === "dark" ||
                          (theme === "system" &&
                            window.matchMedia("(prefers-color-scheme: dark)")
                              .matches)
                            ? (Theme.DARK as Theme)
                            : (Theme.LIGHT as Theme)
                        }
                      />
                    </React.Suspense>
                  </div>
                )}

                {newIcon && (
                  <p className="mt-2 text-sm text-surface-600 dark:text-surface-300 flex items-center gap-2">
                    {t("addToCollection.selected")}
                    <span className="inline-flex items-center justify-center w-8 h-8 bg-surface-100 dark:bg-surface-800 rounded-lg border border-surface-200 dark:border-surface-700">
                      <CollectionIcon
                        icon={ICON_MAP[newIcon] ? `icon:${newIcon}` : newIcon}
                        size={18}
                      />
                    </span>
                  </p>
                )}
              </div>

              {error && (
                <div className="p-3 bg-red-50 dark:bg-red-900/30 text-red-600 dark:text-red-400 text-sm rounded-lg">
                  {error}
                </div>
              )}

              <div className="flex gap-3 pt-2">
                <button
                  type="button"
                  className="flex-1 py-3 bg-white dark:bg-surface-800 border border-surface-200 dark:border-surface-700 text-surface-700 dark:text-surface-200 font-semibold rounded-xl hover:bg-surface-50 dark:hover:bg-surface-700 transition-colors"
                  onClick={() => {
                    setShowNewForm(false);
                    setNewDescription("");
                    setNewIcon("");
                    setError(null);
                  }}
                >
                  {t("addToCollection.back")}
                </button>
                <button
                  type="submit"
                  className="flex-1 py-3 bg-primary-600 text-white font-semibold rounded-xl hover:bg-primary-700 transition-colors disabled:opacity-50 flex items-center justify-center gap-2"
                  disabled={!newName.trim() || creating}
                >
                  {creating && <Loader2 size={16} className="animate-spin" />}
                  {creating
                    ? t("addToCollection.creating")
                    : t("addToCollection.create")}
                </button>
              </div>
            </form>
          ) : (
            <div>
              {error && (
                <div className="mb-4 p-3 bg-red-50 dark:bg-red-900/30 text-red-600 dark:text-red-400 text-sm rounded-lg">
                  {error}
                </div>
              )}

              <button
                className="w-full flex items-center gap-4 p-4 bg-white dark:bg-surface-800 border-2 border-primary-100 dark:border-primary-900/50 hover:border-primary-300 dark:hover:border-primary-700 rounded-2xl shadow-sm hover:shadow-md transition-all group text-left mb-4"
                onClick={() => setShowNewForm(true)}
              >
                <div className="w-10 h-10 bg-primary-50 dark:bg-primary-900/30 rounded-full flex items-center justify-center text-primary-600 dark:text-primary-400 flex-shrink-0">
                  <FolderPlus size={20} />
                </div>
                <div className="flex-1 min-w-0">
                  <h3 className="font-bold text-surface-900 dark:text-white group-hover:text-primary-700 dark:group-hover:text-primary-400 transition-colors">
                    {t("addToCollection.newCollectionButton")}
                  </h3>
                  <span className="text-sm text-surface-500 dark:text-surface-400">
                    {t("addToCollection.createNewDescription")}
                  </span>
                </div>
                <ChevronRight
                  size={20}
                  className="text-surface-300 dark:text-surface-600 group-hover:text-primary-500 dark:group-hover:text-primary-400"
                />
              </button>

              {collections.length === 0 ? (
                <div className="text-center py-6">
                  <p className="text-surface-500 dark:text-surface-400">
                    {t("addToCollection.none")}
                  </p>
                </div>
              ) : (
                <div className="space-y-2 max-h-[300px] overflow-y-auto">
                  {collections.map((col) => {
                    const isAdded = addedTo.has(col.uri);
                    const isAdding = addingTo === col.uri;

                    return (
                      <button
                        key={col.uri}
                        onClick={() => handleAdd(col.uri)}
                        disabled={isAdding || isAdded}
                        className="w-full flex items-center gap-3 p-3 bg-surface-50 dark:bg-surface-800/50 hover:bg-surface-100 dark:hover:bg-surface-800 rounded-xl transition-colors text-left group disabled:opacity-70"
                      >
                        <div className="w-8 h-8 flex items-center justify-center bg-white dark:bg-surface-700 rounded-full shadow-sm text-surface-600 dark:text-surface-300">
                          <CollectionIcon icon={col.icon} size={18} />
                        </div>
                        <div className="flex-1 min-w-0">
                          <h3 className="text-sm font-bold text-surface-900 dark:text-white">
                            {col.name}
                          </h3>
                          {col.description && (
                            <p className="text-xs text-surface-500 dark:text-surface-400 line-clamp-1">
                              {col.description}
                            </p>
                          )}
                        </div>
                        {isAdding ? (
                          <Loader2
                            size={16}
                            className="animate-spin text-surface-400"
                          />
                        ) : isAdded ? (
                          <Check size={16} className="text-green-500" />
                        ) : (
                          <Plus
                            size={16}
                            className="text-surface-300 dark:text-surface-500 group-hover:text-surface-600 dark:group-hover:text-surface-300"
                          />
                        )}
                      </button>
                    );
                  })}
                </div>
              )}

              <button
                onClick={onClose}
                className="w-full mt-4 py-3 bg-surface-900 dark:bg-white text-white dark:text-surface-900 font-semibold rounded-xl hover:bg-surface-800 dark:hover:bg-surface-100 transition-colors"
              >
                {t("addToCollection.done")}
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
