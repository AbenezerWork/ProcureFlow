export function formatDateTime(value?: string | null) {
  if (!value) {
    return "Not set";
  }

  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  }).format(new Date(value));
}

export function formatAmount(value?: string | null, currencyCode = "USD") {
  if (!value) {
    return "Not set";
  }

  const amount = Number(value);
  if (Number.isNaN(amount)) {
    return `${value} ${currencyCode}`;
  }

  return new Intl.NumberFormat(undefined, {
    style: "currency",
    currency: currencyCode,
    maximumFractionDigits: 2,
  }).format(amount);
}

export function initials(name?: string | null) {
  if (!name) {
    return "PF";
  }

  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase())
    .join("");
}
