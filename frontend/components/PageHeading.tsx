import type { ReactNode } from "react";

export default function PageHeading({ eyebrow, title, description, actions }: {
  eyebrow: string;
  title: string;
  description: string;
  actions?: ReactNode;
}) {
  return (
    <div className="crm-page-heading">
      <div>
        <p className="crm-eyebrow">{eyebrow}</p>
        <h1>{title}</h1>
        <p className="crm-page-description">{description}</p>
      </div>
      {actions && <div className="crm-page-actions">{actions}</div>}
    </div>
  );
}
