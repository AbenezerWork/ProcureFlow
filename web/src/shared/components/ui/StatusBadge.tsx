type StatusBadgeProps = {
  status: string;
};

const toneByStatus: Record<string, string> = {
  active: "status-positive",
  approved: "status-positive",
  awarded: "status-positive",
  submitted: "status-attention",
  pending: "status-attention",
  published: "status-attention",
  draft: "status-neutral",
  invited: "status-neutral",
  closed: "status-neutral",
  evaluated: "status-neutral",
  rejected: "status-danger",
  cancelled: "status-danger",
  archived: "status-danger",
  suspended: "status-danger",
  removed: "status-danger",
};

export function StatusBadge({ status }: StatusBadgeProps) {
  const label = status.replaceAll("_", " ");
  return <span className={`status-badge ${toneByStatus[status] ?? "status-neutral"}`}>{label}</span>;
}
