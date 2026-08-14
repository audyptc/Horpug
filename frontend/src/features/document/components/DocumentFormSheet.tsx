import type { FormEvent } from 'react'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Button } from '@/shared/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/shared/components/ui/sheet'
import type { ApiDormitory } from '@/features/dormitory/types'
import type { ApiTenant } from '@/features/tenant/types'
import type { ApiRoom } from '@/features/room/types'
import type { DocumentCategory } from '../types'
import { DOCUMENT_CATEGORIES } from '../utils'

const documentCategoryLabelKeys: Record<DocumentCategory, TranslationKey> = {
  contract: 'documentCategoryContract',
  id_card: 'documentCategoryIdCard',
  receipt: 'documentCategoryReceipt',
  other: 'documentCategoryOther',
}

type DocumentFormSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  isEdit: boolean
  dormitoryId: string
  onDormitoryIdChange: (dormitoryId: string) => void
  dormitories: ApiDormitory[]
  tenantId: string
  onTenantIdChange: (tenantId: string) => void
  tenants: ApiTenant[]
  roomId: string
  onRoomIdChange: (roomId: string) => void
  rooms: ApiRoom[]
  name: string
  onNameChange: (value: string) => void
  category: DocumentCategory
  onCategoryChange: (category: DocumentCategory) => void
  fileUrl: string
  onFileUrlChange: (value: string) => void
  uploadedDate: string
  onUploadedDateChange: (value: string) => void
  note: string
  onNoteChange: (value: string) => void
  saving: boolean
  error: string | null
  onSubmit: (event: FormEvent<HTMLFormElement>) => void
}

export function DocumentFormSheet({
  open,
  onOpenChange,
  isEdit,
  dormitoryId,
  onDormitoryIdChange,
  dormitories,
  tenantId,
  onTenantIdChange,
  tenants,
  roomId,
  onRoomIdChange,
  rooms,
  name,
  onNameChange,
  category,
  onCategoryChange,
  fileUrl,
  onFileUrlChange,
  uploadedDate,
  onUploadedDateChange,
  note,
  onNoteChange,
  saving,
  error,
  onSubmit,
}: DocumentFormSheetProps) {
  const { t } = useLanguage()

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent>
        <form className="flex h-full flex-col gap-4" onSubmit={onSubmit}>
          <SheetHeader>
            <SheetTitle>{isEdit ? t('documentFormEditTitle') : t('documentFormCreateTitle')}</SheetTitle>
            <SheetDescription>
              {isEdit ? t('documentFormEditDescription') : t('documentFormCreateDescription')}
            </SheetDescription>
          </SheetHeader>

          <div className="flex flex-1 flex-col gap-4 overflow-y-auto pr-1">
            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('documentFormDormitoryLabel')}
              {dormitories.length === 0 ? (
                <p className="text-xs font-normal text-muted-foreground">{t('documentFormNoDormitories')}</p>
              ) : (
                <select
                  className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                  value={dormitoryId}
                  onChange={(event) => onDormitoryIdChange(event.target.value)}
                  disabled={isEdit}
                >
                  <option value="">{t('documentFormDormitoryPlaceholder')}</option>
                  {dormitories.map((dormitory) => (
                    <option key={dormitory.id} value={dormitory.id}>
                      {dormitory.name}
                    </option>
                  ))}
                </select>
              )}
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('documentFormTenantLabel')}
              {tenants.length === 0 ? (
                <p className="text-xs font-normal text-muted-foreground">{t('documentFormNoTenants')}</p>
              ) : (
                <select
                  className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                  value={tenantId}
                  onChange={(event) => onTenantIdChange(event.target.value)}
                >
                  <option value="">{t('documentFormTenantPlaceholder')}</option>
                  {tenants.map((tenant) => (
                    <option key={tenant.id} value={tenant.id}>
                      {tenant.first_name} {tenant.last_name}
                    </option>
                  ))}
                </select>
              )}
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('documentFormRoomLabel')}
              {rooms.length === 0 ? (
                <p className="text-xs font-normal text-muted-foreground">{t('documentFormNoRooms')}</p>
              ) : (
                <select
                  className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                  value={roomId}
                  onChange={(event) => onRoomIdChange(event.target.value)}
                >
                  <option value="">{t('documentFormRoomPlaceholder')}</option>
                  {rooms.map((room) => (
                    <option key={room.id} value={room.id}>
                      {room.room_number} {room.dormitory_name ? `(${room.dormitory_name})` : ''}
                    </option>
                  ))}
                </select>
              )}
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('documentFormNameLabel')}
              <input
                type="text"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={name}
                onChange={(event) => onNameChange(event.target.value)}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('documentFormCategoryLabel')}
              <select
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={category}
                onChange={(event) => onCategoryChange(event.target.value as DocumentCategory)}
              >
                {DOCUMENT_CATEGORIES.map((value) => (
                  <option key={value} value={value}>
                    {t(documentCategoryLabelKeys[value])}
                  </option>
                ))}
              </select>
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('documentFormFileUrlLabel')}
              <input
                type="text"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={fileUrl}
                onChange={(event) => onFileUrlChange(event.target.value)}
                placeholder={t('documentFormFileUrlPlaceholder')}
              />
              <span className="text-xs font-normal text-muted-foreground">{t('documentFormFileUrlHint')}</span>
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('documentFormUploadedDateLabel')}
              <input
                type="date"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                value={uploadedDate}
                onChange={(event) => onUploadedDateChange(event.target.value)}
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium">
              {t('documentFormNoteLabel')}
              <textarea
                className="min-h-20 rounded-md border border-input bg-transparent px-3 py-2 text-sm"
                value={note}
                onChange={(event) => onNoteChange(event.target.value)}
              />
            </label>
          </div>

          {error && <p className="resource-error">{error}</p>}

          <SheetFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              {t('documentFormCancel')}
            </Button>
            <Button type="submit" disabled={saving}>
              {saving ? t('documentSaving') : t('documentFormSave')}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}
