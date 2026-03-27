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
  fetch("/auth/logout", { method: "POST" }).then(() => {
    window.location.href = "/";
  });
}
