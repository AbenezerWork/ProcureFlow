import type { ReactNode } from "react";
import { Link } from "react-router-dom";

type AuthCardProps = {
  title: string;
  description: string;
  alternateLabel: string;
  alternateTo: string;
  children: ReactNode;
};

export function AuthCard({
  title,
  description,
  alternateLabel,
  alternateTo,
  children,
}: AuthCardProps) {
  return (
    <main className="auth-screen">
      <section className="auth-panel" aria-label={title}>
        <div className="brand-block auth-brand">
          <div className="brand-mark">PF</div>
          <div>
            <strong>ProcureFlow</strong>
            <span>Sourcing workspace</span>
          </div>
        </div>

        <div className="auth-copy">
          <p className="eyebrow">Procurement operations</p>
          <h1>{title}</h1>
          <p>{description}</p>
        </div>

        {children}

        <Link className="auth-link" to={alternateTo}>
          {alternateLabel}
        </Link>
      </section>
    </main>
  );
}
