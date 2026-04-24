import React, { useEffect, useState } from "react";
import { useStore } from "@nanostores/react";
import { useTranslation } from "react-i18next";
import { $user } from "../../store/auth";
import {
  checkAdminAccess,
  getAdminReports,
  adminTakeAction,
  adminCreateLabel,
  adminDeleteLabel,
  adminGetLabels,
} from "../../api/client";
import type { ModerationReport, HydratedLabel } from "../../types";
import {
  Shield,
  CheckCircle,
  XCircle,
  AlertTriangle,
  Eye,
  ChevronDown,
  ChevronUp,
  Tag,
  FileText,
  Plus,
  Trash2,
  EyeOff,
} from "lucide-react";
import { Avatar, EmptyState, Skeleton, Button } from "../../components/ui";

const STATUS_COLORS: Record<string, string> = {
  pending:
    "bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-300",
  resolved:
    "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300",
  dismissed:
    "bg-surface-100 text-surface-600 dark:bg-surface-800 dark:text-surface-400",
  escalated: "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300",
  acknowledged:
    "bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300",
};

const REASON_LABEL_KEYS: Record<string, string> = {
  spam: "adminModeration.reasons.spam",
  violation: "adminModeration.reasons.violation",
  misleading: "adminModeration.reasons.misleading",
  sexual: "adminModeration.reasons.sexual",
  rude: "adminModeration.reasons.rude",
  other: "adminModeration.reasons.other",
};

const LABEL_VALS = [
  "sexual",
  "nudity",
  "violence",
  "gore",
  "spam",
  "misleading",
];

type Tab = "reports" | "labels" | "actions";

export default function AdminModeration() {
  const { t } = useTranslation();
  const user = useStore($user);
  const [isAdmin, setIsAdmin] = useState(false);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<Tab>("reports");

  const [reports, setReports] = useState<ModerationReport[]>([]);
  const [pendingCount, setPendingCount] = useState(0);
  const [totalCount, setTotalCount] = useState(0);
  const [statusFilter, setStatusFilter] = useState<string>("pending");
  const [expandedReport, setExpandedReport] = useState<number | null>(null);
  const [actionLoading, setActionLoading] = useState<number | null>(null);

  const [labels, setLabels] = useState<HydratedLabel[]>([]);

  const [labelSrc, setLabelSrc] = useState("");
  const [labelUri, setLabelUri] = useState("");
  const [labelVal, setLabelVal] = useState("");
  const [labelSubmitting, setLabelSubmitting] = useState(false);
  const [labelSuccess, setLabelSuccess] = useState(false);

  const loadReports = async (status: string) => {
    const data = await getAdminReports(status || undefined);
    setReports(data.items);
    setPendingCount(data.pendingCount);
    setTotalCount(data.totalItems);
  };

  const loadLabels = async () => {
    const data = await adminGetLabels();
    setLabels(data.items || []);
  };

  useEffect(() => {
    const init = async () => {
      const admin = await checkAdminAccess();
      setIsAdmin(admin);
      if (admin) await loadReports("pending");
      setLoading(false);
    };
    init();
  }, []);

  const handleTabChange = async (tab: Tab) => {
    setActiveTab(tab);
    if (tab === "labels") await loadLabels();
  };

  const handleFilterChange = async (status: string) => {
    setStatusFilter(status);
    await loadReports(status);
  };

  const handleAction = async (reportId: number, action: string) => {
    setActionLoading(reportId);
    const success = await adminTakeAction({ reportId, action });
    if (success) {
      await loadReports(statusFilter);
      setExpandedReport(null);
    }
    setActionLoading(null);
  };

  const handleCreateLabel = async () => {
    if (!labelVal || (!labelSrc && !labelUri)) return;
    setLabelSubmitting(true);
    const success = await adminCreateLabel({
      src: labelSrc || labelUri,
      uri: labelUri || undefined,
      val: labelVal,
    });
    if (success) {
      setLabelSrc("");
      setLabelUri("");
      setLabelVal("");
      setLabelSuccess(true);
      setTimeout(() => setLabelSuccess(false), 2000);
      if (activeTab === "labels") await loadLabels();
    }
    setLabelSubmitting(false);
  };

  const handleDeleteLabel = async (id: number) => {
    if (!window.confirm(t("adminModeration.labels.removeConfirm"))) return;
    const success = await adminDeleteLabel(id);
    if (success) setLabels((prev) => prev.filter((l) => l.id !== id));
  };

  if (loading) {
    return (
      <div className="max-w-3xl mx-auto animate-slide-up">
        <Skeleton className="h-8 w-48 mb-6" />
        <div className="space-y-3">
          <Skeleton className="h-24 rounded-xl" />
          <Skeleton className="h-24 rounded-xl" />
          <Skeleton className="h-24 rounded-xl" />
        </div>
      </div>
    );
  }

  if (!user || !isAdmin) {
    return (
      <EmptyState
        icon={<Shield size={40} />}
        title={t("adminModeration.accessDenied")}
        message={t("adminModeration.accessDeniedMessage")}
      />
    );
  }

  return (
    <div className="max-w-3xl mx-auto animate-slide-up">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-display font-bold text-surface-900 dark:text-white flex items-center gap-2.5">
            <Shield
              size={24}
              className="text-primary-600 dark:text-primary-400"
            />
            {t("adminModeration.title")}
          </h1>
          <p className="text-sm text-surface-500 dark:text-surface-400 mt-1">
            {t("adminModeration.stats", {
              pending: pendingCount,
              total: totalCount,
            })}
          </p>
        </div>
      </div>

      <div className="flex gap-1 mb-5 border-b border-surface-200 dark:border-surface-700">
        {[
          {
            id: "reports" as Tab,
            label: t("adminModeration.tabs.reports"),
            icon: <FileText size={15} />,
          },
          {
            id: "actions" as Tab,
            label: t("adminModeration.tabs.actions"),
            icon: <EyeOff size={15} />,
          },
          {
            id: "labels" as Tab,
            label: t("adminModeration.tabs.labels"),
            icon: <Tag size={15} />,
          },
        ].map((tab) => (
          <button
            key={tab.id}
            onClick={() => handleTabChange(tab.id)}
            className={`flex items-center gap-1.5 px-4 py-2.5 text-sm font-medium border-b-2 -mb-px transition-colors ${
              activeTab === tab.id
                ? "border-primary-600 text-primary-600 dark:border-primary-400 dark:text-primary-400"
                : "border-transparent text-surface-500 dark:text-surface-400 hover:text-surface-700 dark:hover:text-surface-300"
            }`}
          >
            {tab.icon}
            {tab.label}
          </button>
        ))}
      </div>

      {activeTab === "reports" && (
        <>
          <div className="flex gap-2 mb-5">
            {["pending", "resolved", "dismissed", "escalated", ""].map(
              (status) => (
                <button
                  key={status || "all"}
                  onClick={() => handleFilterChange(status)}
                  className={`px-3.5 py-1.5 text-sm font-medium rounded-lg transition-colors ${
                    statusFilter === status
                      ? "bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300"
                      : "text-surface-500 dark:text-surface-400 hover:bg-surface-100 dark:hover:bg-surface-800"
                  }`}
                >
                  {status
                    ? t(`adminModeration.filters.${status}`, {
                        defaultValue:
                          status.charAt(0).toUpperCase() + status.slice(1),
                      })
                    : t("adminModeration.filters.all")}
                </button>
              ),
            )}
          </div>

          {reports.length === 0 ? (
            <EmptyState
              icon={<CheckCircle size={40} />}
              title={t("adminModeration.reports.empty")}
              message={
                statusFilter === "pending"
                  ? t("adminModeration.reports.emptyPending")
                  : t("adminModeration.reports.emptyFiltered", {
                      status: statusFilter || "",
                    })
              }
            />
          ) : (
            <div className="space-y-3">
              {reports.map((report) => (
                <div
                  key={report.id}
                  className="card overflow-hidden transition-all"
                >
                  <button
                    onClick={() =>
                      setExpandedReport(
                        expandedReport === report.id ? null : report.id,
                      )
                    }
                    className="w-full p-4 flex items-center gap-4 text-left hover:bg-surface-50 dark:hover:bg-surface-800/50 transition-colors"
                  >
                    <Avatar
                      did={report.subject.did}
                      avatar={report.subject.avatar}
                      size="sm"
                    />
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-0.5">
                        <span className="font-medium text-surface-900 dark:text-white text-sm truncate">
                          {report.subject.displayName ||
                            report.subject.handle ||
                            report.subject.did}
                        </span>
                        <span
                          className={`text-xs px-2 py-0.5 rounded-full font-medium ${STATUS_COLORS[report.status] || STATUS_COLORS.pending}`}
                        >
                          {report.status}
                        </span>
                      </div>
                      <p className="text-xs text-surface-500 dark:text-surface-400">
                        {REASON_LABEL_KEYS[report.reasonType]
                          ? t(REASON_LABEL_KEYS[report.reasonType])
                          : report.reasonType}{" "}
                        · reported by @
                        {report.reporter.handle || report.reporter.did} ·{" "}
                        {new Date(report.createdAt).toLocaleDateString()}
                      </p>
                    </div>
                    {expandedReport === report.id ? (
                      <ChevronUp size={16} className="text-surface-400" />
                    ) : (
                      <ChevronDown size={16} className="text-surface-400" />
                    )}
                  </button>

                  {expandedReport === report.id && (
                    <div className="px-4 pb-4 border-t border-surface-100 dark:border-surface-800 pt-3 space-y-3">
                      <div className="grid grid-cols-2 gap-3 text-sm">
                        <div>
                          <span className="text-surface-400 dark:text-surface-500 text-xs uppercase tracking-wider">
                            {t("adminModeration.reports.reportedUser")}
                          </span>
                          <a
                            href={`/profile/${report.subject.did}`}
                            className="block mt-1 text-primary-600 dark:text-primary-400 hover:underline font-medium"
                          >
                            @{report.subject.handle || report.subject.did}
                          </a>
                        </div>
                        <div>
                          <span className="text-surface-400 dark:text-surface-500 text-xs uppercase tracking-wider">
                            {t("adminModeration.reports.reporter")}
                          </span>
                          <a
                            href={`/profile/${report.reporter.did}`}
                            className="block mt-1 text-primary-600 dark:text-primary-400 hover:underline font-medium"
                          >
                            @{report.reporter.handle || report.reporter.did}
                          </a>
                        </div>
                      </div>

                      {report.reasonText && (
                        <div>
                          <span className="text-surface-400 dark:text-surface-500 text-xs uppercase tracking-wider">
                            {t("adminModeration.reports.details")}
                          </span>
                          <p className="text-sm text-surface-700 dark:text-surface-300 mt-1">
                            {report.reasonText}
                          </p>
                        </div>
                      )}

                      {report.subjectUri && (
                        <div>
                          <span className="text-surface-400 dark:text-surface-500 text-xs uppercase tracking-wider">
                            {t("adminModeration.reports.contentUri")}
                          </span>
                          <p className="text-xs text-surface-500 font-mono mt-1 break-all">
                            {report.subjectUri}
                          </p>
                        </div>
                      )}

                      {report.status === "pending" && (
                        <div className="flex items-center gap-2 pt-2">
                          <Button
                            size="sm"
                            variant="secondary"
                            onClick={() =>
                              handleAction(report.id, "acknowledge")
                            }
                            loading={actionLoading === report.id}
                            icon={<Eye size={14} />}
                          >
                            {t("adminModeration.reports.acknowledge")}
                          </Button>
                          <Button
                            size="sm"
                            variant="secondary"
                            onClick={() => handleAction(report.id, "dismiss")}
                            loading={actionLoading === report.id}
                            icon={<XCircle size={14} />}
                          >
                            {t("adminModeration.reports.dismiss")}
                          </Button>
                          <Button
                            size="sm"
                            onClick={() => handleAction(report.id, "takedown")}
                            loading={actionLoading === report.id}
                            icon={<AlertTriangle size={14} />}
                            className="!bg-red-600 hover:!bg-red-700 !text-white"
                          >
                            {t("adminModeration.reports.takedown")}
                          </Button>
                        </div>
                      )}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </>
      )}

      {activeTab === "actions" && (
        <div className="space-y-6">
          <div className="card p-5">
            <h3 className="text-base font-semibold text-surface-900 dark:text-white mb-1 flex items-center gap-2">
              <Tag
                size={16}
                className="text-primary-600 dark:text-primary-400"
              />
              {t("adminModeration.actions.applyWarning")}
            </h3>
            <p className="text-sm text-surface-500 dark:text-surface-400 mb-4">
              {t("adminModeration.actions.applyWarningDesc")}
            </p>

            <div className="space-y-3">
              <div>
                <label className="block text-xs font-medium text-surface-600 dark:text-surface-400 mb-1.5">
                  {t("adminModeration.actions.accountDid")}
                </label>
                <input
                  type="text"
                  value={labelSrc}
                  onChange={(e) => setLabelSrc(e.target.value)}
                  placeholder="did:plc:..."
                  className="w-full px-3 py-2 text-sm rounded-lg border border-surface-200 dark:border-surface-700 bg-white dark:bg-surface-800 text-surface-900 dark:text-white placeholder:text-surface-400 focus:outline-none focus:ring-2 focus:ring-primary-500/20 focus:border-primary-500"
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-surface-600 dark:text-surface-400 mb-1.5">
                  {t("adminModeration.actions.contentUri")}{" "}
                  <span className="text-surface-400">
                    ({t("adminModeration.actions.contentUriOptional")})
                  </span>
                </label>
                <input
                  type="text"
                  value={labelUri}
                  onChange={(e) => setLabelUri(e.target.value)}
                  placeholder="at://did:plc:.../at.margin.annotation/..."
                  className="w-full px-3 py-2 text-sm rounded-lg border border-surface-200 dark:border-surface-700 bg-white dark:bg-surface-800 text-surface-900 dark:text-white placeholder:text-surface-400 focus:outline-none focus:ring-2 focus:ring-primary-500/20 focus:border-primary-500"
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-surface-600 dark:text-surface-400 mb-1.5">
                  {t("adminModeration.actions.labelType")}
                </label>
                <div className="grid grid-cols-3 gap-2">
                  {LABEL_VALS.map((val) => (
                    <button
                      key={val}
                      onClick={() => setLabelVal(val)}
                      className={`px-3 py-2 text-sm font-medium rounded-lg border transition-all ${
                        labelVal === val
                          ? "border-primary-500 bg-primary-50 dark:bg-primary-900/20 text-primary-700 dark:text-primary-300 ring-2 ring-primary-500/20"
                          : "border-surface-200 dark:border-surface-700 text-surface-600 dark:text-surface-400 hover:bg-surface-50 dark:hover:bg-surface-800"
                      }`}
                    >
                      {t(`card.labelDescriptions.${val}`)}
                    </button>
                  ))}
                </div>
              </div>

              <div className="flex items-center gap-3 pt-1">
                <Button
                  onClick={handleCreateLabel}
                  loading={labelSubmitting}
                  disabled={!labelVal || (!labelSrc && !labelUri)}
                  icon={<Plus size={14} />}
                  size="sm"
                >
                  {t("adminModeration.actions.applyLabel")}
                </Button>
                {labelSuccess && (
                  <span className="text-sm text-green-600 dark:text-green-400 flex items-center gap-1.5">
                    <CheckCircle size={14} />{" "}
                    {t("adminModeration.actions.labelApplied")}
                  </span>
                )}
              </div>
            </div>
          </div>
        </div>
      )}

      {activeTab === "labels" && (
        <div>
          {labels.length === 0 ? (
            <EmptyState
              icon={<Tag size={40} />}
              title={t("adminModeration.labels.empty")}
              message={t("adminModeration.labels.emptyMessage")}
            />
          ) : (
            <div className="space-y-2">
              {labels.map((label) => (
                <div
                  key={label.id}
                  className="card p-4 flex items-center gap-4"
                >
                  {label.subject && (
                    <Avatar
                      did={label.subject.did}
                      avatar={label.subject.avatar}
                      size="sm"
                    />
                  )}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-0.5">
                      <span
                        className={`text-xs px-2 py-0.5 rounded-full font-medium ${
                          label.val === "sexual" || label.val === "nudity"
                            ? "bg-pink-100 text-pink-800 dark:bg-pink-900/30 dark:text-pink-300"
                            : label.val === "violence" || label.val === "gore"
                              ? "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300"
                              : "bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-300"
                        }`}
                      >
                        {label.val}
                      </span>
                      {label.subject && (
                        <a
                          href={`/profile/${label.subject.did}`}
                          className="text-sm font-medium text-surface-900 dark:text-white hover:text-primary-600 dark:hover:text-primary-400 truncate"
                        >
                          @{label.subject.handle || label.subject.did}
                        </a>
                      )}
                    </div>
                    <p className="text-xs text-surface-500 dark:text-surface-400 truncate">
                      {label.uri !== label.src
                        ? label.uri
                        : t("adminModeration.labels.accountLevel")}{" "}
                      · {new Date(label.createdAt).toLocaleDateString()} · by @
                      {label.createdBy.handle || label.createdBy.did}
                    </p>
                  </div>
                  <button
                    onClick={() => handleDeleteLabel(label.id)}
                    className="p-2 rounded-lg text-surface-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
                    title={t("adminModeration.labels.removeTitle")}
                  >
                    <Trash2 size={14} />
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
