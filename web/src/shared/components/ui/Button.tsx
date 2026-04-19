import type { ButtonHTMLAttributes } from "react";

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "primary" | "secondary" | "ghost" | "danger";
};

const variants = {
  primary: "button button-primary",
  secondary: "button button-secondary",
  ghost: "button button-ghost",
  danger: "button button-danger",
};

export function Button({ className = "", variant = "primary", ...props }: ButtonProps) {
  return <button className={`${variants[variant]} ${className}`} {...props} />;
}
