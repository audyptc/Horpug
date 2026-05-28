import { useEffect, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Search, Loader2, User, DoorOpen, FileText, ScrollText } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import { useSearch } from './useSearch'
import type { SearchBill, SearchContract, SearchRoom, SearchTenant } from './searchService'

interface SearchDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

function billStatusVariant(status: string): 'default' | 'secondary' | 'destructive' | 'outline' {
  if (status === 'paid') return 'default'
  if (status === 'overdue') return 'destructive'
  return 'secondary'
}

function contractStatusVariant(status: string): 'default' | 'secondary' | 'destructive' | 'outline' {
  if (status === 'active') return 'default'
  if (status === 'terminated') return 'destructive'
  return 'secondary'
}

function formatMonth(isoDate: string, lang: string) {
  return new Date(isoDate).toLocaleDateString(lang === 'th' ? 'th-TH' : 'en-US', {
    year: 'numeric',
    month: 'short',
  })
}

export function SearchDialog({ open, onOpenChange }: SearchDialogProps) {
  const { t, i18n } = useTranslation()
  const navigate = useNavigate()
  const inputRef = useRef<HTMLInputElement>(null)
  const { query, setQuery, results, loading, error } = useSearch()

  useEffect(() => {
    if (open) {
      setTimeout(() => inputRef.current?.focus(), 50)
    } else {
      setQuery('')
    }
  }, [open, setQuery])

  function goTo(path: string) {
    navigate(path)
    onOpenChange(false)
  }

  const hasResults = results && (
    results.tenants.length > 0 ||
    results.rooms.length > 0 ||
    results.bills.length > 0 ||
    results.contracts.length > 0
  )

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-xl p-0 gap-0 overflow-hidden">
        <DialogTitle className="sr-only">{t('header.searchTitle')}</DialogTitle>
        <DialogDescription className="sr-only">{t('header.searchHint')}</DialogDescription>

        {/* Input */}
        <div className="flex items-center px-4 py-3 gap-3">
          <Search className="w-4 h-4 text-muted-foreground shrink-0" />
          <input
            ref={inputRef}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={t('header.searchPlaceholder')}
            className="flex-1 bg-transparent outline-none text-sm"
          />
          {loading && <Loader2 className="w-4 h-4 animate-spin text-muted-foreground shrink-0" />}
        </div>

        <Separator />

        {/* Results */}
        <ScrollArea className="max-h-[60vh]">
          {query.trim().length < 2 ? (
            <p className="py-8 text-center text-sm text-muted-foreground">
              {t('header.searchHint')}
            </p>
          ) : error ? (
            <p className="py-8 text-center text-sm text-destructive">
              {t('header.searchError')}
            </p>
          ) : !hasResults && !loading ? (
            <p className="py-8 text-center text-sm text-muted-foreground">
              {t('header.searchNoResults')}
            </p>
          ) : (
            <div className="py-2">
              <TenantSection tenants={results?.tenants ?? []} t={t} onSelect={() => goTo('/tenants')} />
              <RoomSection rooms={results?.rooms ?? []} t={t} onSelect={() => goTo('/rooms')} />
              <BillSection bills={results?.bills ?? []} t={t} lang={i18n.language} onSelect={() => goTo('/bills')} />
              <ContractSection contracts={results?.contracts ?? []} t={t} lang={i18n.language} onSelect={() => goTo('/contracts')} />
            </div>
          )}
        </ScrollArea>
      </DialogContent>
    </Dialog>
  )
}

interface SectionProps<T> {
  t: (key: string) => string
  onSelect: () => void
  items?: T[]
}

function SectionHeader({ label }: { label: string }) {
  return (
    <p className="px-4 py-1.5 text-xs font-medium text-muted-foreground uppercase tracking-wide">
      {label}
    </p>
  )
}

function ResultItem({ onClick, children }: { onClick: () => void; children: React.ReactNode }) {
  return (
    <button
      onClick={onClick}
      className="w-full flex items-center gap-3 px-4 py-2.5 text-sm hover:bg-accent hover:text-accent-foreground transition-colors text-left"
    >
      {children}
    </button>
  )
}

function TenantSection({ tenants, t, onSelect }: SectionProps<SearchTenant> & { tenants: SearchTenant[] }) {
  if (tenants.length === 0) return null
  return (
    <div>
      <SectionHeader label={t('header.searchTenants')} />
      {tenants.map((tenant) => (
        <ResultItem key={tenant.id} onClick={onSelect}>
          <User className="w-4 h-4 text-muted-foreground shrink-0" />
          <span className="flex-1 font-medium">
            {tenant.first_name} {tenant.last_name}
          </span>
          <span className="text-muted-foreground text-xs">{tenant.phone}</span>
        </ResultItem>
      ))}
    </div>
  )
}

function RoomSection({ rooms, t, onSelect }: SectionProps<SearchRoom> & { rooms: SearchRoom[] }) {
  if (rooms.length === 0) return null
  return (
    <div>
      <SectionHeader label={t('header.searchRooms')} />
      {rooms.map((room) => (
        <ResultItem key={room.id} onClick={onSelect}>
          <DoorOpen className="w-4 h-4 text-muted-foreground shrink-0" />
          <span className="flex-1 font-medium">{room.room_number}</span>
          <span className="text-muted-foreground text-xs">
            {t('rooms.floorN', { n: room.floor })} · {room.status}
          </span>
        </ResultItem>
      ))}
    </div>
  )
}

function BillSection({
  bills, t, lang, onSelect,
}: SectionProps<SearchBill> & { bills: SearchBill[]; lang: string }) {
  if (bills.length === 0) return null
  return (
    <div>
      <SectionHeader label={t('header.searchBills')} />
      {bills.map((bill) => (
        <ResultItem key={bill.id} onClick={onSelect}>
          <FileText className="w-4 h-4 text-muted-foreground shrink-0" />
          <span className="flex-1">
            <span className="font-medium">{bill.tenant_first_name} {bill.tenant_last_name}</span>
            <span className="text-muted-foreground ml-2 text-xs">
              {bill.room_number} · {formatMonth(bill.billing_month, lang)}
            </span>
          </span>
          <Badge variant={billStatusVariant(bill.status)} className="text-xs">
            {t(`bills.statuses.${bill.status}`)}
          </Badge>
        </ResultItem>
      ))}
    </div>
  )
}

function ContractSection({
  contracts, t, lang, onSelect,
}: SectionProps<SearchContract> & { contracts: SearchContract[]; lang: string }) {
  if (contracts.length === 0) return null
  return (
    <div>
      <SectionHeader label={t('header.searchContracts')} />
      {contracts.map((contract) => (
        <ResultItem key={contract.id} onClick={onSelect}>
          <ScrollText className="w-4 h-4 text-muted-foreground shrink-0" />
          <span className="flex-1">
            <span className="font-medium">{contract.tenant_first_name} {contract.tenant_last_name}</span>
            <span className="text-muted-foreground ml-2 text-xs">
              {contract.room_number}
            </span>
          </span>
          <Badge variant={contractStatusVariant(contract.status)} className="text-xs">
            {t(`contracts.statuses.${contract.status}`)}
          </Badge>
        </ResultItem>
      ))}
    </div>
  )
}
