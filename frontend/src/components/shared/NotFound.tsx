import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { Home } from 'lucide-react'

export function NotFound() {
  const { t } = useTranslation()
  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] gap-4 text-center">
      <div className="text-8xl font-bold text-muted-foreground/30">404</div>
      <h1 className="text-2xl font-bold">{t('notFound.title')}</h1>
      <p className="text-muted-foreground">{t('notFound.desc')}</p>
      <Button asChild className="gap-2">
        <Link to="/">
          <Home className="w-4 h-4" />
          {t('common.back')}
        </Link>
      </Button>
    </div>
  )
}
