import { ChevronLeft, ChevronRight, ExternalLink, Pencil, Trash2 } from 'lucide-react'
import { useLanguage, type TranslationKey } from '@/shared/i18n/language'
import { Badge } from '@/shared/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/components/ui/table'
import { Button } from '@/shared/components/ui/button'
import type { ApiDocument, DocumentCategory } from '../types'
import { DOCUMENT_PAGE_SIZE_OPTIONS, toDateInputValue } from '../utils'

const documentCategoryLabelKeys: Record<DocumentCategory, TranslationKey> = {
  contract: 'documentCategoryContract',
  id_card: 'documentCategoryIdCard',
  receipt: 'documentCategoryReceipt',
  other: 'documentCategoryOther',
}

const documentCategoryBadgeVariant: Record<DocumentCategory, 'default' | 'outline' | 'destructive' | 'secondary'> = {
  contract: 'default',
  id_card: 'secondary',
  receipt: 'outline',
  other: 'outline',
}

type DocumentListCardProps = {
  isLoading: boolean
  loadError: string | null
  deleteError: string | null
  documents: ApiDocument[] | null
  query: string
  onQueryChange: (query: string) => void
  filteredDocuments: ApiDocument[]
  paginatedDocuments: ApiDocument[]
  currentPage: number
  totalPages: number
  rangeStart: number
  rangeEnd: number
  pageSize: number
  onPageSizeChange: (size: number) => void
  onPrevPage: () => void
  onNextPage: () => void
  deletingDocumentId: string | null
  onCreateDocument: () => void
  onEditDocument: (document: ApiDocument) => void
  onDeleteDocument: (document: ApiDocument) => void
}

export function DocumentListCard({
  isLoading,
  loadError,
  deleteError,
  documents,
  query,
  onQueryChange,
  filteredDocuments,
  paginatedDocuments,
  currentPage,
  totalPages,
  rangeStart,
  rangeEnd,
  pageSize,
  onPageSizeChange,
  onPrevPage,
  onNextPage,
  deletingDocumentId,
  onCreateDocument,
  onEditDocument,
  onDeleteDocument,
}: DocumentListCardProps) {
  const { t } = useLanguage()

  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between gap-4">
        <div>
          <CardTitle>{t('menuDocuments')}</CardTitle>
        </div>
        <Button onClick={onCreateDocument} disabled={isLoading}>
          {t('documentCreate')}
        </Button>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {loadError && <p className="resource-error">{loadError}</p>}
        {deleteError && <p className="resource-error">{deleteError}</p>}

        {!loadError && isLoading && <p className="metric-detail">{t('loading')}</p>}

        {!loadError && !isLoading && documents && documents.length === 0 && (
          <p className="metric-detail">{t('documentNoDocuments')}</p>
        )}

        {!loadError && !isLoading && documents && documents.length > 0 && (
          <>
            <label className="flex w-full max-w-md flex-col gap-1.5 text-sm font-medium">
              {t('documentSearchLabel')}
              <input
                type="search"
                className="h-10 rounded-md border border-input bg-transparent px-3 text-sm"
                placeholder={t('documentSearchPlaceholder')}
                value={query}
                onChange={(event) => onQueryChange(event.target.value)}
              />
            </label>

            {filteredDocuments.length === 0 && <p className="metric-detail">{t('documentNoMatching')}</p>}

            {filteredDocuments.length > 0 && (
              <div className="table-wrap">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('documentNameColumn')}</TableHead>
                      <TableHead>{t('documentCategoryColumn')}</TableHead>
                      <TableHead>{t('documentDormitoryColumn')}</TableHead>
                      <TableHead>{t('documentTenantColumn')}</TableHead>
                      <TableHead>{t('documentRoomColumn')}</TableHead>
                      <TableHead>{t('documentUploadedDateColumn')}</TableHead>
                      <TableHead className="text-right">{t('documentActionsColumn')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {paginatedDocuments.map((document) => (
                      <TableRow key={document.id}>
                        <TableCell className="font-semibold">
                          <a
                            href={document.file_url}
                            target="_blank"
                            rel="noreferrer"
                            className="inline-flex items-center gap-1.5 hover:underline"
                            title={t('documentFileLink')}
                          >
                            {document.name}
                            <ExternalLink size={14} />
                          </a>
                        </TableCell>
                        <TableCell>
                          <Badge variant={documentCategoryBadgeVariant[document.category]}>
                            {t(documentCategoryLabelKeys[document.category])}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-muted-foreground">{document.dormitory_name || '—'}</TableCell>
                        <TableCell className="text-muted-foreground">{document.tenant_name || '—'}</TableCell>
                        <TableCell className="text-muted-foreground">{document.room_number || '—'}</TableCell>
                        <TableCell className="text-muted-foreground">
                          {toDateInputValue(document.uploaded_date)}
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex flex-wrap justify-end gap-2">
                            <Button
                              type="button"
                              size="icon"
                              variant="outline"
                              title={t('documentEdit')}
                              aria-label={t('documentEdit')}
                              onClick={() => onEditDocument(document)}
                            >
                              <Pencil />
                            </Button>
                            <Button
                              type="button"
                              size="icon"
                              variant="destructive"
                              title={t('documentDelete')}
                              aria-label={t('documentDelete')}
                              onClick={() => onDeleteDocument(document)}
                              disabled={deletingDocumentId === document.id}
                            >
                              <Trash2 />
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            )}

            {filteredDocuments.length > 0 && (
              <div className="flex items-center justify-between gap-3">
                <p className="text-sm text-muted-foreground">
                  {t('rolePermissionsShowingLabel')} {rangeStart}-{rangeEnd}{' '}
                  {t('rolePermissionsOfLabel')} {filteredDocuments.length} {t('rolePermissionsResultsLabel')}
                  {totalPages > 1 && (
                    <>
                      {' '}
                      · {t('rolePermissionsPageLabel')} {currentPage} / {totalPages}
                    </>
                  )}
                </p>
                <div className="flex items-center gap-3">
                  <label className="flex items-center gap-1.5 text-sm text-muted-foreground">
                    {t('rolePermissionsPageSizeLabel')}
                    <select
                      className="h-9 rounded-md border border-input bg-transparent px-2 text-sm"
                      value={pageSize}
                      onChange={(event) => onPageSizeChange(Number(event.target.value))}
                    >
                      {DOCUMENT_PAGE_SIZE_OPTIONS.map((size) => (
                        <option key={size} value={size}>
                          {size}
                        </option>
                      ))}
                    </select>
                  </label>

                  {totalPages > 1 && (
                    <div className="flex gap-2">
                      <Button
                        type="button"
                        size="icon"
                        variant="outline"
                        title={t('rolePermissionsPrevPage')}
                        aria-label={t('rolePermissionsPrevPage')}
                        disabled={currentPage <= 1}
                        onClick={onPrevPage}
                      >
                        <ChevronLeft />
                      </Button>
                      <Button
                        type="button"
                        size="icon"
                        variant="outline"
                        title={t('rolePermissionsNextPage')}
                        aria-label={t('rolePermissionsNextPage')}
                        disabled={currentPage >= totalPages}
                        onClick={onNextPage}
                      >
                        <ChevronRight />
                      </Button>
                    </div>
                  )}
                </div>
              </div>
            )}
          </>
        )}
      </CardContent>
    </Card>
  )
}
