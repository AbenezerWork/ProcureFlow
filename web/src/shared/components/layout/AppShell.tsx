import { NavLink, Outlet } from "react-router-dom";

import { useAuth } from "@/features/auth/auth-context";
import { initials } from "@/shared/lib/format";
import { OrganizationSwitcher } from "@/shared/components/layout/OrganizationSwitcher";

const navItems = [
  { to: "/app/dashboard", label: "Dashboard" },
  { to: "/app/requests", label: "Requests" },
  { to: "/app/approvals", label: "Approvals" },
  { to: "/app/rfqs", label: "RFQs" },
  { to: "/app/quotations", label: "Quotations" },
  { to: "/app/awards", label: "Awards" },
  { to: "/app/vendors", label: "Vendors" },
  { to: "/app/members", label: "Members" },
  { to: "/app/activity", label: "Activity" },
  { to: "/app/settings", label: "Settings" },
];

export function AppShell() {
  const { user, logout } = useAuth();

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="brand-block">
          <div className="brand-mark">PF</div>
          <div>
            <strong>ProcureFlow</strong>
            <span>Sourcing workspace</span>
          </div>
        </div>

        <nav className="sidebar-nav" aria-label="Primary navigation">
          {navItems.map((item) => (
            <NavLink key={item.to} to={item.to}>
              {item.label}
            </NavLink>
          ))}
        </nav>
      </aside>

      <div className="workspace">
        <header className="topbar">
          <OrganizationSwitcher />
          <div className="user-menu">
            <div className="avatar" aria-hidden="true">
              {initials(user?.full_name)}
            </div>
            <div>
              <strong>{user?.full_name}</strong>
              <span>{user?.email}</span>
            </div>
            <button type="button" onClick={logout}>
              Sign out
            </button>
          </div>
        </header>

        <main className="main-content">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
