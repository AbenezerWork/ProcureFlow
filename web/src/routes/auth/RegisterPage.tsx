import { FormEvent, useState } from "react";

import { useAuth } from "@/features/auth/auth-context";
import { AuthCard } from "@/routes/auth/AuthCard";
import { Button } from "@/shared/components/ui/Button";
import { Notice } from "@/shared/components/ui/Notice";

export function RegisterPage() {
  const { register } = useAuth();
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setIsSubmitting(true);

    const formData = new FormData(event.currentTarget);

    try {
      await register({
        full_name: String(formData.get("full_name") ?? ""),
        email: String(formData.get("email") ?? ""),
        password: String(formData.get("password") ?? ""),
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Registration failed");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <AuthCard
      title="Create account"
      description="Start with a user account, then create or select an organization workspace."
      alternateLabel="I already have an account"
      alternateTo="/auth/login"
    >
      {error ? <Notice title="Unable to register" tone="danger">{error}</Notice> : null}
      <form className="form-stack" onSubmit={handleSubmit}>
        <label>
          <span>Full name</span>
          <input name="full_name" type="text" autoComplete="name" required />
        </label>
        <label>
          <span>Email</span>
          <input name="email" type="email" autoComplete="email" required />
        </label>
        <label>
          <span>Password</span>
          <input name="password" type="password" autoComplete="new-password" required />
        </label>
        <Button disabled={isSubmitting} type="submit">
          {isSubmitting ? "Creating account" : "Create account"}
        </Button>
      </form>
    </AuthCard>
  );
}
