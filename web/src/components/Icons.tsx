type IconProps = { className?: string }

export function CopyIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" className={className} aria-hidden="true" fill="none" stroke="currentColor" strokeWidth="1.8">
      <rect x="8" y="8" width="12" height="13" rx="1.5" />
      <path d="M16 8V4.5A1.5 1.5 0 0 0 14.5 3h-9A1.5 1.5 0 0 0 4 4.5v11A1.5 1.5 0 0 0 5.5 17H8" />
    </svg>
  )
}

export function ArrowIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" className={className} aria-hidden="true" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M4 12h15M13 6l6 6-6 6" />
    </svg>
  )
}

export function LockIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 24 24" className={className} aria-hidden="true" fill="currentColor">
      <path d="M12 2a5 5 0 0 0-5 5v3H6a1.5 1.5 0 0 0-1.5 1.5v9A1.5 1.5 0 0 0 6 22h12a1.5 1.5 0 0 0 1.5-1.5v-9A1.5 1.5 0 0 0 18 10h-1V7a5 5 0 0 0-5-5zm-3 8V7a3 3 0 1 1 6 0v3H9zm3 3.5a1.75 1.75 0 0 1 .9 3.25V19h-1.8v-2.25A1.75 1.75 0 0 1 12 13.5z" />
    </svg>
  )
}

export function PersonIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 32 32" className={className} aria-hidden="true">
      <circle cx="16" cy="16" r="16" fill="currentColor" opacity="0.18" />
      <circle cx="16" cy="12.5" r="5.5" fill="currentColor" />
      <path d="M6.5 27.5c1.2-5.2 5-8 9.5-8s8.3 2.8 9.5 8a15.9 15.9 0 0 1-19 0z" fill="currentColor" />
    </svg>
  )
}
