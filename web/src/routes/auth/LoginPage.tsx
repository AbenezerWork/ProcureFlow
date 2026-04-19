import { FormEvent, useState } from "react";

import { useAuth } from "@/features/auth/auth-context";
import { Button } from "@/shared/components/ui/Button";
import { Notice } from "@/shared/components/ui/Notice";
import { AuthCard } from "@/routes/auth/AuthCard";

export function LoginPage() {
  const { login } = useAuth();
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setIsSubmitting(true);

    const formData = new FormData(event.currentTarget);

    try {
      await login({
        email: String(formData.get("email") ?? ""),
        password: String(formData.get("password") ?? ""),
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <AuthCard
      title="Sign in"
      description="Continue to requests, approvals, RFQs, vendor records, and award decisions."
      alternateLabel="Create an account"
      alternateTo="/auth/register"
    >
      {error ? <Notice title="Unable to sign in" tone="danger">{error}</Notice> : null}
      <form className="form-stack" onSubmit={handleSubmit}>
        <label>
          <span>Email</span>
          <input name="email" type="email" autoComplete="email" required />
        </label>
        <label>
          <span>Password</span>
          <input name="password" type="password" autoComplete="current-password" required />
        </label>
        <Button disabled={isSubmitting} type="submit">
          {isSubmitting ? "Signing in" : "Sign in"}
        </Button>
      </form>
    </AuthCard>
  );
}
