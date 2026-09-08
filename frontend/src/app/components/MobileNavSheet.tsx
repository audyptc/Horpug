import { SidebarNav } from '@/features/menu/SidebarNav'
import { SidebarFooter } from '@/features/menu/SidebarFooter'
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
      {/* p-0 with the padding pushed onto the header: the menu rows carry
          their own 1.5rem gutter, and doubling it up cost a phone-width sheet
          most of its label space. */}
      <SheetContent side="left" className="mobile-sheet gap-0 p-0">
        <SheetHeader className="p-4 pb-2 pr-12">
          <SheetTitle>{t('mobileMenuTitle')}</SheetTitle>
          <SheetDescription>{t('mobileMenuDescription')}</SheetDescription>
        </SheetHeader>
        {/* The nav is ~20 links across 6 collapsible groups — taller than an
            iPhone 13 mini viewport, so it has to scroll on its own or the
            lower half of the menu is simply unreachable. min-h-0 is what lets
            a flex child actually shrink far enough to do that. */}
        <div className="mobile-sheet-nav min-h-0 flex-1 overflow-y-auto">
          <SidebarNav onNavigate={() => onOpenChange(false)} />
        </div>
        <SidebarFooter />
      </SheetContent>
    </Sheet>
  )
}
