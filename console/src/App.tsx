import { ConfigProvider, App as AntApp } from 'antd'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { RouterProvider } from '@tanstack/react-router'
import { router } from './router'
import { AuthProvider } from './contexts/AuthContext'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1
    }
  }
})

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <AntApp>
          <ConfigProvider
            theme={{
              token: {
                colorPrimary: '#1677ff'
              }
            }}
          >
            <RouterProvider router={router} />
          </ConfigProvider>
        </AntApp>
      </AuthProvider>
    </QueryClientProvider>
  )
}

export default App
