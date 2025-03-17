import '@testing-library/jest-dom'
import { expect, afterEach, beforeAll, vi } from 'vitest'
import { cleanup } from '@testing-library/react'
import * as matchers from '@testing-library/jest-dom/matchers'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider } from '../contexts/AuthContext'
import { ConfigProvider, App as AntApp } from 'antd'
import React from 'react'

// Extend Vitest's expect method with methods from react-testing-library
expect.extend(matchers)

// Mock window.matchMedia for Ant Design
beforeAll(() => {
  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: vi.fn().mockImplementation((query: string) => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: vi.fn(), // Deprecated
      removeListener: vi.fn(), // Deprecated
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn()
    }))
  })
})

// Create a new QueryClient for each test
const createTestQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: {
        retry: false
      }
    }
  })

// Create a wrapper component that includes all necessary providers
export function createWrapper() {
  const testQueryClient = createTestQueryClient()
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={testQueryClient}>
        <AntApp>
          <ConfigProvider
            theme={{
              token: {
                colorPrimary: '#1677ff'
              }
            }}
          >
            <AuthProvider>{children}</AuthProvider>
          </ConfigProvider>
        </AntApp>
      </QueryClientProvider>
    )
  }
}

// Cleanup after each test case (e.g. clearing jsdom)
afterEach(() => {
  cleanup()
  localStorage.clear()
})
