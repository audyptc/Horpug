import { useEffect, useState } from 'react'
import axios from 'axios'
import { Download, Eye, File, FileImage, FileText, Pencil, Plus } from 'lucide-react'
import { api, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Button } from '@/shared/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/shared/components/ui/sheet'
import type { ApiDocument, DocumentCategory } from '../types'
import { downloadDocumentFile } from '../files'
import { useDocumentForm, type DocumentFormPrefill } from '../useDocumentForm'
import { getDocumentKind, isOpenableUrl, toDateInputValue } from '../utils'
import { DocumentFormSheet } from './DocumentFormSheet'
import { DocumentPreviewSheet } from './DocumentPreviewSheet'

const PAGE_SIZE = 100

const categoryLabelKeys: Record<DocumentCategory, TranslationKey> = {
  contract: 'documentCategoryContract',
  id_card: 'documentCategoryIdCard',
  receipt: 'documentCategoryReceipt',
  other: 'documentCategoryOther',
}

const kindIcons = { image: FileImage, pdf: FileText, google: FileText, other: File } as const

type EntityDocumentsSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  // What the documents belong to: exactly one of the two is used to filter the
  // list, and the same one prefills the "add" form.
  scope: { tenantId: string } | { roomId: string }
  title: string
  prefill: DocumentFormPrefill
}

// The documents attached to one tenant or one room, with the same
// view / download / add / edit actions as the Documents page. Deleting stays
// on the Documents page.
export function EntityDocumentsSheet({ open, onOpenChange, scope, title, prefill }: EntityDocumentsSheetProps) {
  const { t } = useLanguage()

  const [documents, setDocuments] = useState<ApiDocument[] | null>(null)
  const [total, setTotal] = useState(0)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)
  const [refreshToken, setRefreshToken] = useState(0)
  const [previewDocument, setPreviewDocument] = useState<ApiDocument | null>(null)

  const form = useDocumentForm({ onSaved: () => setRefreshToken((value) => value + 1) })

  const tenantId = 'tenantId' in scope ? scope.tenantId : undefined
  const roomId = 'roomId' in scope ? scope.roomId : undefined

  useEffect(() => {
    if (!open) return

    const controller = new AbortController()

    api
      .get<ApiPage<ApiDocument[]>>('/documents', {
        signal: controller.signal,
        params: {
          tenant_id: tenantId,
          room_id: roomId,
          per_page: PAGE_SIZE,
          sort: 'uploaded_date',
          order: 'desc',
        },
      })
      .then(({ data }) => {
        setDocuments(data.data)
        setTotal(data.meta.total)
        setLoadError(null)
      })
      .catch((err) => {
        if (axios.isCancel(err)) return
        setLoadError(extractErrorMessage(err, t('resourceLoadError')))
      })

    // Reset on close/scope change, so reopening for someone else never flashes
    // the previous list.
    return () => {
      controller.abort()
      setDocuments(null)
      setLoadError(null)
      setActionError(null)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, tenantId, roomId, refreshToken])

  async function handleDownload(document: ApiDocument) {
    setActionError(null)
    try {
      await downloadDocumentFile(document)
    } catch {
      setActionError(t('documentDownloadError'))
    }
  }

  const isLoading = !loadError && documents === null

  return (
    <>
      <Sheet open={open} onOpenChange={onOpenChange}>
        <SheetContent className="sm:max-w-xl">
          <SheetHeader>
            <SheetTitle className="truncate">{title}</SheetTitle>
            <SheetDescription>
              {documents ? `${total} ${t('documentEntityCountUnit')}` : t('documentEntityDescription')}
            </SheetDescription>
          </SheetHeader>

          <div>
            <Button type="button" onClick={() => form.openCreate(prefill)}>
              <Plus />
              {t('documentCreate')}
            </Button>
          </div>

          <div className="flex min-h-0 flex-1 flex-col gap-2 overflow-y-auto">
            {loadError && <p className="resource-error">{loadError}</p>}
            {actionError && <p className="resource-error">{actionError}</p>}
            {isLoading && <p className="metric-detail">{t('loading')}</p>}

            {documents?.length === 0 && <p className="metric-detail">{t('documentEntityEmpty')}</p>}

            {documents?.map((document) => {
              const Icon = kindIcons[getDocumentKind(document)]
              const canOpenLink = !document.has_file && isOpenableUrl(document.file_url)

              return (
                <div
                  key={document.id}
                  className="flex items-center gap-3 rounded-md border border-border p-3"
                >
                  <Icon className="size-5 shrink-0 text-muted-foreground" />

                  <button
                    type="button"
                    className="flex min-w-0 flex-1 flex-col items-start gap-1 text-left"
                    title={t('documentPreview')}
                    onClick={() => setPreviewDocument(document)}
                  >
                    <span className="w-full truncate text-sm font-semibold hover:underline">
                      {document.name}
                    </span>
                    <span className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                      <Badge variant="outline">{t(categoryLabelKeys[document.category])}</Badge>
                      {toDateInputValue(document.uploaded_date)}
                    </span>
                  </button>

                  <div className="flex shrink-0 gap-1.5">
                    <Button
                      type="button"
                      size="icon"
                      variant="outline"
                      title={t('documentPreview')}
                      aria-label={t('documentPreview')}
                      onClick={() => setPreviewDocument(document)}
                    >
                      <Eye />
                    </Button>
                    {document.has_file && (
                      <Button
                        type="button"
                        size="icon"
                        variant="outline"
                        title={t('documentDownload')}
                        aria-label={t('documentDownload')}
                        onClick={() => handleDownload(document)}
                      >
                        <Download />
                      </Button>
                    )}
                    {canOpenLink && (
                      <Button asChild size="icon" variant="outline">
                        <a
                          href={document.file_url}
                          download
                          target="_blank"
                          rel="noreferrer"
                          title={t('documentDownload')}
                          aria-label={t('documentDownload')}
                        >
                          <Download />
                        </a>
                      </Button>
                    )}
                    <Button
                      type="button"
                      size="icon"
                      variant="outline"
                      title={t('documentEdit')}
                      aria-label={t('documentEdit')}
                      onClick={() => form.openEdit(document)}
                    >
                      <Pencil />
                    </Button>
                  </div>
                </div>
              )
            })}

            {documents && total > documents.length && (
              <p className="metric-detail">{t('documentEntityMore')}</p>
            )}
          </div>
        </SheetContent>
      </Sheet>

      <DocumentPreviewSheet
        document={previewDocument}
        onOpenChange={(next) => !next && setPreviewDocument(null)}
      />
      <DocumentFormSheet {...form.sheetProps} />
    </>
  )
}
