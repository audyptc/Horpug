import { useState, useEffect } from 'react'
import { searchService, type SearchResults } from './searchService'

export function useSearch() {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<SearchResults | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(false)

  useEffect(() => {
    if (query.trim().length < 2) {
      setResults(null)
      setError(false)
      return
    }

    const timer = setTimeout(async () => {
      setLoading(true)
      setError(false)
      try {
        const data = await searchService.search(query.trim())
        setResults(data)
      } catch {
        setResults(null)
        setError(true)
      } finally {
        setLoading(false)
      }
    }, 500)

    return () => clearTimeout(timer)
  }, [query])

  return { query, setQuery, results, loading, error }
}
