import React from "react";

/**
 * User-group icon matching the Figma design (Font Awesome Duotone "user-group" style).
 * Uses currentColor so the parent can control the fill colour.
 */
export function UserGroupIcon({
  className,
  style,
}: {
  className?: string;
  style?: React.CSSProperties;
}): React.ReactElement {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="currentColor"
      className={className}
      style={style}
      aria-hidden="true"
    >
      {/* Back person — secondary/dimmed */}
      <g opacity="0.4">
        <circle cx="8.5" cy="6.5" r="2.5" />
        <path d="M3 20v-1.2C3 16.6 5.5 14.5 8.5 14.5c.9 0 1.7.2 2.5.5-.7 1.1-1.1 2.5-1.1 3.9V20H3z" />
      </g>
      {/* Front person — primary */}
      <circle cx="15" cy="8" r="3.5" />
      <path d="M15 13.5c-4.1 0-7.5 2.8-7.5 6.2V21h15v-1.3c0-3.4-3.4-6.2-7.5-6.2z" />
    </svg>
  );
}
