import React from "react";
import { useStore } from "@nanostores/react";
import { useTranslation } from "react-i18next";
import "../i18n";

import { $theme, cycleTheme } from "../store/theme";
import { $user } from "../store/auth";
import {
  MessageSquareText,
  Highlighter,
  Bookmark,
  FolderOpen,
  Keyboard,
  PanelRight,
  MousePointerClick,
  Shield,
  Users,
  Chrome,
  ArrowRight,
  Github,
  ExternalLink,
  Hash,
  Coffee,
  Heart,
  Eye,
  Sun,
  Moon,
  Monitor,
} from "lucide-react";
import { AppleIcon, TangledIcon } from "../components/common/Icons";
import { FaFirefox, FaEdge } from "react-icons/fa";

function FeatureCard({
  icon: Icon,
  title,
  description,
  accent = false,
}: {
  icon: React.ElementType;
  title: string;
  description: string;
  accent?: boolean;
}) {
  return (
    <div
      className={`group p-6 rounded-2xl border transition-all duration-200 hover:-translate-y-0.5 hover:shadow-sm ${
        accent
          ? "bg-primary-50 dark:bg-primary-950/30 border-primary-200/50 dark:border-primary-800/40"
          : "bg-white dark:bg-surface-800 border-surface-200/60 dark:border-surface-700/60"
      }`}
    >
      <div
        className={`w-11 h-11 rounded-xl flex items-center justify-center mb-4 transition-colors ${
          accent
            ? "bg-primary-600 text-white"
            : "bg-surface-100 dark:bg-surface-700/50 text-surface-500 dark:text-surface-400 group-hover:bg-primary-600 group-hover:text-white dark:group-hover:bg-primary-500"
        }`}
      >
        <Icon size={20} />
      </div>
      <h3 className="font-display font-semibold text-base mb-2 text-surface-900 dark:text-white">
        {title}
      </h3>
      <p className="text-sm text-surface-500 dark:text-surface-400 leading-relaxed">
        {description}
      </p>
    </div>
  );
}

function ExtensionFeature({
  icon: Icon,
  title,
  description,
}: {
  icon: React.ElementType;
  title: string;
  description: string;
}) {
  return (
    <div className="flex gap-4 items-start">
      <div className="w-9 h-9 rounded-lg bg-primary-100 dark:bg-primary-900/30 flex items-center justify-center flex-shrink-0 text-primary-600 dark:text-primary-400">
        <Icon size={18} />
      </div>
      <div>
        <h4 className="font-semibold text-sm text-surface-900 dark:text-white mb-1">
          {title}
        </h4>
        <p className="text-sm text-surface-500 dark:text-surface-400 leading-relaxed">
          {description}
        </p>
      </div>
    </div>
  );
}

export default function About() {
  const { t } = useTranslation();
  const theme = useStore($theme); // ensure theme is applied on this page
  const user = useStore($user);

  const [browser] = React.useState<
    "chrome" | "firefox" | "edge" | "safari" | "other"
  >(() => {
    if (typeof navigator === "undefined") return "other";
    const ua = navigator.userAgent;
    if (/Edg\//i.test(ua)) return "edge";
    if (/Firefox/i.test(ua)) return "firefox";
    if (/^((?!chrome|android).)*safari/i.test(ua)) return "safari";
    if (/Chrome/i.test(ua)) return "chrome";
    return "other";
  });

  const extensionLink =
    browser === "firefox"
      ? "https://addons.mozilla.org/en-US/firefox/addon/margin/"
      : browser === "edge"
        ? "https://microsoftedge.microsoft.com/addons/detail/margin/nfjnmllpdgcdnhmmggjihjbidmeadddn"
        : "https://chromewebstore.google.com/detail/margin/cgpmbiiagnehkikhcbnhiagfomajncpa";

  const ExtensionIcon =
    browser === "firefox" ? FaFirefox : browser === "edge" ? FaEdge : Chrome;
  const extensionLabel =
    browser === "firefox"
      ? "Firefox"
      : browser === "edge"
        ? "Edge"
        : browser === "safari"
          ? "Chrome"
          : "Chrome";

  const [isScrolled, setIsScrolled] = React.useState(false);

  React.useEffect(() => {
    const handleScroll = () => {
      setIsScrolled((prev) =>
        prev ? window.scrollY > 10 : window.scrollY > 50,
      );
    };
    window.addEventListener("scroll", handleScroll, { passive: true });
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  return (
    <div className="min-h-screen bg-surface-100 dark:bg-surface-900">
      <nav className="sticky top-0 z-50 pt-3 pb-1 px-4 sm:px-6 mx-auto max-w-5xl">
        <div
          className={`relative flex items-center justify-between rounded-2xl px-4 sm:px-5 transition-all duration-300 ease-out ${
            isScrolled ? "h-12" : "h-14"
          }`}
        >
          <div
            className={`absolute inset-0 rounded-2xl bg-white/75 dark:bg-surface-900/75 backdrop-blur-lg border border-surface-200/40 dark:border-surface-700/40 shadow-sm transition-opacity duration-300 ease-out ${
              isScrolled ? "opacity-100" : "opacity-0"
            }`}
          />
          <div className="relative flex items-center gap-6">
            <a
              href="/"
              className="flex items-center gap-2.5 hover:opacity-80 transition-opacity"
            >
              <img src="/logo.svg" alt="Margin" className="w-7 h-7" />
              <span className="font-display font-bold text-lg tracking-tight text-surface-900 dark:text-white">
                Margin
              </span>
            </a>
            <div className="hidden md:flex items-center gap-0.5">
              <a
                href="/home"
                className="text-[13px] font-medium text-surface-500 dark:text-surface-400 hover:text-surface-900 dark:hover:text-white transition-colors px-3 py-1.5 rounded-lg hover:bg-surface-100 dark:hover:bg-surface-800"
              >
                {t("nav.feed")}
              </a>
              <a
                href="/discover"
                className="text-[13px] font-medium text-surface-500 dark:text-surface-400 hover:text-surface-900 dark:hover:text-white transition-colors px-3 py-1.5 rounded-lg hover:bg-surface-100 dark:hover:bg-surface-800"
              >
                {t("nav.discover")}
              </a>
            </div>
          </div>
          <div className="relative flex items-center gap-2">
            {!user && (
              <a
                href="/login"
                className="text-[13px] font-medium text-surface-600 dark:text-surface-300 hover:text-surface-900 dark:hover:text-white transition-colors px-3 py-1.5 rounded-lg hover:bg-surface-100 dark:hover:bg-surface-800"
              >
                {t("nav.signIn")}
              </a>
            )}

            <a
              href={extensionLink}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1.5 text-[13px] font-semibold px-3.5 py-1.5 bg-surface-900 dark:bg-white text-white dark:text-surface-900 rounded-lg hover:bg-surface-800 dark:hover:bg-surface-100 transition-colors"
            >
              <ExtensionIcon size={14} />
              <span className="hidden sm:inline">
                {t("about.nav.getExtension")}
              </span>
              <span className="sm:hidden">{t("about.nav.install")}</span>
            </a>
          </div>
        </div>
      </nav>

      <section className="relative overflow-hidden">
        <div className="absolute top-0 inset-x-0 h-[500px] bg-gradient-to-b from-primary-50/50 via-transparent to-transparent dark:from-primary-900/10 dark:to-transparent -z-10 pointer-events-none" />

        <div className="max-w-4xl mx-auto px-6 pt-8 pb-20 md:pt-16 md:pb-28 text-center relative z-10">
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4 mb-10">
            <div className="group relative inline-flex items-center gap-2 px-1 py-1 rounded-full bg-surface-50/50 dark:bg-surface-800/30 border border-surface-200 dark:border-surface-700/50 hover:bg-surface-100/50 dark:hover:bg-surface-800/50 transition-colors cursor-pointer">
              <div className="flex items-center -space-x-2">
                <a
                  href="https://github.com/margin-at"
                  target="_blank"
                  rel="noreferrer"
                  className="relative z-10 flex items-center justify-center w-8 h-8 rounded-full bg-white dark:bg-surface-900 text-surface-600 hover:text-surface-900 dark:text-surface-400 dark:hover:text-white border-2 border-surface-50 dark:border-surface-800 shadow-sm transition-transform hover:z-20 hover:scale-110"
                  title="GitHub"
                >
                  <Github size={15} />
                </a>
                <a
                  href="https://tangled.org/margin.at/margin"
                  target="_blank"
                  rel="noreferrer"
                  className="relative z-0 flex items-center justify-center w-8 h-8 rounded-full bg-white dark:bg-surface-900 border-2 border-surface-50 dark:border-surface-800 shadow-sm transition-transform hover:z-20 hover:scale-110"
                  title="Tangled"
                >
                  <TangledIcon
                    size={16}
                    className="text-surface-600 hover:text-surface-900 dark:text-surface-400 dark:hover:text-white transition-colors"
                  />
                </a>
              </div>
              <span className="pr-4 pl-0.5 text-[13px] font-semibold text-surface-600 dark:text-surface-300">
                {t("about.hero.openSource")}{" "}
                <span className="text-primary-500 font-normal">✨</span>
              </span>
            </div>
          </div>

          <h1 className="font-display text-5xl md:text-6xl lg:text-7xl font-bold tracking-tight text-surface-900 dark:text-white leading-[1.05] mb-6">
            {t("about.hero.headline")} <br className="hidden sm:block" />
            <span className="text-primary-600 dark:text-primary-400">
              {t("about.hero.headlineAccent")}
            </span>
          </h1>

          <p className="text-lg md:text-xl text-surface-500 dark:text-surface-400 max-w-2xl mx-auto leading-relaxed mb-10">
            {t("about.hero.descriptionPre")}{" "}
            <a
              href="https://atproto.com"
              target="_blank"
              rel="noreferrer"
              className="text-surface-800 dark:text-surface-200 hover:text-primary-600 dark:hover:text-primary-400 border-b border-surface-300 dark:border-surface-600 hover:border-primary-400 transition-colors font-medium"
            >
              {t("about.hero.atProtocol")}
            </a>
            {t("about.hero.descriptionPost")}
          </p>

          <div className="flex flex-col sm:flex-row items-center justify-center gap-4 mt-4">
            <a
              href={user ? "/home" : "/login"}
              className="group inline-flex items-center justify-center gap-2 px-8 py-3.5 bg-surface-900 dark:bg-white text-white dark:text-surface-900 rounded-[14px] font-semibold hover:bg-surface-800 dark:hover:bg-surface-200 hover:scale-[1.02] shadow-sm transition-all duration-200 w-full sm:w-auto"
            >
              {user ? t("about.hero.openApp") : t("about.hero.getStarted")}
              <ArrowRight
                size={18}
                className="transition-transform group-hover:translate-x-1"
              />
            </a>
            <a
              href={extensionLink}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center justify-center gap-2 px-8 py-3.5 bg-surface-50 dark:bg-surface-800/50 text-surface-700 dark:text-surface-200 hover:text-surface-900 dark:hover:text-white rounded-[14px] font-semibold hover:bg-surface-100 dark:hover:bg-surface-800 hover:scale-[1.02] transition-all duration-200 w-full sm:w-auto"
            >
              <ExtensionIcon size={18} />
              {t("about.hero.installFor", { browser: extensionLabel })}
            </a>
          </div>
        </div>
      </section>

      <section className="max-w-5xl mx-auto px-6 py-20 md:py-24">
        <div className="text-center mb-12">
          <h2 className="font-display text-2xl md:text-3xl font-bold tracking-tight text-surface-900 dark:text-white mb-4">
            {t("about.features.title")}
          </h2>
          <p className="text-surface-500 dark:text-surface-400 max-w-xl mx-auto leading-relaxed">
            {t("about.features.subtitle")}
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
          <FeatureCard
            icon={MessageSquareText}
            title={t("about.features.annotations.title")}
            description={t("about.features.annotations.description")}
            accent
          />
          <FeatureCard
            icon={Highlighter}
            title={t("about.features.highlights.title")}
            description={t("about.features.highlights.description")}
          />
          <FeatureCard
            icon={Bookmark}
            title={t("about.features.bookmarks.title")}
            description={t("about.features.bookmarks.description")}
          />
          <FeatureCard
            icon={FolderOpen}
            title={t("about.features.collections.title")}
            description={t("about.features.collections.description")}
          />
          <FeatureCard
            icon={Users}
            title={t("about.features.socialDiscovery.title")}
            description={t("about.features.socialDiscovery.description")}
          />
          <FeatureCard
            icon={Hash}
            title={t("about.features.tagsSearch.title")}
            description={t("about.features.tagsSearch.description")}
          />
        </div>
      </section>

      <section className="bg-white/50 dark:bg-surface-800/50 border-y border-surface-200/60 dark:border-surface-800/60">
        <div className="max-w-5xl mx-auto px-6 py-20 md:py-24">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 lg:gap-16 items-center">
            <div>
              <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-surface-100 dark:bg-surface-800 text-surface-600 dark:text-surface-400 text-xs font-medium mb-5 border border-surface-200/60 dark:border-surface-700/60">
                <ExtensionIcon size={13} />
                {t("about.extension.badge")}
              </div>
              <h2 className="font-display text-2xl md:text-3xl font-bold tracking-tight text-surface-900 dark:text-white mb-4">
                {t("about.extension.title")}
                <br />
                {t("about.extension.titleLine2")}
              </h2>
              <p className="text-surface-500 dark:text-surface-400 leading-relaxed mb-8">
                {t("about.extension.description")}
              </p>

              <div className="space-y-5">
                <ExtensionFeature
                  icon={Eye}
                  title={t("about.extension.features.inlineOverlay.title")}
                  description={t(
                    "about.extension.features.inlineOverlay.description",
                  )}
                />
                <ExtensionFeature
                  icon={MousePointerClick}
                  title={t("about.extension.features.contextMenu.title")}
                  description={t(
                    "about.extension.features.contextMenu.description",
                  )}
                />
                <ExtensionFeature
                  icon={Keyboard}
                  title={t("about.extension.features.keyboard.title")}
                  description={t(
                    "about.extension.features.keyboard.description",
                  )}
                />
                <ExtensionFeature
                  icon={PanelRight}
                  title={t("about.extension.features.sidePanel.title")}
                  description={t(
                    "about.extension.features.sidePanel.description",
                  )}
                />
              </div>

              <div className="flex flex-col sm:flex-row gap-3 mt-8 flex-wrap">
                <a
                  href={extensionLink}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center justify-center gap-2 px-5 py-2.5 bg-surface-900 dark:bg-white text-white dark:text-surface-900 rounded-lg font-medium text-sm transition-all hover:opacity-90"
                >
                  <ExtensionIcon size={15} />
                  {t("about.hero.installFor", { browser: extensionLabel })}
                  <ExternalLink size={12} />
                </a>
                {browser !== "chrome" && (
                  <a
                    href="https://chromewebstore.google.com/detail/margin/cgpmbiiagnehkikhcbnhiagfomajncpa"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center justify-center gap-2 px-5 py-2.5 bg-surface-100 dark:bg-surface-800 text-surface-700 dark:text-surface-200 rounded-lg font-medium text-sm transition-all hover:bg-surface-200 dark:hover:bg-surface-700 border border-surface-200/80 dark:border-surface-700/80"
                  >
                    <Chrome size={15} />
                    Chrome
                    <ExternalLink size={12} />
                  </a>
                )}
                {browser !== "firefox" && (
                  <a
                    href="https://addons.mozilla.org/en-US/firefox/addon/margin/"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center justify-center gap-2 px-5 py-2.5 bg-surface-100 dark:bg-surface-800 text-surface-700 dark:text-surface-200 rounded-lg font-medium text-sm transition-all hover:bg-surface-200 dark:hover:bg-surface-700 border border-surface-200/80 dark:border-surface-700/80"
                  >
                    <FaFirefox size={15} />
                    Firefox
                    <ExternalLink size={12} />
                  </a>
                )}
                {browser !== "edge" && (
                  <a
                    href="https://microsoftedge.microsoft.com/addons/detail/margin/nfjnmllpdgcdnhmmggjihjbidmeadddn"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center justify-center gap-2 px-5 py-2.5 bg-surface-100 dark:bg-surface-800 text-surface-700 dark:text-surface-200 rounded-lg font-medium text-sm transition-all hover:bg-surface-200 dark:hover:bg-surface-700 border border-surface-200/80 dark:border-surface-700/80"
                  >
                    <FaEdge size={15} />
                    Edge
                    <ExternalLink size={12} />
                  </a>
                )}
                <a
                  href="https://www.icloud.com/shortcuts/1e33ebf52f55431fae1e187cfe9738c3"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center justify-center gap-2 px-5 py-2.5 bg-surface-100 dark:bg-surface-800 text-surface-700 dark:text-surface-200 rounded-lg font-medium text-sm transition-all hover:bg-surface-200 dark:hover:bg-surface-700 border border-surface-200/80 dark:border-surface-700/80"
                >
                  <AppleIcon size={15} />
                  {t("about.extension.iosShortcut")}
                  <ExternalLink size={12} />
                </a>
              </div>
            </div>

            <div className="relative hidden lg:block">
              <div className="relative rounded-2xl border border-surface-200/60 dark:border-surface-700/60 bg-white dark:bg-surface-800 p-6 shadow-xl">
                <div className="flex items-center gap-2 mb-4">
                  <div className="flex gap-1.5">
                    <div className="w-3 h-3 rounded-full bg-red-400/60" />
                    <div className="w-3 h-3 rounded-full bg-yellow-400/60" />
                    <div className="w-3 h-3 rounded-full bg-green-400/60" />
                  </div>
                  <div className="flex-1 mx-3 bg-surface-200 dark:bg-surface-700 rounded-md h-6 flex items-center px-3">
                    <span className="text-[10px] text-surface-400 truncate">
                      example.com/article/how-to-think-clearly
                    </span>
                  </div>
                </div>

                <div className="space-y-3 text-sm text-surface-600 dark:text-surface-300 leading-relaxed">
                  <div className="h-3 bg-surface-200 dark:bg-surface-700 rounded w-3/4" />
                  <div className="h-3 bg-surface-200 dark:bg-surface-700 rounded w-full" />
                  <div className="flex gap-0.5 flex-wrap items-center">
                    <div className="h-3 bg-surface-200 dark:bg-surface-700 rounded w-1/4" />
                    <span className="px-1 py-0.5 bg-yellow-200/70 dark:bg-yellow-500/30 rounded text-xs text-surface-700 dark:text-yellow-200 font-medium leading-none">
                      The point here is that Margin is indeed
                    </span>
                    <div className="h-3 bg-surface-200 dark:bg-surface-700 rounded w-1/5" />
                  </div>
                  <div className="h-3 bg-surface-200 dark:bg-surface-700 rounded w-5/6" />
                  <div className="flex gap-0.5 flex-wrap items-center">
                    <div className="h-3 bg-surface-200 dark:bg-surface-700 rounded w-2/5" />
                    <span className="px-1 py-0.5 bg-primary-200/70 dark:bg-primary-500/30 rounded text-xs text-primary-700 dark:text-primary-200 font-medium leading-none">
                      the best thing ever
                    </span>
                    <div className="h-3 bg-surface-200 dark:bg-surface-700 rounded w-1/4" />
                  </div>
                  <div className="h-3 bg-surface-200 dark:bg-surface-700 rounded w-2/3" />
                </div>

                <div className="absolute -right-4 top-1/3 w-56 bg-white dark:bg-surface-900 rounded-xl border border-surface-200/60 dark:border-surface-700/60 shadow-lg p-3.5">
                  <div className="flex items-center gap-2 mb-2">
                    <div className="w-6 h-6 rounded-full bg-gradient-to-br from-primary-400 to-primary-600 flex items-center justify-center text-white text-[10px] font-bold">
                      S
                    </div>
                    <span className="text-xs font-semibold text-surface-900 dark:text-white">
                      @scan.margin.cafe
                    </span>
                  </div>
                  <p className="text-xs text-surface-600 dark:text-surface-300 leading-relaxed">
                    I agree, Margin is just so good, like the other day I was
                    drinking some of that Margin for breakfast
                  </p>
                  <div className="flex items-center gap-3 mt-2.5 pt-2 border-t border-surface-100 dark:border-surface-700">
                    <span className="text-[10px] text-surface-400 flex items-center gap-1">
                      <Heart size={10} /> 3
                    </span>
                    <span className="text-[10px] text-surface-400 flex items-center gap-1">
                      <MessageSquareText size={10} /> 1
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="max-w-5xl mx-auto px-6 py-20 md:py-24">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-12 items-center">
          <div>
            <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-white dark:bg-surface-800 text-surface-600 dark:text-surface-400 text-xs font-medium mb-5 border border-surface-200/60 dark:border-surface-700/60">
              <Shield size={13} />
              {t("about.protocol.badge")}
            </div>
            <h2 className="font-display text-2xl md:text-3xl font-bold tracking-tight text-surface-900 dark:text-white mb-4">
              {t("about.protocol.title")}
            </h2>
            <p className="text-surface-500 dark:text-surface-400 leading-relaxed mb-6">
              {t("about.protocol.descriptionPre")}{" "}
              <a
                href="https://atproto.com"
                target="_blank"
                rel="noreferrer"
                className="text-primary-600 dark:text-primary-400 hover:underline font-medium"
              >
                {t("about.hero.atProtocol")}
              </a>
              {t("about.protocol.descriptionPost")}
            </p>
            <ul className="space-y-3 text-sm text-surface-600 dark:text-surface-300">
              {([0, 1, 2, 3] as const).map((i) => (
                <li key={i} className="flex items-start gap-3">
                  <div className="w-5 h-5 rounded-full bg-primary-100 dark:bg-primary-900/30 flex items-center justify-center flex-shrink-0 mt-0.5">
                    <div className="w-1.5 h-1.5 rounded-full bg-primary-600 dark:bg-primary-400" />
                  </div>
                  {t(`about.protocol.point${i}`)}
                </li>
              ))}
            </ul>
          </div>

          <div className="rounded-2xl bg-surface-900 dark:bg-surface-950 p-5 text-sm font-mono shadow-xl border border-surface-800 dark:border-surface-800">
            <div className="flex items-center gap-2 mb-4">
              <div className="text-xs text-surface-500">lexicon</div>
              <div className="text-xs text-primary-400 px-2 py-0.5 rounded bg-primary-400/10">
                at.margin.annotation
              </div>
            </div>
            <div className="space-y-1 text-[13px] leading-relaxed">
              <span className="text-surface-500">{"{"}</span>
              <div className="pl-4">
                <span className="text-green-400">"type"</span>
                <span className="text-surface-400">: </span>
                <span className="text-amber-400">"record"</span>
                <span className="text-surface-400">,</span>
              </div>
              <div className="pl-4">
                <span className="text-green-400">"record"</span>
                <span className="text-surface-400">: {"{"}</span>
              </div>
              <div className="pl-8">
                <span className="text-green-400">"body"</span>
                <span className="text-surface-400">: </span>
                <span className="text-amber-400">"Great insight..."</span>
                <span className="text-surface-400">,</span>
              </div>
              <div className="pl-8">
                <span className="text-green-400">"target"</span>
                <span className="text-surface-400">: {"{"}</span>
              </div>
              <div className="pl-12">
                <span className="text-green-400">"source"</span>
                <span className="text-surface-400">: </span>
                <span className="text-sky-400">"https://..."</span>
                <span className="text-surface-400">,</span>
              </div>
              <div className="pl-12">
                <span className="text-green-400">"selector"</span>
                <span className="text-surface-400">: {"{"}</span>
              </div>
              <div className="pl-16">
                <span className="text-green-400">"exact"</span>
                <span className="text-surface-400">: </span>
                <span className="text-amber-400">"selected text"</span>
              </div>
              <div className="pl-12">
                <span className="text-surface-400">{"}"}</span>
              </div>
              <div className="pl-8">
                <span className="text-surface-400">{"}"}</span>
              </div>
              <div className="pl-4">
                <span className="text-surface-400">{"}"}</span>
              </div>
              <span className="text-surface-500">{"}"}</span>
            </div>
          </div>
        </div>
      </section>

      <section className="border-t border-surface-200/60 dark:border-surface-800/60">
        <div className="max-w-5xl mx-auto px-6 py-20 md:py-24 text-center">
          <h2 className="font-display text-2xl md:text-3xl font-bold tracking-tight text-surface-900 dark:text-white mb-4">
            {t("about.cta.title")}
          </h2>
          <p className="text-surface-500 dark:text-surface-400 max-w-lg mx-auto mb-8">
            {t("about.cta.description")}
          </p>
          <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
            <a
              href={user ? "/home" : "/login"}
              className="inline-flex items-center gap-2 px-7 py-3 bg-primary-600 hover:bg-primary-700 dark:bg-primary-500 dark:hover:bg-primary-400 text-white rounded-xl font-semibold transition-colors"
            >
              {user ? t("about.hero.openApp") : t("about.cta.signIn")}
              <ArrowRight size={16} />
            </a>
            <a
              href="https://github.com/margin-at"
              target="_blank"
              rel="noreferrer"
              className="inline-flex items-center gap-2 px-7 py-3 text-surface-600 dark:text-surface-300 hover:text-surface-900 dark:hover:text-white transition-colors font-medium"
            >
              <Github size={16} />
              {t("about.cta.viewGitHub")}
            </a>
            <a
              href="https://tangled.org/margin.at/margin"
              target="_blank"
              rel="noreferrer"
              className="inline-flex items-center gap-2 px-7 py-3 text-surface-600 dark:text-surface-300 hover:text-surface-900 dark:hover:text-white transition-colors font-medium"
            >
              <TangledIcon size={16} />
              {t("about.cta.viewTangled")}
            </a>
          </div>
          <div className="mt-10 flex items-center gap-5 flex-wrap justify-center">
            <a
              href="https://ko-fi.com/scan"
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-2 text-surface-500 dark:text-surface-400 hover:text-[#FF5E5B] dark:hover:text-[#FF5E5B] transition-colors font-medium"
            >
              <Coffee size={16} />
              Ko-fi
            </a>
          </div>
        </div>
      </section>

      <footer className="border-t border-surface-200/60 dark:border-surface-800/60">
        <div className="max-w-5xl mx-auto px-6 py-8">
          <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
            <div className="flex items-center gap-2.5">
              <img
                src="/logo.svg"
                alt="Margin"
                className="w-5 h-5 opacity-60"
              />
              <span className="text-sm text-surface-400 dark:text-surface-500">
                {t("sidebar.copyright")}
              </span>
            </div>
            <div className="flex items-center gap-5 text-sm text-surface-400 dark:text-surface-500">
              {user && (
                <a
                  href="/home"
                  className="hover:text-surface-600 dark:hover:text-surface-300 transition-colors"
                >
                  {t("nav.feed")}
                </a>
              )}
              <a
                href="/privacy"
                className="hover:text-surface-600 dark:hover:text-surface-300 transition-colors"
              >
                {t("about.footer.privacy")}
              </a>
              <a
                href="/terms"
                className="hover:text-surface-600 dark:hover:text-surface-300 transition-colors"
              >
                {t("about.footer.terms")}
              </a>
              <a
                href="https://github.com/margin-at"
                target="_blank"
                rel="noreferrer"
                className="hover:text-surface-600 dark:hover:text-surface-300 transition-colors"
              >
                GitHub
              </a>
              <a
                href="https://tangled.org/margin.at/margin"
                target="_blank"
                rel="noreferrer"
                className="hover:text-surface-600 dark:hover:text-surface-300 transition-colors"
              >
                Tangled
              </a>
              <a
                href="mailto:hello@margin.at"
                className="hover:text-surface-600 dark:hover:text-surface-300 transition-colors"
              >
                {t("about.footer.contact")}
              </a>
              <div className="w-px h-4 bg-surface-200 dark:bg-surface-700 ml-1" />
              <button
                onClick={cycleTheme}
                title={
                  theme === "light"
                    ? t("nav.themeLight")
                    : theme === "dark"
                      ? t("nav.themeDark")
                      : t("nav.themeSystem")
                }
                className="flex items-center gap-1.5 hover:text-surface-600 dark:hover:text-surface-300 transition-colors"
              >
                {theme === "light" ? (
                  <Sun size={15} />
                ) : theme === "dark" ? (
                  <Moon size={15} />
                ) : (
                  <Monitor size={15} />
                )}
                <span className="hidden sm:inline capitalize">{theme}</span>
              </button>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}
