import { useEffect, useState } from 'react'
import { GameEventResponse } from '@/types/api'
import { API_BASE_URL } from '@/lib/apiClient'

export function useGameEvents(gameId: string | null, devAuth: { devUserEmail?: string; devUserName?: string; devUserRole?: string } = {}) {
  const [events, setEvents] = useState<GameEventResponse[]>([])
  const [lastEvent, setLastEvent] = useState<GameEventResponse | null>(null)
  const [error, setError] = useState<Error | null>(null)
  const [connected, setConnected] = useState(false)

  useEffect(() => {
    if (!gameId) return

    let abortController = new AbortController()
    
    async function connect() {
      try {
        const headers: Record<string, string> = {
          'Accept': 'text/event-stream',
        }
        if (devAuth.devUserEmail) headers['X-Dev-User-Email'] = devAuth.devUserEmail
        if (devAuth.devUserName) headers['X-Dev-User-Name'] = devAuth.devUserName
        if (devAuth.devUserRole) headers['X-Dev-User-Role'] = devAuth.devUserRole

        const response = await fetch(`${API_BASE_URL}/games/${gameId}/events`, {
          headers,
          signal: abortController.signal
        })

        if (!response.ok) {
          throw new Error(`Failed to connect to SSE: ${response.status}`)
        }

        setConnected(true)
        const reader = response.body?.getReader()
        if (!reader) throw new Error('No readable stream')

        const decoder = new TextDecoder()
        let buffer = ''

        while (true) {
          const { done, value } = await reader.read()
          if (done) break

          buffer += decoder.decode(value, { stream: true })
          const lines = buffer.split('\n\n')
          buffer = lines.pop() || '' // Keep incomplete event in buffer

          for (const line of lines) {
            if (line.startsWith('data: ')) {
              try {
                const data = JSON.parse(line.slice(6))
                if (data.data) {
                  // The backend might wrap the event in { data: event } or just event
                  const event = data.data as GameEventResponse
                  setEvents(prev => [...prev, event])
                  setLastEvent(event)
                } else {
                  setEvents(prev => [...prev, data])
                  setLastEvent(data)
                }
              } catch (e) {
                console.error('Failed to parse SSE message', e, line)
              }
            }
          }
        }
      } catch (err: any) {
        if (err.name !== 'AbortError') {
          console.error('SSE Error:', err)
          setError(err)
        }
      } finally {
        setConnected(false)
      }
    }

    connect()

    return () => {
      abortController.abort()
    }
  }, [gameId, devAuth.devUserEmail, devAuth.devUserName, devAuth.devUserRole])

  return { events, lastEvent, connected, error }
}
