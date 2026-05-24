import { useTranslation } from 'react-i18next'
import { Construction } from 'lucide-react'

interface PlaceholderProps {
  title: string
}

export function Placeholder({ title }: PlaceholderProps) {
  const { t } = useTranslation()
  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] gap-4 text-center">
      <div className="p-4 rounded-2xl bg-muted">
        <Construction className="w-10 h-10 text-muted-foreground" />
      </div>
      <h1 className="text-2xl font-bold">{title}</h1>
      <p className="text-muted-foreground max-w-xs">{t('common.underDev')}</p>
    </div>
  )
}
