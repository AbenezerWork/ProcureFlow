import type { ReactNode } from "react";

type NoticeProps = {
  title: string;
  children?: ReactNode;
  tone?: "info" | "danger";
};

export function Notice({ title, children, tone = "info" }: NoticeProps) {
  return (
    <div className={`notice notice-${tone}`}>
      <strong>{title}</strong>
      {children ? <p>{children}</p> : null}
    </div>
  );
}
