import { useTranslation } from 'react-i18next'
import { MoreHorizontal, Pencil, Trash2, CheckCircle2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { EmptyState } from '@/components/shared/EmptyState'
import type { ApiBill, BillStatus } from '@/types/bill'
import { cn } from '@/lib/utils'
import { formatDate } from '@/lib/dateUtils'

const STATUS_VARIANTS: Record<BillStatus, 'default' | 'secondary' | 'destructive'> = {
  unpaid: 'secondary',
  paid: 'default',
  overdue: 'destructive',
}

interface BillTableProps {
  bills: ApiBill[]
  hasFilters: boolean
  onEdit: (bill: ApiBill) => void
  onMarkPaid: (bill: ApiBill) => void
  onDelete: (bill: ApiBill) => void
  onClearFilters: () => void
}

export function BillTable({ bills, hasFilters, onEdit, onMarkPaid, onDelete, onClearFilters }: BillTableProps) {
  const { t } = useTranslation()

  if (bills.length === 0) {
    return (
      <EmptyState
        message={t('bills.noBills')}
        onClear={hasFilters ? onClearFilters : undefined}
        clearLabel={hasFilters ? t('bills.clearFilters') : undefined}
      />
    )
  }

  return (
    <>
      {/* Desktop table */}
      <div className="hidden md:block overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b bg-muted/40">
              <th className="text-left px-6 py-3 font-medium text-muted-foreground">{t('bills.colTenant')}</th>
              <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('bills.colRoom')}</th>
              <th className="text-left px-4 py-3 font-medium text-muted-foreground">{t('bills.colMonth')}</th>
              <th className="text-right px-4 py-3 font-medium text-muted-foreground">{t('bills.colRent')}</th>
              <th className="text-right px-4 py-3 font-medium text-muted-foreground">{t('bills.colElectric')}</th>
              <th className="text-right px-4 py-3 font-medium text-muted-foreground">{t('bills.colWater')}</th>
              <th className="text-right px-4 py-3 font-medium text-muted-foreground">{t('bills.colTotal')}</th>
              <th className="text-center px-4 py-3 font-medium text-muted-foreground">{t('bills.colStatus')}</th>
              <th className="text-right px-6 py-3 font-medium text-muted-foreground">{t('bills.colActions')}</th>
            </tr>
          </thead>
          <tbody>
            {bills.map((bill, i) => (
              <tr
                key={bill.id}
                className={cn('border-b transition-colors hover:bg-muted/30', i === bills.length - 1 && 'border-0')}
              >
                <td className="px-6 py-4 font-medium">{bill.tenant_first_name} {bill.tenant_last_name}</td>
                <td className="px-4 py-4 text-muted-foreground">{bill.room_number}</td>
                <td className="px-4 py-4 text-muted-foreground">{formatDate(bill.billing_month)}</td>
                <td className="px-4 py-4 text-right text-muted-foreground">{bill.rent_amount.toLocaleString()}</td>
                <td className="px-4 py-4 text-right text-muted-foreground">{bill.electric_amount.toLocaleString()}</td>
                <td className="px-4 py-4 text-right text-muted-foreground">{bill.water_amount.toLocaleString()}</td>
                <td className="px-4 py-4 text-right font-semibold">
                  {bill.total_amount.toLocaleString()}
                  <span className="text-xs text-muted-foreground ml-1">฿</span>
                </td>
                <td className="px-4 py-4 text-center">
                  <Badge variant={STATUS_VARIANTS[bill.status]}>
                    {t(`bills.statuses.${bill.status}`)}
                  </Badge>
                </td>
                <td className="px-6 py-4 text-right">
                  <BillActions bill={bill} onEdit={onEdit} onMarkPaid={onMarkPaid} onDelete={onDelete} />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Mobile cards */}
      <div className="md:hidden divide-y">
        {bills.map((bill) => (
          <div key={bill.id} className="p-4 flex items-start gap-3">
            <div className="flex-1 min-w-0 space-y-1">
              <p className="text-sm font-medium">{bill.tenant_first_name} {bill.tenant_last_name}</p>
              <p className="text-xs text-muted-foreground">
                {t('bills.colRoom')} {bill.room_number} · {formatDate(bill.billing_month)}
              </p>
              <p className="text-sm font-semibold">{bill.total_amount.toLocaleString()} ฿</p>
              <Badge variant={STATUS_VARIANTS[bill.status]} className="text-xs">
                {t(`bills.statuses.${bill.status}`)}
              </Badge>
            </div>
            <BillActions bill={bill} onEdit={onEdit} onMarkPaid={onMarkPaid} onDelete={onDelete} />
          </div>
        ))}
      </div>
    </>
  )
}

interface BillActionsProps {
  bill: ApiBill
  onEdit: (bill: ApiBill) => void
  onMarkPaid: (bill: ApiBill) => void
  onDelete: (bill: ApiBill) => void
}

function BillActions({ bill, onEdit, onMarkPaid, onDelete }: BillActionsProps) {
  const { t } = useTranslation()
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" className="h-8 w-8 shrink-0">
          <MoreHorizontal className="w-4 h-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        {bill.status !== 'paid' && (
          <DropdownMenuItem onClick={() => onMarkPaid(bill)} className="gap-2">
            <CheckCircle2 className="w-4 h-4" /> {t('bills.markPaid')}
          </DropdownMenuItem>
        )}
        <DropdownMenuItem onClick={() => onEdit(bill)} className="gap-2">
          <Pencil className="w-4 h-4" /> {t('common.edit')}
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          className="gap-2 text-destructive focus:text-destructive"
          onClick={() => onDelete(bill)}
        >
          <Trash2 className="w-4 h-4" /> {t('common.delete')}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
