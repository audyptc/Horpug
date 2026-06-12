import { useState } from 'react'
import { format } from 'date-fns'
import { th } from 'date-fns/locale'
import { CalendarIcon, ChevronLeft, ChevronRight } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'

type Props = {
  value: string // "YYYY-MM" or ""
  onChange: (value: string) => void
  placeholder?: string
  disabled?: boolean
  className?: string
}

const MONTHS_TH = [
  'ม.ค.', 'ก.พ.', 'มี.ค.', 'เม.ย.', 'พ.ค.', 'มิ.ย.',
  'ก.ค.', 'ส.ค.', 'ก.ย.', 'ต.ค.', 'พ.ย.', 'ธ.ค.',
]

export function MonthPicker({ value, onChange, placeholder = 'เลือกเดือน', disabled, className }: Props) {
  const today = new Date()
  const [viewYear, setViewYear] = useState(() => {
    if (value) return parseInt(value.split('-')[0])
    return today.getFullYear()
  })

  const selectedYear = value ? parseInt(value.split('-')[0]) : null
  const selectedMonth = value ? parseInt(value.split('-')[1]) - 1 : null // 0-indexed

  function handleSelect(monthIndex: number) {
    const pad = String(monthIndex + 1).padStart(2, '0')
    onChange(`${viewYear}-${pad}`)
  }

  const displayText = value
    ? format(new Date(parseInt(value.split('-')[0]), parseInt(value.split('-')[1]) - 1, 1), 'MMMM yyyy', { locale: th })
    : null

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          disabled={disabled}
          className={cn(
            'w-full justify-start text-left font-normal',
            !value && 'text-muted-foreground',
            className,
          )}
        >
          <CalendarIcon className="mr-2 h-4 w-4" />
          {displayText ?? placeholder}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-60 p-3" align="start">
        <div className="flex items-center justify-between mb-3">
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7"
            onClick={() => setViewYear((y) => y - 1)}
          >
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <span className="text-sm font-medium">{viewYear + 543}</span>
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7"
            onClick={() => setViewYear((y) => y + 1)}
          >
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
        <div className="grid grid-cols-3 gap-1.5">
          {MONTHS_TH.map((name, i) => {
            const isSelected = selectedYear === viewYear && selectedMonth === i
            return (
              <Button
                key={i}
                variant={isSelected ? 'default' : 'ghost'}
                size="sm"
                className="h-8 text-xs"
                onClick={() => handleSelect(i)}
              >
                {name}
              </Button>
            )
          })}
        </div>
      </PopoverContent>
    </Popover>
  )
}
