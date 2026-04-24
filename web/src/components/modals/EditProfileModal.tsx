import React, { useState, useRef } from "react";
import { updateProfile, uploadAvatar, getAvatarUrl } from "../../api/client";
import type { UserProfile } from "../../types";
import { Loader2, X, Plus, User as UserIcon } from "lucide-react";
import { useTranslation } from "react-i18next";

interface EditProfileModalProps {
  profile: UserProfile;
  onClose: () => void;
  onUpdate: (updatedProfile: UserProfile) => void;
}

export default function EditProfileModal({
  profile,
  onClose,
  onUpdate,
}: EditProfileModalProps) {
  const { t } = useTranslation();
  const [displayName, setDisplayName] = useState(profile.displayName || "");
  const [description, setDescription] = useState(profile.description || "");
  const [website, setWebsite] = useState(profile.website || "");
  const [links, setLinks] = useState<string[]>(profile.links || []);
  const [newLink, setNewLink] = useState("");

  const [avatarBlob, setAvatarBlob] = useState<Blob | string | null>(null);
  const [avatarPreview, setAvatarPreview] = useState<string | null>(null);
  const [uploading, setUploading] = useState(false);

  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleAvatarChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (!["image/jpeg", "image/png"].includes(file.type)) {
      setError(t("editProfile.avatarTypeError"));
      return;
    }

    if (file.size > 1024 * 1024 * 2) {
      setError(t("editProfile.avatarSizeError"));
      return;
    }

    setAvatarPreview(URL.createObjectURL(file));
    setAvatarBlob(file);

    setUploading(true);
    try {
      const result = await uploadAvatar(file);
      setAvatarBlob(result.blob);
      setAvatarBlob(result.blob);
    } catch (err) {
      setError(
        t("editProfile.avatarUploadError", {
          message: err instanceof Error ? err.message : "Unknown error",
        }),
      );
      setAvatarPreview(null);
    } finally {
      setUploading(false);
    }
  };

  const handleAddLink = () => {
    if (!newLink) return;
    if (!links.includes(newLink)) {
      setLinks([...links, newLink]);
      setNewLink("");
    }
  };

  const handleRemoveLink = (index: number) => {
    setLinks(links.filter((_, i) => i !== index));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setError(null);

    try {
      await updateProfile({
        displayName,
        description,
        website,
        links,
        avatar: avatarBlob,
      });
      onUpdate({
        ...profile,
        displayName,
        description,
        website,
        links,
        avatar: avatarPreview || profile.avatar,
      });
      onClose();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
    } finally {
      setSaving(false);
    }
  };

  const currentAvatar =
    avatarPreview || getAvatarUrl(profile.did, profile.avatar);

  return (
    <div
      className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4"
      onClick={onClose}
    >
      <div
        className="bg-white dark:bg-surface-900 rounded-xl w-full max-w-md overflow-hidden shadow-2xl ring-1 ring-black/5 dark:ring-white/10"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between p-4 border-b border-surface-100 dark:border-surface-800">
          <h2 className="text-lg font-bold text-surface-900 dark:text-white">
            {t("editProfile.title")}
          </h2>
          <button
            onClick={onClose}
            className="p-1.5 hover:bg-surface-100 dark:hover:bg-surface-800 rounded-lg transition-colors text-surface-500 dark:text-surface-400"
          >
            <X size={18} />
          </button>
        </div>

        <form
          onSubmit={handleSubmit}
          className="p-5 overflow-y-auto max-h-[80vh]"
        >
          {error && (
            <div className="mb-4 p-3 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 rounded-lg text-sm border border-red-100 dark:border-red-800">
              {error}
            </div>
          )}

          <div className="mb-5">
            <label className="block text-sm font-medium text-surface-700 dark:text-surface-300 mb-2">
              {t("editProfile.avatarLabel")}
            </label>
            <div className="flex items-center gap-3">
              <div
                className="relative w-16 h-16 rounded-full bg-surface-100 dark:bg-surface-800 overflow-hidden cursor-pointer group border border-surface-200 dark:border-surface-700"
                onClick={() => fileInputRef.current?.click()}
              >
                {currentAvatar ? (
                  <img
                    src={currentAvatar}
                    alt=""
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center text-surface-400 dark:text-surface-500">
                    <UserIcon size={24} />
                  </div>
                )}
                <div className="absolute inset-0 bg-black/40 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                  <span className="text-white text-xs font-medium">Edit</span>
                </div>
              </div>
              <input
                ref={fileInputRef}
                type="file"
                accept="image/jpeg,image/png"
                onChange={handleAvatarChange}
                className="hidden"
              />
              <button
                type="button"
                onClick={() => fileInputRef.current?.click()}
                className="px-3 py-1.5 rounded-lg bg-surface-100 dark:bg-surface-800 hover:bg-surface-200 dark:hover:bg-surface-700 text-surface-900 dark:text-white font-medium text-sm transition-colors"
                disabled={uploading}
              >
                {uploading
                  ? t("editProfile.uploading")
                  : t("editProfile.uploadButton")}
              </button>
            </div>
          </div>

          <div className="mb-4">
            <label className="block text-sm font-medium text-surface-700 dark:text-surface-300 mb-1">
              {t("editProfile.displayNameLabel")}
            </label>
            <input
              type="text"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              className="w-full px-3 py-2 rounded-lg bg-white dark:bg-surface-800 border border-surface-200 dark:border-surface-700 text-surface-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500/20 focus:border-primary-500 dark:focus:border-primary-400"
              maxLength={64}
            />
          </div>

          <div className="mb-4">
            <label className="block text-sm font-medium text-surface-700 dark:text-surface-300 mb-1">
              {t("editProfile.bioLabel")}
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full px-3 py-2 rounded-lg bg-white dark:bg-surface-800 border border-surface-200 dark:border-surface-700 text-surface-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500/20 focus:border-primary-500 dark:focus:border-primary-400 min-h-[80px] resize-none"
              maxLength={300}
            />
          </div>

          <div className="mb-4">
            <label className="block text-sm font-medium text-surface-700 dark:text-surface-300 mb-1">
              {t("editProfile.websiteLabel")}
            </label>
            <input
              type="url"
              value={website}
              onChange={(e) => setWebsite(e.target.value)}
              placeholder="https://example.com"
              className="w-full px-3 py-2 rounded-lg bg-white dark:bg-surface-800 border border-surface-200 dark:border-surface-700 text-surface-900 dark:text-white placeholder:text-surface-400 dark:placeholder:text-surface-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 focus:border-primary-500 dark:focus:border-primary-400 text-sm"
            />
          </div>

          <div className="mb-5">
            <label className="block text-sm font-medium text-surface-700 dark:text-surface-300 mb-1">
              {t("editProfile.linksLabel")}
            </label>
            <div className="space-y-2">
              {links.map((link, i) => (
                <div key={i} className="flex items-center gap-2">
                  <input
                    type="text"
                    value={link}
                    readOnly
                    className="flex-1 px-3 py-2 rounded-lg bg-surface-50 dark:bg-surface-800 border border-surface-200 dark:border-surface-700 text-sm text-surface-600 dark:text-surface-300"
                  />
                  <button
                    type="button"
                    onClick={() => handleRemoveLink(i)}
                    className="p-2 text-surface-400 dark:text-surface-500 hover:text-red-500 dark:hover:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg"
                  >
                    <X size={14} />
                  </button>
                </div>
              ))}
              <div className="flex items-center gap-2">
                <input
                  type="url"
                  value={newLink}
                  onChange={(e) => setNewLink(e.target.value)}
                  placeholder={t("editProfile.addLinkPlaceholder")}
                  className="flex-1 px-3 py-2 rounded-lg bg-white dark:bg-surface-800 border border-surface-200 dark:border-surface-700 text-surface-900 dark:text-white placeholder:text-surface-400 dark:placeholder:text-surface-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 focus:border-primary-500 dark:focus:border-primary-400 text-sm"
                  onKeyDown={(e) =>
                    e.key === "Enter" && (e.preventDefault(), handleAddLink())
                  }
                />
                <button
                  type="button"
                  onClick={handleAddLink}
                  className="p-2 bg-surface-900 dark:bg-surface-700 text-white rounded-lg hover:bg-surface-800 dark:hover:bg-surface-600"
                >
                  <Plus size={18} />
                </button>
              </div>
            </div>
          </div>

          <div className="flex items-center justify-end gap-2 pt-4 border-t border-surface-100 dark:border-surface-800">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 rounded-lg text-surface-600 dark:text-surface-300 font-medium hover:bg-surface-100 dark:hover:bg-surface-800 transition-colors"
              disabled={saving}
            >
              {t("editProfile.cancel")}
            </button>
            <button
              type="submit"
              className="px-4 py-2 rounded-lg bg-primary-600 text-white font-medium hover:bg-primary-700 transition-colors flex items-center gap-2"
              disabled={saving}
            >
              {saving && <Loader2 size={14} className="animate-spin" />}
              {saving ? t("editProfile.saving") : t("editProfile.save")}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
