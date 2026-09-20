import { api } from '@/shared/api/client'
import type { ApiDocument } from './types'

// The stored file is only reachable through an authenticated endpoint, so it
// can't be put in an <a href>/<img src>/<iframe src> (those don't carry the
// bearer token). It's fetched as a blob instead and shown from an object URL.
export async function fetchDocumentFile(documentId: string, signal?: AbortSignal): Promise<Blob> {
  const { data } = await api.get<Blob>(`/documents/${documentId}/file`, {
    responseType: 'blob',
    signal,
  })
  return data
}

export async function downloadDocumentFile(document: ApiDocument): Promise<void> {
  const blob = await fetchDocumentFile(document.id)
  const objectUrl = URL.createObjectURL(blob)

  const link = window.document.createElement('a')
  link.href = objectUrl
  link.download = document.file_name || document.name
  window.document.body.appendChild(link)
  link.click()
  link.remove()

  // Revoked on the next tick: revoking synchronously can cancel the download
  // in some browsers before it has started.
  window.setTimeout(() => URL.revokeObjectURL(objectUrl), 0)
}
