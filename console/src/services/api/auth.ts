import { api } from './client'
import { SignInRequest, VerifyCodeRequest, VerifyResponse } from './types'

export const authService = {
  signIn: (data: SignInRequest) => api.post<{ message: string }>('/api/user.signin', data),

  verifyCode: (data: VerifyCodeRequest) => api.post<VerifyResponse>('/api/user.verify', data)
}
