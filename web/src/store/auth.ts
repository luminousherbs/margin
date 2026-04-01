import { atom } from "nanostores";
import { loadPreferences } from "./preferences";
import type { UserProfile } from "../types";

export const $user = atom<UserProfile | null>(null);

$user.subscribe((user) => {
  if (user) {
    loadPreferences();
  }
});

export function logout() {
  $user.set(null);
  fetch("/auth/logout", { method: "POST" }).finally(() => {
    window.location.href = "/";
  });
}
