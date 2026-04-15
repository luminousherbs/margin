import { atom } from "nanostores";
import { getPreferences, updatePreferences } from "../api/client";
import type {
  LabelerSubscription,
  LabelPreference,
  LabelVisibility,
} from "../types";

export interface Preferences {
  externalLinkSkippedHostnames: string[];
  subscribedLabelers: LabelerSubscription[];
  labelPreferences: LabelPreference[];
  disableExternalLinkWarning: boolean;
  enableCommunityBookmarks: boolean;
}

export const $preferences = atom<Preferences>({
  externalLinkSkippedHostnames: [],
  subscribedLabelers: [],
  labelPreferences: [],
  disableExternalLinkWarning: false,
  enableCommunityBookmarks: true,
});

export async function loadPreferences() {
  const prefs = await getPreferences();
  $preferences.set({
    externalLinkSkippedHostnames: prefs.externalLinkSkippedHostnames || [],
    subscribedLabelers: prefs.subscribedLabelers || [],
    labelPreferences: prefs.labelPreferences || [],
    disableExternalLinkWarning: !!prefs.disableExternalLinkWarning,
    enableCommunityBookmarks: !!prefs.enableCommunityBookmarks,
  });
}

export async function addSkippedHostname(hostname: string) {
  const current = $preferences.get();
  if (current.externalLinkSkippedHostnames.includes(hostname)) return;

  const updated = {
    ...current,
    externalLinkSkippedHostnames: [
      ...current.externalLinkSkippedHostnames,
      hostname,
    ],
  };
  $preferences.set(updated);
  await updatePreferences(updated);
}

export async function addLabeler(did: string) {
  const current = $preferences.get();
  if (current.subscribedLabelers.some((l) => l.did === did)) return;

  const updated = {
    ...current,
    subscribedLabelers: [...current.subscribedLabelers, { did }],
  };
  $preferences.set(updated);
  await updatePreferences(updated);
}

export async function removeLabeler(did: string) {
  const current = $preferences.get();
  const updated = {
    ...current,
    subscribedLabelers: current.subscribedLabelers.filter((l) => l.did !== did),
  };
  $preferences.set(updated);
  await updatePreferences(updated);
}

export async function setLabelVisibility(
  labelerDid: string,
  label: string,
  visibility: LabelVisibility,
) {
  const current = $preferences.get();
  const filtered = current.labelPreferences.filter(
    (p) => !(p.labelerDid === labelerDid && p.label === label),
  );
  const newPrefs =
    visibility === "warn"
      ? filtered
      : [...filtered, { labelerDid, label, visibility }];
  const updated = { ...current, labelPreferences: newPrefs };
  $preferences.set(updated);
  await updatePreferences(updated);
}

export function getLabelVisibility(
  labelerDid: string,
  label: string,
): LabelVisibility {
  const prefs = $preferences.get();
  const pref = prefs.labelPreferences.find(
    (p) => p.labelerDid === labelerDid && p.label === label,
  );
  return pref?.visibility || "warn";
}

export async function setDisableExternalLinkWarning(disabled: boolean) {
  const current = $preferences.get();
  if (current.disableExternalLinkWarning === disabled) return;

  const updated = {
    ...current,
    disableExternalLinkWarning: disabled,
  };
  $preferences.set(updated);
  await updatePreferences(updated);
}

export async function setEnableCommunityBookmarks(enabled: boolean) {
  const current = $preferences.get();
  if (current.enableCommunityBookmarks === enabled) return;

  const updated = {
    ...current,
    enableCommunityBookmarks: enabled,
  };
  $preferences.set(updated);
  await updatePreferences(updated);
}
