import { useMemo, useState } from 'react'

export function usePagination<T>(items: T[], initialPageSize: number) {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSizeState] = useState(initialPageSize)

  const totalPages = Math.max(1, Math.ceil(items.length / pageSize))
  const currentPage = Math.min(page, totalPages)
  const rangeStart = items.length === 0 ? 0 : (currentPage - 1) * pageSize + 1
  const rangeEnd = Math.min(currentPage * pageSize, items.length)

  const paginatedItems = useMemo(
    () => items.slice((currentPage - 1) * pageSize, currentPage * pageSize),
    [items, currentPage, pageSize]
  )

  function setPageSize(size: number) {
    setPageSizeState(size)
    setPage(1)
  }

  function resetPage() {
    setPage(1)
  }

  function firstPage() {
    setPage(1)
  }

  function prevPage() {
    setPage((value) => Math.max(1, value - 1))
  }

  function nextPage() {
    setPage((value) => Math.min(totalPages, value + 1))
  }

  function lastPage() {
    setPage(totalPages)
  }

  return {
    page: currentPage,
    setPage,
    pageSize,
    setPageSize,
    totalPages,
    rangeStart,
    rangeEnd,
    paginatedItems,
    resetPage,
    firstPage,
    prevPage,
    nextPage,
    lastPage,
  }
}
