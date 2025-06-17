import type { EmailBlock } from '../email_builder/types'

/**
 * Browser-compatible MJML to JSON converter using DOMParser
 * This is a fallback when mjml2json doesn't work in browser environment
 */
export function convertMjmlToJsonBrowser(mjmlString: string): EmailBlock {
  try {
    // Parse MJML using browser's DOMParser
    const parser = new DOMParser()
    const doc = parser.parseFromString(mjmlString, 'text/xml')

    // Check for parsing errors
    const parserError = doc.querySelector('parsererror')
    if (parserError) {
      throw new Error('Invalid MJML syntax: ' + parserError.textContent)
    }

    // Find the root element (should be mjml)
    const rootElement = doc.documentElement
    if (rootElement.tagName.toLowerCase() !== 'mjml') {
      throw new Error('Root element must be <mjml>')
    }

    // Convert DOM node to EmailBlock format
    return convertDomNodeToEmailBlock(rootElement)
  } catch (error) {
    console.error('Browser MJML to JSON conversion error:', error)
    throw new Error(`Failed to convert MJML to JSON: ${error}`)
  }
}

/**
 * Convert kebab-case to camelCase for React compatibility
 * More comprehensive version that handles all cases
 */
function kebabToCamelCase(str: string): string {
  // Handle special cases first
  if (!str.includes('-')) {
    return str
  }

  // Convert kebab-case to camelCase
  return str.replace(/-([a-zA-Z])/g, (_, letter) => letter.toUpperCase())
}

/**
 * Recursively converts a DOM element to EmailBlock format
 */
function convertDomNodeToEmailBlock(element: Element): EmailBlock {
  // Generate a unique ID for each block
  const generateId = () => Math.random().toString(36).substr(2, 9)

  const block: EmailBlock = {
    id: generateId(),
    type: element.tagName.toLowerCase() as any,
    attributes: {}
  }

  // Extract attributes
  if (element.attributes.length > 0) {
    const attributes: Record<string, any> = {}
    for (let i = 0; i < element.attributes.length; i++) {
      const attr = element.attributes[i]
      // Convert kebab-case to camelCase for React compatibility
      const attributeName = kebabToCamelCase(attr.name)
      attributes[attributeName] = attr.value
    }
    block.attributes = attributes
  }

  // Special handling for mj-raw - store inner HTML as content, don't parse children
  if (element.tagName.toLowerCase() === 'mj-raw') {
    const innerHTML = element.innerHTML
    if (innerHTML.trim()) {
      ;(block as any).content = innerHTML.trim()
    }
    return block
  }

  // Handle content and children for other elements
  const children: EmailBlock[] = []
  let textContent = ''

  for (let i = 0; i < element.childNodes.length; i++) {
    const child = element.childNodes[i]

    if (child.nodeType === Node.ELEMENT_NODE) {
      // It's an element, recursively convert it
      children.push(convertDomNodeToEmailBlock(child as Element))
    } else if (child.nodeType === Node.TEXT_NODE) {
      // It's text content
      const text = child.textContent?.trim()
      if (text) {
        textContent += text
      }
    }
  }

  // If there are child elements, add them
  if (children.length > 0) {
    ;(block as any).children = children
  }

  // If there's text content but no child elements, add it as content
  if (textContent && children.length === 0) {
    ;(block as any).content = textContent
  }

  return block
}
