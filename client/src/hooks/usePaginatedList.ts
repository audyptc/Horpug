import { useState, useCallback, useEffect } from 'react'
import type { PaginationMeta } from '@/types'

type FetchFn<T> = (page: number, perPage: number) => Promise<{ data: T[]; meta: PaginationMeta }>

interface UsePaginatedListOptions {
  initialPerPage?: number
}

export function usePaginatedList<T>(fetchFn: FetchFn<T>, options: UsePaginatedListOptions = {}) {
  const [items, setItems] = useState<T[]>([])
  const [loading, setLoading] = useState(true)
  const [hasError, setHasError] = useState(false)
  const [page, setPage] = useState(1)
  const [perPage, setPerPageState] = useState(options.initialPerPage ?? 20)
  const [total, setTotal] = useState(0)
  const [totalPages, setTotalPages] = useState(1)

  const doFetch = useCallback(
    async (p: number, pp: number) => {
      setLoading(true)
      setHasError(false)
      try {
        const resp = await fetchFn(p, pp)
        setItems(resp.data)
        setTotal(resp.meta.total)
        setTotalPages(resp.meta.total_pages)
      } catch {
        setHasError(true)
      } finally {
        setLoading(false)
      }
    },
    [fetchFn]
  )

  useEffect(() => {
    doFetch(page, perPage)
  }, [doFetch, page, perPage])

  function setPerPage(n: number) {
    setPerPageState(n)
    setPage(1)
  }

  function refresh() {
    doFetch(page, perPage)
  }

  function deletedOnPage() {
    const newPage = items.length === 1 && page > 1 ? page - 1 : page
    setPage(newPage)
    doFetch(newPage, perPage)
  }

  return {
    items,
    loading,
    hasError,
    page,
    perPage,
    total,
    totalPages,
    setPage,
    setPerPage,
    refresh,
    deletedOnPage,
  }
}
