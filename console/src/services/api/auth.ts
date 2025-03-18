import { api } from './client'
import {
  SignInRequest,
  SignInResponse,
  VerifyCodeRequest,
  VerifyResponse,
  GetCurrentUserResponse
} from './types'

export const authService = {
  signIn: (data: SignInRequest) => api.post<SignInResponse>('/api/user.signin', data),
  verifyCode: (data: VerifyCodeRequest) => api.post<VerifyResponse>('/api/user.verify', data),
  getCurrentUser: () => api.get<GetCurrentUserResponse>('/api/user.me')
}
