import React, { useState } from "react";
import { Flag, X } from "lucide-react";
import { useTranslation } from "react-i18next";
import { reportUser } from "../../api/client";
import type { ReportReasonType } from "../../types";

interface ReportModalProps {
  isOpen: boolean;
  onClose: () => void;
  subjectDid: string;
  subjectUri?: string;
  subjectHandle?: string;
}

const REASON_VALUES: { value: ReportReasonType; descKey: string }[] = [
  { value: "spam", descKey: "spam" },
  { value: "violation", descKey: "ruleViolation" },
  { value: "misleading", descKey: "misleading" },
  { value: "rude", descKey: "rudeOrHarassing" },
  { value: "sexual", descKey: "inappropriateContent" },
  { value: "other", descKey: "other" },
];

export default function ReportModal({
  isOpen,
  onClose,
  subjectDid,
  subjectUri,
  subjectHandle,
}: ReportModalProps) {
  const { t } = useTranslation();
  const [selectedReason, setSelectedReason] = useState<ReportReasonType | null>(
    null,
  );
  const [additionalText, setAdditionalText] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [submitted, setSubmitted] = useState(false);

  if (!isOpen) return null;

  const handleSubmit = async () => {
    if (!selectedReason) return;

    setSubmitting(true);
    const success = await reportUser({
      subjectDid: subjectDid,
      subjectUri: subjectUri,
      reasonType: selectedReason,
      reasonText: additionalText || undefined,
    });

    setSubmitting(false);
    if (success) {
      setSubmitted(true);
      setTimeout(() => {
        onClose();
        setSubmitted(false);
        setSelectedReason(null);
        setAdditionalText("");
      }, 1500);
    }
  };

  const handleClose = () => {
    onClose();
    setSelectedReason(null);
    setAdditionalText("");
    setSubmitted(false);
  };

  return (
    <div
      className="fixed inset-0 z-[200] flex items-center justify-center bg-black/50 backdrop-blur-sm animate-fade-in"
      onClick={handleClose}
    >
      <div
        className="bg-white dark:bg-surface-900 rounded-2xl shadow-2xl border border-surface-200 dark:border-surface-700 w-full max-w-md mx-4 overflow-hidden"
        onClick={(e) => e.stopPropagation()}
      >
        {submitted ? (
          <div className="p-8 text-center">
            <div className="w-12 h-12 bg-green-100 dark:bg-green-900/30 rounded-full flex items-center justify-center mx-auto mb-3">
              <Flag size={20} className="text-green-600 dark:text-green-400" />
            </div>
            <h3 className="text-lg font-semibold text-surface-900 dark:text-white">
              {t("report.submitted")}
            </h3>
            <p className="text-surface-500 dark:text-surface-400 text-sm mt-1">
              {t("report.submittedMessage")}
            </p>
          </div>
        ) : (
          <>
            <div className="flex items-center justify-between p-4 border-b border-surface-200 dark:border-surface-700">
              <div className="flex items-center gap-2.5">
                <div className="w-8 h-8 bg-red-100 dark:bg-red-900/30 rounded-full flex items-center justify-center">
                  <Flag size={16} className="text-red-600 dark:text-red-400" />
                </div>
                <div>
                  <h3 className="text-base font-semibold text-surface-900 dark:text-white">
                    {subjectHandle
                      ? t("report.titleUser", { handle: subjectHandle })
                      : t("report.titleGeneric")}
                  </h3>
                  {subjectUri && (
                    <p className="text-xs text-surface-400 dark:text-surface-500">
                      {t("report.reportingContent")}
                    </p>
                  )}
                </div>
              </div>
              <button
                onClick={handleClose}
                className="p-1.5 text-surface-400 hover:text-surface-600 dark:hover:text-surface-300 rounded-lg hover:bg-surface-100 dark:hover:bg-surface-800 transition-colors"
              >
                <X size={18} />
              </button>
            </div>

            <div className="p-4 space-y-2">
              <p className="text-sm font-medium text-surface-700 dark:text-surface-300 mb-3">
                {t("report.issueLabel")}
              </p>
              {REASON_VALUES.map((r) => (
                <button
                  key={r.value}
                  onClick={() => setSelectedReason(r.value)}
                  className={`w-full text-left px-3.5 py-2.5 rounded-xl border transition-all ${
                    selectedReason === r.value
                      ? "border-primary-500 bg-primary-50 dark:bg-primary-900/20"
                      : "border-surface-200 dark:border-surface-700 hover:border-surface-300 dark:hover:border-surface-600"
                  }`}
                >
                  <span
                    className={`text-sm font-medium ${
                      selectedReason === r.value
                        ? "text-primary-700 dark:text-primary-300"
                        : "text-surface-800 dark:text-surface-200"
                    }`}
                  >
                    {t(`report.reasons.${r.descKey}`)}
                  </span>
                </button>
              ))}
            </div>

            {selectedReason && (
              <div className="px-4 pb-2">
                <textarea
                  value={additionalText}
                  onChange={(e) => setAdditionalText(e.target.value)}
                  placeholder={t("report.detailsPlaceholder")}
                  rows={2}
                  className="w-full px-3.5 py-2.5 text-sm rounded-xl border border-surface-200 dark:border-surface-700 bg-white dark:bg-surface-800 text-surface-800 dark:text-surface-200 placeholder:text-surface-400 focus:outline-none focus:ring-2 focus:ring-primary-500/30 focus:border-primary-500 resize-none"
                />
              </div>
            )}

            <div className="flex items-center justify-end gap-2 p-4 border-t border-surface-200 dark:border-surface-700">
              <button
                onClick={handleClose}
                className="px-4 py-2 text-sm font-medium text-surface-600 dark:text-surface-400 hover:text-surface-800 dark:hover:text-surface-200 rounded-xl hover:bg-surface-100 dark:hover:bg-surface-800 transition-colors"
              >
                {t("report.cancel")}
              </button>
              <button
                onClick={handleSubmit}
                disabled={!selectedReason || submitting}
                className="px-4 py-2 text-sm font-medium text-white bg-red-600 hover:bg-red-700 rounded-xl transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {submitting ? t("report.submitting") : t("report.submit")}
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
