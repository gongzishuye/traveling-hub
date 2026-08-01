import type { ReactNode, Ref } from 'react'

type CollapsiblePanelProps = {
  id: string
  title: string
  expanded: boolean
  onToggle(): void
  className?: string
  toggleRef?: Ref<HTMLButtonElement>
  children: ReactNode
}

export function CollapsiblePanel({
  id,
  title,
  expanded,
  onToggle,
  className,
  toggleRef,
  children,
}: CollapsiblePanelProps) {
  const action = expanded ? '收起' : '展开'

  return (
    <div className={`collapsible-panel ${className ?? ''}`}>
      <button
        type="button"
        className="collapsible-panel__toggle"
        ref={toggleRef}
        aria-controls={`${id}-content`}
        aria-expanded={expanded}
        onClick={onToggle}
      >
        {action}{title}
      </button>
      {expanded ? <div id={`${id}-content`}>{children}</div> : null}
    </div>
  )
}
