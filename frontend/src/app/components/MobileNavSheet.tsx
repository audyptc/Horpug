import { SidebarNav } from '@/features/menu/SidebarNav'
import { Button } from '@/shared/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/shared/components/ui/sheet'
import { useLanguage } from '@/shared/i18n/language'
import { Menu } from 'lucide-react'

export function MobileNavSheet({
  open,
  onOpenChange,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  const { t } = useLanguage()

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetTrigger asChild>
        <Button size="icon" variant="ghost" className="mobile-only" aria-label={t('openMenu')}>
          <Menu size={18} />
        </Button>
      </SheetTrigger>
      <SheetContent side="left" className="mobile-sheet">
        <SheetHeader>
          <SheetTitle>{t('mobileMenuTitle')}</SheetTitle>
          <SheetDescription>{t('mobileMenuDescription')}</SheetDescription>
        </SheetHeader>
        <div className="mobile-sheet-nav">
          <SidebarNav onNavigate={() => onOpenChange(false)} />
        </div>
      </SheetContent>
    </Sheet>
  )
}
