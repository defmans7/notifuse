import React from 'react'
import type { MJMLComponentType, EmailBlock } from '../types'
import {
  BaseEmailBlock,
  type OnUpdateAttributesFunction,
  type PreviewProps
} from './BaseEmailBlock'
import { MJML_COMPONENT_DEFAULTS } from '../mjml-defaults'
import PanelLayout from '../panels/PanelLayout'

/**
 * Implementation for mj-attributes blocks
 */
export class MjAttributesBlock extends BaseEmailBlock {
  getIcon(): React.ReactNode {
    return null
  }

  getLabel(): string {
    return 'Default attributes'
  }

  getDescription(): React.ReactNode {
    return 'Defines default attribute values for MJML components'
  }

  getCategory(): 'content' | 'layout' {
    return 'layout'
  }

  getDefaults(): Record<string, any> {
    return MJML_COMPONENT_DEFAULTS['mj-attributes'] || {}
  }

  canHaveChildren(): boolean {
    return true
  }

  getValidChildTypes(): MJMLComponentType[] {
    // mj-attributes can contain attribute elements for any MJML component type
    return ['mj-text', 'mj-button', 'mj-image', 'mj-section', 'mj-column', 'mj-wrapper', 'mj-body']
  }

  /**
   * Render the settings panel for the attributes block
   */
  renderSettingsPanel(
    _onUpdate: OnUpdateAttributesFunction,
    _blockDefaults: Record<string, any>,
    _emailTree?: EmailBlock
  ): React.ReactNode {
    return (
      <PanelLayout title="Default Attributes">
        <div className="text-sm text-gray-500 text-center py-8">
          No settings available for the attributes container.
          <br />
          Add child elements for specific components (mj-text, mj-button, etc.) to set their default
          values.
        </div>
      </PanelLayout>
    )
  }

  getEdit(_props: PreviewProps): React.ReactNode {
    // Attributes blocks don't render in preview (they contain configuration)
    return null
  }
}
