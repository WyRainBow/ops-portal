/**
 * Server-Sent Events (SSE) client for streaming chat responses.
 */

export type SSEEvent = {
  data: string
  event?: string
  id?: string
}

export type SSEOptions = {
  onMessage?: (event: SSEEvent) => void
  onError?: (error: Error) => void
  onClose?: () => void
  onOpen?: () => void
}

/**
 * SSEClient connects to a Server-Sent Events endpoint and streams responses.
 */
export class SSEClient {
  private eventSource: EventSource | null = null
  private abortController: AbortController | null = null
  private reader: ReadableStreamDefaultReader<Uint8Array> | null = null
  private decoder = new TextDecoder()
  private closed = false

  constructor(private url: string, private options: SSEOptions = {}) {}

  /**
   * Connect using fetch with streaming (preferred for better error handling).
   */
  async connect(token: string, payload: any): Promise<void> {
    this.closed = false
    this.abortController = new AbortController()

    try {
      const response = await fetch(this.url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify(payload),
        signal: this.abortController.signal,
      })

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }

      this.options.onOpen?.()

      // Get the reader from the response body stream
      const reader = response.body?.getReader()
      if (!reader) {
        throw new Error('Response body is not readable')
      }

      this.reader = reader

      // Read the stream
      await this.readStream()

    } catch (error: any) {
      if (error.name === 'AbortError') {
        // Connection was closed by client
        return
      }
      this.options.onError?.(error)
    } finally {
      this.close()
    }
  }

  /**
   * Read the stream line by line, parsing SSE format.
   */
  private async readStream(): Promise<void> {
    if (!this.reader) return

    let buffer = ''

    while (!this.closed) {
      const { done, value } = await this.reader.read()

      if (done) {
        break
      }

      // Decode the chunk and add to buffer
      buffer += this.decoder.decode(value, { stream: true })

      // Process complete lines
      const lines = buffer.split('\n')
      buffer = lines.pop() || '' // Keep the last incomplete line in buffer

      for (const line of lines) {
        this.parseLine(line)
      }
    }
  }

  /**
   * Parse a single SSE line and emit events.
   */
  private parseLine(line: string): void {
    const trimmed = line.trim()

    // Ignore empty lines and comments
    if (!trimmed || trimmed.startsWith(':')) {
      return
    }

    // Parse field: value
    const colonIndex = trimmed.indexOf(':')
    if (colonIndex === -1) {
      return // Invalid line, ignore
    }

    const field = trimmed.slice(0, colonIndex).trim()
    let value = trimmed.slice(colonIndex + 1).trim()

    // Handle special case: ": " after colon means discard the first space
    if (value.startsWith(' ')) {
      value = value.slice(1)
    }

    // Emit event
    this.options.onMessage?.({
      event: field === 'event' ? value : undefined,
      data: field === 'data' ? value : '',
      id: field === 'id' ? value : undefined,
    })
  }

  /**
   * Close the SSE connection.
   */
  close(): void {
    if (this.closed) return

    this.closed = true

    // Close reader
    if (this.reader) {
      this.reader.cancel().catch(() => {})
      this.reader = null
    }

    // Abort fetch
    if (this.abortController) {
      this.abortController.abort()
      this.abortController = null
    }

    // Close EventSource if used
    if (this.eventSource) {
      this.eventSource.close()
      this.eventSource = null
    }

    this.options.onClose?.()
  }

  /**
   * Check if the connection is closed.
   */
  isClosed(): boolean {
    return this.closed
  }
}

/**
 * Chat stream message format from backend.
 */
export type ChatStreamMessage = {
  type: 'content' | 'tool' | 'error' | 'end'
  content?: string
  tool_name?: string
  tool_result?: string
  error?: string
}

/**
 * Stream chat response with streaming updates.
 *
 * @param token - JWT auth token
 * @param payload - Chat request payload
 * @param onChunk - Callback for each content chunk
 * @param onTool - Callback when a tool is invoked
 * @param onError - Callback for errors
 * @returns Promise that resolves when stream ends
 */
export async function streamChat(
  token: string,
  payload: {
    question: string
    id?: string
    stream?: boolean
  },
  callbacks: {
    onChunk?: (chunk: string) => void
    onTool?: (toolName: string, result: string) => void
    onError?: (error: string) => void
    onEnd?: () => void
  }
): Promise<void> {
  const client = new SSEClient('/api/chat/stream', {
    onMessage: (event) => {
      try {
        const data = event.data

        // Parse SSE data
        if (data.startsWith('data: ')) {
          const jsonStr = data.slice(6)
          const msg = JSON.parse(jsonStr) as ChatStreamMessage

          switch (msg.type) {
            case 'content':
              callbacks.onChunk?.(msg.content || '')
              break
            case 'tool':
              callbacks.onTool?.(msg.tool_name || '', msg.tool_result || '')
              break
            case 'error':
              callbacks.onError?.(msg.error || 'Unknown error')
              break
            case 'end':
              callbacks.onEnd?.()
              break
          }
        }
      } catch (e) {
        console.error('Failed to parse SSE message:', e)
      }
    },
    onError: (error) => {
      callbacks.onError?.(error.message)
    },
  })

  await client.connect(token, { ...payload, stream: true })

  return
}

/**
 * Simple text-based streaming for when SSE is not available.
 * Falls back to regular POST with progress indication.
 */
export async function chatWithProgress(
  token: string,
  payload: {
    question: string
    id?: string
  },
  callbacks: {
    onStart?: () => void
    onEnd?: () => void
    onError?: (error: string) => void
  }
): Promise<string> {
  callbacks.onStart?.()

  try {
    const response = await fetch('/api/chat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      body: JSON.stringify(payload),
    })

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`)
    }

    const result = await response.json()
    callbacks.onEnd?.()
    return result.data?.answer || result.answer || ''
  } catch (error: any) {
    callbacks.onError?.(error.message)
    throw error
  }
}
