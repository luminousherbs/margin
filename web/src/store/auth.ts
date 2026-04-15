import { atom } from "nanostores";
import { loadPreferences } from "./preferences";
import type { UserProfile } from "../types";
import { analytics } from "../lib/analytics";

export const $user = atom<UserProfile | null>(null);

$user.subscribe((user) => {
  if (user) {
    loadPreferences();
    analytics.identify(user.did, {
      handle: user.handle,
      displayName: user.displayName,
    });
  }
});

export function logout() {
  analytics.capture("user_logged_out");
  analytics.reset();
  $user.set(null);
  fetch("/auth/logout", { method: "POST" }).finally(() => {
    window.location.href = "/";
  });
}
