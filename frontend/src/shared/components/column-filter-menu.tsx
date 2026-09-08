import { useState } from 'react'
import { Check, ListFilter } from 'lucide-react'
import { useLanguage } from '@/shared/i18n/language'
import { cn } from '@/shared/lib/utils'
import { Popover, PopoverContent, PopoverTrigger } from '@/shared/components/ui/popover'
import { Button } from '@/shared/components/ui/button'

type ColumnFilterMenuProps<T extends string> = {
  label: string
  // Enumerable side: the neutral "show everything" choice is always first, so
  // any other selection means this column is narrowing the list.
  options?: { value: T; label: string }[]
  optionValue?: T
  onOptionChange?: (value: T) => void
  // Free-text side, applied on submit rather than per keystroke.
  textValue?: string
  onTextChange?: (value: string) => void
}

// A Popover rather than a DropdownMenu: menus own keyboard focus and typeahead,
// which fights any text input placed inside them.
export function ColumnFilterMenu<T extends string>({
  label,
  options,
  optionValue,
  onOptionChange,
  textValue,
  onTextChange,
}: ColumnFilterMenuProps<T>) {
  const { t } = useLanguage()
  const [open, setOpen] = useState(false)
  const [draft, setDraft] = useState(textValue ?? '')

  const isFiltered =
    Boolean(textValue) || (options !== undefined && optionValue !== options[0].value)

  function applyText(value: string) {
    onTextChange?.(value)
    setOpen(false)
  }

  return (
    <Popover
      open={open}
      onOpenChange={(next) => {
        // Re-seed from what's actually applied on open, so an abandoned
        // half-typed value doesn't come back next time.
        if (next) setDraft(textValue ?? '')
        setOpen(next)
      }}
    >
      <PopoverTrigger asChild>
        <button
          type="button"
          title={label}
          aria-label={label}
          className={cn(
            // Roomy enough to be a real tap target on touch screens without
            // pushing the 40px header row taller.
            'inline-flex size-7 shrink-0 items-center justify-center rounded-sm transition-colors hover:bg-accent hover:text-foreground',
            isFiltered ? 'text-primary' : 'opacity-40'
          )}
        >
          <ListFilter size={13} />
        </button>
      </PopoverTrigger>

      <PopoverContent
        align="start"
        collisionPadding={12}
        // The trigger can sit far right in a horizontally scrolled table, so
        // cap the width against the viewport rather than the table.
        className="flex w-[min(14rem,calc(100vw-1.5rem))] flex-col gap-2 p-2"
      >
        {options && (
          <div className="flex flex-col">
            {options.map((option) => (
              <button
                key={option.value}
                type="button"
                onClick={() => {
                  onOptionChange?.(option.value)
                  setOpen(false)
                }}
                className="flex items-center gap-1.5 rounded-sm px-2 py-1.5 text-left text-sm font-normal transition-colors hover:bg-accent hover:text-accent-foreground"
              >
                <Check
                  size={13}
                  className={cn('shrink-0', option.value === optionValue ? 'opacity-100' : 'opacity-0')}
                />
                {option.label}
              </button>
            ))}
          </div>
        )}

        {options && onTextChange && <div className="h-px bg-border" />}

        {onTextChange && (
          <form
            className="flex flex-col gap-2"
            onSubmit={(event) => {
              event.preventDefault()
              applyText(draft.trim())
            }}
          >
            <input
              autoFocus
              type="search"
              value={draft}
              onChange={(event) => setDraft(event.target.value)}
              placeholder={t('filterContainsPlaceholder')}
              className="h-9 rounded-md border border-input bg-transparent px-2 text-sm font-normal"
            />
            <div className="flex justify-end gap-1.5">
              {textValue && (
                <Button type="button" size="sm" variant="ghost" onClick={() => applyText('')}>
                  {t('filterClear')}
                </Button>
              )}
              <Button type="submit" size="sm">
                {t('filterApply')}
              </Button>
            </div>
          </form>
        )}
      </PopoverContent>
    </Popover>
  )
}
