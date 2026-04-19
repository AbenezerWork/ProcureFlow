export function formString(formData: FormData, key: string) {
  const value = String(formData.get(key) ?? "").trim();
  return value || undefined;
}

export function formNumber(formData: FormData, key: string) {
  const value = formString(formData, key);
  return value === undefined ? undefined : Number(value);
}

export function compactRecord<T extends Record<string, unknown>>(input: T) {
  return Object.fromEntries(
    Object.entries(input).filter(([, value]) => value !== undefined && value !== ""),
  ) as Partial<T>;
}
