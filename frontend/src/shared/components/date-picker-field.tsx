import { useState } from 'react'
import { CalendarIcon, X } from 'lucide-react'
import { useLanguage } from '@/shared/i18n/language'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Calendar } from '@/shared/components/ui/calendar'
import { Popover, PopoverContent, PopoverTrigger } from '@/shared/components/ui/popover'

function parseDateInput(value: string): Date | undefined {
  if (!value) return undefined
  const [year, month, day] = value.split('-').map(Number)
  return new Date(year, month - 1, day)
}

function formatDateInput(date: Date | undefined): string {
  if (!date) return ''
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

type DatePickerFieldProps = {
  // yyyy-mm-dd, or '' when no date is picked.
  value: string
  onChange: (value: string) => void
  placeholder: string
}

export function DatePickerField({ value, onChange, placeholder }: DatePickerFieldProps) {
  const { t, language } = useLanguage()
  const [open, setOpen] = useState(false)
  const date = parseDateInput(value)
  const dateLocale = language === 'th' ? 'th-TH' : 'en-US'

  return (
    <div className="flex min-w-0 items-center gap-1.5">
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            type="button"
            variant="outline"
            className={cn(
              'h-10 min-w-0 flex-1 justify-start gap-2 px-3 text-sm font-normal',
              !date && 'text-muted-foreground'
            )}
          >
            <CalendarIcon className="size-4 shrink-0" />
            <span className="truncate">{date ? date.toLocaleDateString(dateLocale, { dateStyle: 'medium' }) : placeholder}</span>
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0" align="start">
          <Calendar
            mode="single"
            defaultMonth={date ?? new Date()}
            selected={date}
            onSelect={(next) => {
              onChange(formatDateInput(next))
              setOpen(false)
            }}
          />
        </PopoverContent>
      </Popover>

      {value && (
        <Button
          type="button"
          size="icon"
          variant="ghost"
          className="h-10 w-10 shrink-0 text-muted-foreground"
          title={t('datePickerClear')}
          aria-label={t('datePickerClear')}
          onClick={() => onChange('')}
        >
          <X className="size-4" />
        </Button>
      )}
    </div>
  )
}
