export function CaseArt({ kind }: { kind: string }) {
  switch (kind) {
    case 'train':
      return (
        <svg viewBox="0 0 320 140" className="ink-art">
          <rect x="20" y="50" width="220" height="60" fill="none" stroke="currentColor" strokeWidth="3" />
          <circle cx="60" cy="120" r="14" fill="none" stroke="currentColor" strokeWidth="3" />
          <circle cx="180" cy="120" r="14" fill="none" stroke="currentColor" strokeWidth="3" />
          <path d="M240 70h40l20 20v20H240z" fill="none" stroke="currentColor" strokeWidth="3" />
          <rect x="40" y="65" width="30" height="25" fill="currentColor" opacity="0.15" />
          <rect x="90" y="65" width="30" height="25" fill="currentColor" opacity="0.15" />
        </svg>
      )
    case 'photo':
      return (
        <svg viewBox="0 0 320 140" className="ink-art">
          <rect x="70" y="20" width="180" height="100" fill="none" stroke="currentColor" strokeWidth="3" />
          <circle cx="130" cy="70" r="18" fill="none" stroke="currentColor" strokeWidth="2.5" />
          <circle cx="190" cy="70" r="18" fill="none" stroke="currentColor" strokeWidth="2.5" />
          <path d="M100 105c20-18 40-18 60 0" fill="none" stroke="currentColor" strokeWidth="2" />
          <path d="M160 105c20-18 40-18 60 0" fill="none" stroke="currentColor" strokeWidth="2" />
        </svg>
      )
    case 'garden':
      return (
        <svg viewBox="0 0 320 140" className="ink-art">
          <path d="M40 120 L80 40 L120 120 Z" fill="none" stroke="currentColor" strokeWidth="3" />
          <path d="M130 120 L170 50 L210 120 Z" fill="none" stroke="currentColor" strokeWidth="3" />
          <path d="M220 120 L260 35 L300 120 Z" fill="none" stroke="currentColor" strokeWidth="3" opacity="0.35" />
          <path d="M20 120h280" stroke="currentColor" strokeWidth="2" />
        </svg>
      )
    case 'letter':
      return (
        <svg viewBox="0 0 320 140" className="ink-art">
          <rect x="70" y="35" width="180" height="80" fill="none" stroke="currentColor" strokeWidth="3" />
          <path d="M70 35l90 45 90-45" fill="none" stroke="currentColor" strokeWidth="3" />
          <circle cx="250" cy="40" r="8" fill="currentColor" opacity="0.4" />
        </svg>
      )
    case 'office':
      return (
        <svg viewBox="0 0 320 140" className="ink-art">
          <rect x="60" y="30" width="200" height="90" fill="none" stroke="currentColor" strokeWidth="3" />
          <path d="M80 50h160M80 70h120M80 90h140" stroke="currentColor" strokeWidth="2" opacity="0.5" />
          <rect x="200" y="85" width="40" height="20" fill="currentColor" opacity="0.2" />
        </svg>
      )
    default:
      return (
        <svg viewBox="0 0 320 140" className="ink-art">
          <rect x="30" y="40" width="260" height="70" fill="none" stroke="currentColor" strokeWidth="3" />
          <path d="M50 110V40M270 110V40" stroke="currentColor" strokeWidth="2" />
          <rect x="140" y="55" width="40" height="40" fill="none" stroke="currentColor" strokeWidth="2.5" />
          <circle cx="160" cy="30" r="10" fill="none" stroke="currentColor" strokeWidth="2" />
          <path d="M160 40v15" stroke="currentColor" strokeWidth="2" />
        </svg>
      )
  }
}
