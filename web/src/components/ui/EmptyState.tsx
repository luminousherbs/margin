import React from "react";
import { clsx } from "clsx";

interface EmptyStateProps {
  icon?: React.ReactNode;
  title?: string;
  message: string;
  action?: React.ReactNode | { label: string; onClick: () => void };
  className?: string;
}

export default function EmptyState({
  icon,
  title,
  message,
  action,
  className,
}: EmptyStateProps) {
  return (
    <div
      className={clsx(
        "relative text-center py-14 px-6 overflow-hidden",
        "bg-gradient-to-b from-surface-50/80 to-surface-50/20 dark:from-surface-800/60 dark:to-surface-800/20 rounded-2xl",
        "border border-dashed border-surface-200 dark:border-surface-700",
        className,
      )}
    >
      {icon && (
        <div className="relative flex justify-center mb-4">
          <div className="absolute inset-0 flex justify-center items-center">
            <div className="h-16 w-16 rounded-full bg-primary-100/60 dark:bg-primary-900/20 blur-xl" />
          </div>
          <div className="relative text-surface-400 dark:text-surface-500">
            {icon}
          </div>
        </div>
      )}
      {title && (
        <h3 className="text-lg font-display font-semibold text-surface-900 dark:text-white mb-2">
          {title}
        </h3>
      )}
      <p className="text-surface-500 dark:text-surface-400 max-w-sm mx-auto leading-relaxed">
        {message}
      </p>
      {action && (
        <div className="mt-6">
          {typeof action === "object" &&
          "label" in action &&
          "onClick" in action ? (
            <button
              onClick={action.onClick}
              className="inline-flex items-center gap-2 px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-lg transition-colors"
            >
              {action.label}
            </button>
          ) : (
            action
          )}
        </div>
      )}
    </div>
  );
}
