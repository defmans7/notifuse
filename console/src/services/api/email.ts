import { api } from './client'
import {
  EmailProvider,
  TestEmailProviderResponse,
  TestTemplateRequest,
  TestTemplateResponse
} from './types'

export const emailService = {
  /**
   * Test an email provider configuration by sending a test email
   * @param workspaceId The ID of the workspace
   * @param provider The email provider configuration to test
   * @param to The recipient email address for the test
   * @returns A response indicating success or failure
   */
  testProvider: (
    workspaceId: string,
    provider: EmailProvider,
    to: string
  ): Promise<TestEmailProviderResponse> => {
    return api.post<TestEmailProviderResponse>('/api/email.testProvider', {
      provider,
      to,
      workspace_id: workspaceId
    })
  },

  /**
   * Test a template by sending a test email
   * @param workspaceId The ID of the workspace
   * @param templateId The ID of the template to test
   * @param integrationId The ID of the integration to use
   * @param recipientEmail The email address to send the test email to
   * @param cc Optional array of CC email addresses
   * @param bcc Optional array of BCC email addresses
   * @param replyTo Optional Reply-To email address
   * @returns A response indicating success or failure
   */
  testTemplate: (
    workspaceId: string,
    templateId: string,
    integrationId: string,
    recipientEmail: string,
    cc?: string[],
    bcc?: string[],
    replyTo?: string
  ): Promise<TestTemplateResponse> => {
    const request: TestTemplateRequest = {
      workspace_id: workspaceId,
      template_id: templateId,
      integration_id: integrationId,
      recipient_email: recipientEmail,
      cc,
      bcc,
      reply_to: replyTo
    }
    return api.post<TestTemplateResponse>('/api/email.testTemplate', request)
  }
}
