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
 * Implementation for mj-head blocks
 */
export class MjHeadBlock extends BaseEmailBlock {
  getIcon(): React.ReactNode {
    return null
  }

  getLabel(): string {
    return 'Head'
  }

  getDescription(): React.ReactNode {
    return 'Contains metadata and configuration for the email'
  }

  getCategory(): 'content' | 'layout' {
    return 'layout'
  }

  getDefaults(): Record<string, any> {
    return MJML_COMPONENT_DEFAULTS['mj-head'] || {}
  }

  canHaveChildren(): boolean {
    return true
  }

  getValidChildTypes(): MJMLComponentType[] {
    return [
      'mj-attributes',
      'mj-breakpoint',
      'mj-font',
      'mj-html-attributes',
      'mj-preview',
      'mj-style',
      'mj-title'
    ]
  }

  /**
   * Render the settings panel for the head block
   */
  renderSettingsPanel(
    _onUpdate: OnUpdateAttributesFunction,
    _blockDefaults: Record<string, any>,
    _emailTree?: EmailBlock
  ): React.ReactNode {
    return (
      <PanelLayout title="Head Attributes">
        <div className="text-sm text-gray-500 text-center py-8">
          No settings available for the head element.
          <br />
          Add child elements like mj-font, mj-style, or mj-preview to configure email metadata.
        </div>
      </PanelLayout>
    )
  }

  getEdit(_props: PreviewProps): React.ReactNode {
    // Head blocks don't render in preview (they contain metadata)
    return null
  }
}
