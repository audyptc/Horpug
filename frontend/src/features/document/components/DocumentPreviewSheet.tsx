import { useEffect, useState } from 'react'
import axios from 'axios'
import { Download, ExternalLink } from 'lucide-react'
import { useLanguage } from '@/shared/i18n/language'
import { Button } from '@/shared/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/shared/components/ui/sheet'
import type { ApiDocument } from '../types'
import { downloadDocumentFile, fetchDocumentFile } from '../files'
import { getDocumentFileKind, getDocumentKind, getDocumentPreviewUrl, isOpenableUrl } from '../utils'

type DocumentPreviewSheetProps = {
  document: ApiDocument | null
  onOpenChange: (open: boolean) => void
}

// The object URL for an uploaded document's file, fetched with the user's
// credentials. Held per document id so a stale response for a previously
// opened document is never shown against the current one.
function useStoredFileUrl(document: ApiDocument | null) {
  const [state, setState] = useState<{ id: string; url: string | null; failed: boolean } | null>(null)

  const id = document?.has_file ? document.id : null

  useEffect(() => {
    if (!id) return

    const controller = new AbortController()
    let objectUrl: string | null = null

    fetchDocumentFile(id, controller.signal)
      .then((blob) => {
        objectUrl = URL.createObjectURL(blob)
        setState({ id, url: objectUrl, failed: false })
      })
      .catch((err) => {
        if (axios.isCancel(err)) return
        setState({ id, url: null, failed: true })
      })

    return () => {
      controller.abort()
      if (objectUrl) URL.revokeObjectURL(objectUrl)
      // Drop the (now revoked) URL, or reopening the same document would
      // briefly render it before the fresh fetch lands.
      setState(null)
    }
  }, [id])

  if (!id || state?.id !== id) return { url: null, loading: id !== null, failed: false }
  return { url: state.url, loading: false, failed: state.failed }
}

export function DocumentPreviewSheet({ document, onOpenChange }: DocumentPreviewSheetProps) {
  const { t } = useLanguage()
  const [downloading, setDownloading] = useState(false)
  const [downloadFailed, setDownloadFailed] = useState(false)

  const stored = useStoredFileUrl(document)

  const isStored = document?.has_file ?? false
  const kind = document ? getDocumentKind(document) : 'other'
  const linkOpenable = document && !isStored ? isOpenableUrl(document.file_url) : false

  // Stored files preview from the fetched blob (only images and PDFs can be
  // shown inline); linked ones from their own URL.
  const previewUrl = isStored
    ? kind === 'image' || kind === 'pdf'
      ? stored.url
      : null
    : document
      ? getDocumentPreviewUrl(document.file_url)
      : null
  const frameKind = isStored ? kind : document ? getDocumentFileKind(document.file_url) : 'other'

  async function handleDownload() {
    if (!document) return
    setDownloading(true)
    setDownloadFailed(false)
    try {
      await downloadDocumentFile(document)
    } catch {
      setDownloadFailed(true)
    } finally {
      setDownloading(false)
    }
  }

  function renderBody() {
    if (isStored && stored.loading) {
      return <p className="text-sm text-muted-foreground">{t('loading')}</p>
    }
    if (isStored && stored.failed) {
      return <p className="px-4 text-center text-sm text-destructive">{t('documentFileLoadError')}</p>
    }
    if (previewUrl && frameKind === 'image') {
      return <img src={previewUrl} alt={document?.name} className="max-h-full max-w-full object-contain" />
    }
    if (previewUrl && (frameKind === 'pdf' || frameKind === 'google')) {
      return <iframe src={previewUrl} title={document?.name} className="size-full border-0" />
    }
    return (
      <p className="px-4 text-center text-sm text-muted-foreground">
        {isStored || linkOpenable ? t('documentPreviewUnsupported') : t('documentPreviewInvalidLink')}
      </p>
    )
  }

  return (
    <Sheet open={document !== null} onOpenChange={onOpenChange}>
      <SheetContent className="sm:max-w-3xl">
        <SheetHeader>
          <SheetTitle className="truncate">{document?.name}</SheetTitle>
          <SheetDescription>{document?.file_name || t('documentPreviewDescription')}</SheetDescription>
        </SheetHeader>

        <div className="flex min-h-0 flex-1 items-center justify-center overflow-auto rounded-md border border-border bg-muted/30">
          {renderBody()}
        </div>

        {downloadFailed && <p className="resource-error">{t('documentDownloadError')}</p>}

        {/* Always offered: some hosts refuse to be framed, and a file that
            can't be previewed still has to be reachable. */}
        {document && (isStored || linkOpenable) && (
          <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
            {isStored && stored.url && (kind === 'image' || kind === 'pdf') && (
              <Button asChild variant="outline">
                <a href={stored.url} target="_blank" rel="noreferrer">
                  <ExternalLink />
                  {t('documentOpenInNewTab')}
                </a>
              </Button>
            )}
            {!isStored && (
              <Button asChild variant="outline">
                <a href={document.file_url} target="_blank" rel="noreferrer">
                  <ExternalLink />
                  {t('documentOpenInNewTab')}
                </a>
              </Button>
            )}
            {isStored ? (
              <Button type="button" onClick={handleDownload} disabled={downloading}>
                <Download />
                {t('documentDownload')}
              </Button>
            ) : (
              <Button asChild>
                <a href={document.file_url} download target="_blank" rel="noreferrer">
                  <Download />
                  {t('documentDownload')}
                </a>
              </Button>
            )}
          </div>
        )}
      </SheetContent>
    </Sheet>
  )
}
