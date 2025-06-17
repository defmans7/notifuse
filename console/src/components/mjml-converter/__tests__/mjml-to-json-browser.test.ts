import { describe, test, expect } from 'vitest'
import { convertMjmlToJsonBrowser } from '../mjml-to-json-browser'

describe('MJML to JSON Browser Converter', () => {
  describe('Basic Conversion', () => {
    test('should convert simple MJML to EmailBlock format', () => {
      const mjmlInput = `
        <mjml>
          <mj-body>
            <mj-section>
              <mj-column>
                <mj-text>Hello World</mj-text>
              </mj-column>
            </mj-section>
          </mj-body>
        </mjml>
      `

      const result = convertMjmlToJsonBrowser(mjmlInput)

      expect(result.type).toBe('mjml')
      expect(result.id).toBeDefined()
      expect(result.children).toBeDefined()
      expect(result.children?.length).toBe(1)

      const bodyBlock = result.children?.[0]
      expect(bodyBlock?.type).toBe('mj-body')
      expect(bodyBlock?.children?.length).toBe(1)

      const sectionBlock = bodyBlock?.children?.[0]
      expect(sectionBlock?.type).toBe('mj-section')
      expect(sectionBlock?.children?.length).toBe(1)

      const columnBlock = sectionBlock?.children?.[0]
      expect(columnBlock?.type).toBe('mj-column')
      expect(columnBlock?.children?.length).toBe(1)

      const textBlock = columnBlock?.children?.[0]
      expect(textBlock?.type).toBe('mj-text')
      expect((textBlock as any)?.content).toBe('Hello World')
    })

    test('should handle MJML with attributes', () => {
      const mjmlInput = `
        <mjml>
          <mj-body width="600px" background-color="#ffffff">
            <mj-section padding="20px">
              <mj-column>
                <mj-text font-size="16px" color="#333333">Styled Text</mj-text>
              </mj-column>
            </mj-section>
          </mj-body>
        </mjml>
      `

      const result = convertMjmlToJsonBrowser(mjmlInput)

      const bodyBlock = result.children?.[0]
      expect((bodyBlock?.attributes as any)?.width).toBe('600px')
      expect((bodyBlock?.attributes as any)?.backgroundColor).toBe('#ffffff')

      const sectionBlock = bodyBlock?.children?.[0]
      expect((sectionBlock?.attributes as any)?.padding).toBe('20px')

      const textBlock = sectionBlock?.children?.[0]?.children?.[0]
      expect((textBlock?.attributes as any)?.fontSize).toBe('16px')
      expect((textBlock?.attributes as any)?.color).toBe('#333333')
      expect((textBlock as any)?.content).toBe('Styled Text')
    })

    test('should handle self-closing elements', () => {
      const mjmlInput = `
        <mjml>
          <mj-body>
            <mj-section>
              <mj-column>
                <mj-spacer height="20px" />
                <mj-divider border-width="1px" border-color="#ccc" />
              </mj-column>
            </mj-section>
          </mj-body>
        </mjml>
      `

      const result = convertMjmlToJsonBrowser(mjmlInput)

      const columnBlock = result.children?.[0]?.children?.[0]?.children?.[0]
      expect(columnBlock?.children?.length).toBe(2)

      const spacerBlock = columnBlock?.children?.[0]
      expect(spacerBlock?.type).toBe('mj-spacer')
      expect((spacerBlock?.attributes as any)?.height).toBe('20px')
      expect(spacerBlock?.children).toBeUndefined()

      const dividerBlock = columnBlock?.children?.[1]
      expect(dividerBlock?.type).toBe('mj-divider')
      expect((dividerBlock?.attributes as any)?.borderWidth).toBe('1px')
      expect((dividerBlock?.attributes as any)?.borderColor).toBe('#ccc')
    })
  })

  describe('Complex Structures', () => {
    test('should handle multiple sections and columns', () => {
      const mjmlInput = `
        <mjml>
          <mj-body>
            <mj-section>
              <mj-column width="50%">
                <mj-text>Left Column</mj-text>
              </mj-column>
              <mj-column width="50%">
                <mj-text>Right Column</mj-text>
              </mj-column>
            </mj-section>
            <mj-section>
              <mj-column>
                <mj-button href="https://example.com">Click Me</mj-button>
              </mj-column>
            </mj-section>
          </mj-body>
        </mjml>
      `

      const result = convertMjmlToJsonBrowser(mjmlInput)

      const bodyBlock = result.children?.[0]
      expect(bodyBlock?.children?.length).toBe(2)

      // First section with two columns
      const firstSection = bodyBlock?.children?.[0]
      expect(firstSection?.type).toBe('mj-section')
      expect(firstSection?.children?.length).toBe(2)

      const leftColumn = firstSection?.children?.[0]
      const rightColumn = firstSection?.children?.[1]
      expect((leftColumn?.attributes as any)?.width).toBe('50%')
      expect((rightColumn?.attributes as any)?.width).toBe('50%')
      expect((leftColumn?.children?.[0] as any)?.content).toBe('Left Column')
      expect((rightColumn?.children?.[0] as any)?.content).toBe('Right Column')

      // Second section with button
      const secondSection = bodyBlock?.children?.[1]
      const buttonBlock = secondSection?.children?.[0]?.children?.[0]
      expect(buttonBlock?.type).toBe('mj-button')
      expect((buttonBlock?.attributes as any)?.href).toBe('https://example.com')
      expect((buttonBlock as any)?.content).toBe('Click Me')
    })

    test('should handle MJML head section', () => {
      const mjmlInput = `
        <mjml>
          <mj-head>
            <mj-title>Test Email</mj-title>
            <mj-preview>This is a preview</mj-preview>
          </mj-head>
          <mj-body>
            <mj-section>
              <mj-column>
                <mj-text>Body content</mj-text>
              </mj-column>
            </mj-section>
          </mj-body>
        </mjml>
      `

      const result = convertMjmlToJsonBrowser(mjmlInput)

      expect(result.children?.length).toBe(2)

      const headBlock = result.children?.[0]
      const bodyBlock = result.children?.[1]

      expect(headBlock?.type).toBe('mj-head')
      expect(headBlock?.children?.length).toBe(2)

      const titleBlock = headBlock?.children?.[0]
      const previewBlock = headBlock?.children?.[1]

      expect(titleBlock?.type).toBe('mj-title')
      expect((titleBlock as any)?.content).toBe('Test Email')

      expect(previewBlock?.type).toBe('mj-preview')
      expect((previewBlock as any)?.content).toBe('This is a preview')

      expect(bodyBlock?.type).toBe('mj-body')
    })
  })

  describe('Error Handling', () => {
    test('should throw error for invalid XML', () => {
      const invalidMjml = `
        <mjml>
          <mj-body>
            <mj-section>
              <mj-column>
                <mj-text>Unclosed tag
              </mj-column>
            </mj-section>
          </mj-body>
      `

      expect(() => convertMjmlToJsonBrowser(invalidMjml)).toThrow('Invalid MJML syntax')
    })

    test('should throw error for non-mjml root element', () => {
      const invalidMjml = `
        <html>
          <body>
            <p>This is not MJML</p>
          </body>
        </html>
      `

      expect(() => convertMjmlToJsonBrowser(invalidMjml)).toThrow('Root element must be <mjml>')
    })

    test('should handle empty MJML', () => {
      const emptyMjml = '<mjml></mjml>'

      const result = convertMjmlToJsonBrowser(emptyMjml)
      expect(result.type).toBe('mjml')
      expect(result.children).toBeUndefined()
    })
  })

  describe('Content and Attribute Parsing', () => {
    test('should handle mixed content and whitespace', () => {
      const mjmlInput = `
        <mjml>
          <mj-body>
            <mj-section>
              <mj-column>
                <mj-text>
                  Hello World with whitespace
                </mj-text>
              </mj-column>
            </mj-section>
          </mj-body>
        </mjml>
      `

      const result = convertMjmlToJsonBrowser(mjmlInput)
      const textBlock = result.children?.[0]?.children?.[0]?.children?.[0]?.children?.[0]
      expect((textBlock as any)?.content).toBe('Hello World with whitespace')
    })

    test('should handle boolean-style attributes', () => {
      const mjmlInput = `
        <mjml>
          <mj-body>
            <mj-section>
              <mj-column>
                <mj-image src="test.jpg" fluid-on-mobile="true" />
              </mj-column>
            </mj-section>
          </mj-body>
        </mjml>
      `

      const result = convertMjmlToJsonBrowser(mjmlInput)
      const imageBlock = result.children?.[0]?.children?.[0]?.children?.[0]?.children?.[0]
      expect(imageBlock?.type).toBe('mj-image')
      expect((imageBlock?.attributes as any)?.src).toBe('test.jpg')
      expect((imageBlock?.attributes as any)?.fluidOnMobile).toBe('true')
    })

    test('should generate unique IDs for each block', () => {
      const mjmlInput = `
        <mjml>
          <mj-body>
            <mj-section>
              <mj-column>
                <mj-text>First</mj-text>
                <mj-text>Second</mj-text>
              </mj-column>
            </mj-section>
          </mj-body>
        </mjml>
      `

      const result = convertMjmlToJsonBrowser(mjmlInput)
      const columnBlock = result.children?.[0]?.children?.[0]?.children?.[0]
      const firstText = columnBlock?.children?.[0]
      const secondText = columnBlock?.children?.[1]

      expect(firstText?.id).toBeDefined()
      expect(secondText?.id).toBeDefined()
      expect(firstText?.id).not.toBe(secondText?.id)
      expect(result.id).not.toBe(firstText?.id)
    })

    test('should convert kebab-case attributes to camelCase', () => {
      const mjmlInput = `
        <mjml>
          <mj-body>
            <mj-section>
              <mj-column>
                <mj-divider border-width="2px" border-color="#ff0000" border-style="solid" />
                <mj-text font-size="18px" background-color="#f0f0f0" line-height="1.5">Test</mj-text>
              </mj-column>
            </mj-section>
          </mj-body>
        </mjml>
      `

      const result = convertMjmlToJsonBrowser(mjmlInput)
      const columnBlock = result.children?.[0]?.children?.[0]?.children?.[0]

      const dividerBlock = columnBlock?.children?.[0]
      expect(dividerBlock?.type).toBe('mj-divider')
      expect((dividerBlock?.attributes as any)?.borderWidth).toBe('2px')
      expect((dividerBlock?.attributes as any)?.borderColor).toBe('#ff0000')
      expect((dividerBlock?.attributes as any)?.borderStyle).toBe('solid')

      const textBlock = columnBlock?.children?.[1]
      expect(textBlock?.type).toBe('mj-text')
      expect((textBlock?.attributes as any)?.fontSize).toBe('18px')
      expect((textBlock?.attributes as any)?.backgroundColor).toBe('#f0f0f0')
      expect((textBlock?.attributes as any)?.lineHeight).toBe('1.5')
    })

    test('should convert SVG-specific attributes like stroke-width to camelCase', () => {
      const mjmlInput = `
        <mjml>
          <mj-body>
            <mj-section>
              <mj-column>
                <mj-divider stroke-width="3px" stroke-color="#0000ff" stroke-dasharray="5,5" />
                <mj-image stroke-width="1px" src="test.jpg" fluid-on-mobile="true" />
              </mj-column>
            </mj-section>
          </mj-body>
        </mjml>
      `

      const result = convertMjmlToJsonBrowser(mjmlInput)
      const columnBlock = result.children?.[0]?.children?.[0]?.children?.[0]

      const dividerBlock = columnBlock?.children?.[0]
      expect(dividerBlock?.type).toBe('mj-divider')
      expect((dividerBlock?.attributes as any)?.strokeWidth).toBe('3px')
      expect((dividerBlock?.attributes as any)?.strokeColor).toBe('#0000ff')
      expect((dividerBlock?.attributes as any)?.strokeDasharray).toBe('5,5')

      const imageBlock = columnBlock?.children?.[1]
      expect(imageBlock?.type).toBe('mj-image')
      expect((imageBlock?.attributes as any)?.strokeWidth).toBe('1px')
      expect((imageBlock?.attributes as any)?.src).toBe('test.jpg')
      expect((imageBlock?.attributes as any)?.fluidOnMobile).toBe('true')

      // Verify no kebab-case attributes remain
      const allAttributes = [
        ...Object.keys((dividerBlock?.attributes as any) || {}),
        ...Object.keys((imageBlock?.attributes as any) || {})
      ]
      const kebabAttributes = allAttributes.filter((attr) => attr.includes('-'))
      expect(kebabAttributes).toEqual([])
    })

    test('should handle mj-raw content as HTML string, not child elements', () => {
      const mjmlInput = `
        <mjml>
          <mj-body>
            <mj-raw>
              <style type="text/css">
                .custom { color: red; }
              </style>
              <div style="background: blue;">
                <p>Custom HTML content</p>
                <span>With nested elements</span>
              </div>
            </mj-raw>
            <mj-section>
              <mj-column>
                <mj-text>Regular text</mj-text>
              </mj-column>
            </mj-section>
          </mj-body>
        </mjml>
      `

      const result = convertMjmlToJsonBrowser(mjmlInput)
      const bodyBlock = result.children?.[0]

      // Should have mj-raw and mj-section as children
      expect(bodyBlock?.children?.length).toBe(2)

      const rawBlock = bodyBlock?.children?.[0]
      const sectionBlock = bodyBlock?.children?.[1]

      expect(rawBlock?.type).toBe('mj-raw')
      expect(sectionBlock?.type).toBe('mj-section')

      // mj-raw should have NO children - content should be stored as string
      expect(rawBlock?.children).toBeUndefined()

      // mj-raw should have content as HTML string
      const rawContent = (rawBlock as any)?.content
      expect(rawContent).toBeDefined()
      expect(rawContent).toContain('<style type="text/css">')
      expect(rawContent).toContain('.custom { color: red; }')
      expect(rawContent).toContain('<div style="background: blue;">')
      expect(rawContent).toContain('<p>Custom HTML content</p>')
      expect(rawContent).toContain('<span>With nested elements</span>')

      // Regular section should still have children
      expect(sectionBlock?.children?.length).toBe(1)
      expect(sectionBlock?.children?.[0]?.type).toBe('mj-column')
    })

    test('should handle mj-raw blocks in different contexts (wrapper, section, column)', () => {
      const mjmlInput = `
        <mjml>
          <mj-body>
            <mj-wrapper>
              <mj-raw>
                <style>
                  .custom-wrapper { background: red; }
                </style>
              </mj-raw>
              <mj-section>
                <mj-raw>
                  <div class="section-raw">Section level raw HTML</div>
                </mj-raw>
                <mj-column>
                  <mj-text>Regular text</mj-text>
                  <mj-raw>
                    <p>Column level raw HTML</p>
                  </mj-raw>
                </mj-column>
              </mj-section>
            </mj-wrapper>
          </mj-body>
        </mjml>
      `

      const result = convertMjmlToJsonBrowser(mjmlInput)

      // Verify the structure
      const bodyBlock = result.children?.[0]
      expect(bodyBlock?.type).toBe('mj-body')

      const wrapperBlock = bodyBlock?.children?.[0]
      expect(wrapperBlock?.type).toBe('mj-wrapper')
      expect(wrapperBlock?.children).toHaveLength(2) // raw + section

      // Wrapper level raw
      const wrapperRawBlock = wrapperBlock?.children?.[0]
      expect(wrapperRawBlock?.type).toBe('mj-raw')
      expect((wrapperRawBlock as any)?.content).toContain('.custom-wrapper')

      // Section block
      const sectionBlock = wrapperBlock?.children?.[1]
      expect(sectionBlock?.type).toBe('mj-section')
      expect(sectionBlock?.children).toHaveLength(2) // raw + column

      // Section level raw
      const sectionRawBlock = sectionBlock?.children?.[0]
      expect(sectionRawBlock?.type).toBe('mj-raw')
      expect((sectionRawBlock as any)?.content).toContain('Section level raw HTML')

      // Column block
      const columnBlock = sectionBlock?.children?.[1]
      expect(columnBlock?.type).toBe('mj-column')
      expect(columnBlock?.children).toHaveLength(2) // text + raw

      // Column level raw
      const columnRawBlock = columnBlock?.children?.[1]
      expect(columnRawBlock?.type).toBe('mj-raw')
      expect((columnRawBlock as any)?.content).toContain('Column level raw HTML')
    })
  })
})
