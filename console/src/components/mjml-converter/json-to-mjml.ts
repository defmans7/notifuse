import type { EmailBlock } from '../email_builder/types'

/**
 * Converts an EmailBlock JSON tree to MJML string
 */
export function convertJsonToMjml(tree: EmailBlock): string {
  return convertBlockToMjml(tree)
}

/**
 * Recursively converts a single EmailBlock to MJML string
 */
export function convertBlockToMjml(block: EmailBlock, indentLevel: number = 0): string {
  const indent = '  '.repeat(indentLevel)
  const tagName = block.type

  // Handle self-closing tags that don't have children but may have content
  if (!block.children || block.children.length === 0) {
    // Check if the block has content (for mj-text, mj-button, etc.)
    const content = (block as any).content || ''

    if (content) {
      // Block with content - don't escape for mj-raw, mj-text, and mj-button (they can contain HTML)
      const attributeString = formatAttributes(block.attributes || {})
      if (block.type === 'mj-raw' || block.type === 'mj-text' || block.type === 'mj-button') {
        return `${indent}<${tagName}${attributeString}>${content}</${tagName}>`
      } else {
        return `${indent}<${tagName}${attributeString}>${escapeContent(content)}</${tagName}>`
      }
    } else {
      // Self-closing block or empty block
      const attributeString = formatAttributes(block.attributes || {})
      if (attributeString) {
        return `${indent}<${tagName}${attributeString} />`
      } else {
        return `${indent}<${tagName} />`
      }
    }
  }

  // Block with children
  const attributeString = formatAttributes(block.attributes || {})
  const openTag = `${indent}<${tagName}${attributeString}>`
  const closeTag = `${indent}</${tagName}>`

  // Process children
  const childrenMjml = block.children
    .map((child) => convertBlockToMjml(child, indentLevel + 1))
    .join('\n')

  return `${openTag}\n${childrenMjml}\n${closeTag}`
}

/**
 * Formats attributes object into MJML attribute string
 */
export function formatAttributes(attributes: Record<string, any>): string {
  if (!attributes || Object.keys(attributes).length === 0) {
    return ''
  }

  const attrPairs = Object.entries(attributes)
    .filter(([_, value]) => shouldIncludeAttribute(value))
    .map(([key, value]) => formatSingleAttribute(key, value))
    .filter((attr) => attr !== '')

  return attrPairs.join('')
}

/**
 * Determines if an attribute value should be included in the output
 */
export function shouldIncludeAttribute(value: any): boolean {
  return value !== undefined && value !== null && value !== ''
}

/**
 * Formats a single attribute key-value pair
 */
export function formatSingleAttribute(key: string, value: any): string {
  // Convert camelCase to kebab-case for MJML attributes
  const kebabKey = camelToKebab(key)

  // Handle boolean attributes
  if (typeof value === 'boolean') {
    return value ? ` ${kebabKey}` : ''
  }

  // Handle string/number attributes
  const escapedValue = escapeAttributeValue(String(value), kebabKey)
  return ` ${kebabKey}="${escapedValue}"`
}

/**
 * Converts camelCase to kebab-case
 */
export function camelToKebab(str: string): string {
  return str.replace(/[A-Z]/g, (letter) => `-${letter.toLowerCase()}`)
}

/**
 * Escapes attribute values for safe HTML output
 * For URL attributes (src, href, action), we don't escape & to preserve URL query parameters
 */
export function escapeAttributeValue(value: string, attributeName?: string): string {
  // Check if this is a URL attribute and the value looks like a URL
  const isURLAttribute =
    attributeName === 'src' || attributeName === 'href' || attributeName === 'action'
  const looksLikeURL =
    value.startsWith('http://') || value.startsWith('https://') || value.startsWith('//')

  let result = value
  // Only skip escaping ampersands if it's a URL attribute AND the value looks like a URL
  if (!(isURLAttribute && looksLikeURL)) {
    result = result.replace(/&/g, '&amp;')
  }
  return result
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
}

/**
 * Escapes content for safe HTML output
 */
export function escapeContent(content: string): string {
  return content.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}
