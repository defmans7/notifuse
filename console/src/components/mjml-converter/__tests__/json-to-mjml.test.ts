import { describe, it, expect } from 'vitest'
import {
  convertJsonToMjml,
  convertBlockToMjml,
  formatAttributes,
  formatSingleAttribute,
  camelToKebab,
  escapeAttributeValue,
  escapeContent,
  shouldIncludeAttribute
} from '../json-to-mjml'
import type { EmailBlock } from '../../components/email_builder/types'

describe('JSON to MJML Converter', () => {
  describe('camelToKebab', () => {
    it('should convert camelCase to kebab-case', () => {
      expect(camelToKebab('fontSize')).toBe('font-size')
      expect(camelToKebab('backgroundColor')).toBe('background-color')
      expect(camelToKebab('paddingTop')).toBe('padding-top')
      expect(camelToKebab('borderRadius')).toBe('border-radius')
      expect(camelToKebab('innerPadding')).toBe('inner-padding')
      expect(camelToKebab('textAlign')).toBe('text-align')
    })

    it('should handle already kebab-case strings', () => {
      expect(camelToKebab('font-size')).toBe('font-size')
      expect(camelToKebab('background-color')).toBe('background-color')
    })

    it('should handle single words', () => {
      expect(camelToKebab('color')).toBe('color')
      expect(camelToKebab('width')).toBe('width')
      expect(camelToKebab('height')).toBe('height')
    })

    it('should handle complex camelCase', () => {
      expect(camelToKebab('containerBackgroundColor')).toBe('container-background-color')
      expect(camelToKebab('fluidOnMobile')).toBe('fluid-on-mobile')
    })
  })

  describe('escapeAttributeValue', () => {
    it('should escape HTML entities in non-URL attribute values', () => {
      expect(escapeAttributeValue('Hello & World', 'title')).toBe('Hello &amp; World')
      expect(escapeAttributeValue('Say "Hello"', 'title')).toBe('Say &quot;Hello&quot;')
      expect(escapeAttributeValue("Say 'Hello'", 'title')).toBe('Say &#39;Hello&#39;')
      expect(escapeAttributeValue('<script>', 'title')).toBe('&lt;script&gt;')
    })

    it('should not escape ampersands in URL attributes', () => {
      const imageUrl = 'https://example.com/image.jpg?param1=value1&param2=value2'
      const actionUrl = 'https://example.com/action?test=1&foo=bar'
      const hrefUrl = 'https://example.com/submit?a=1&b=2'

      // Should NOT escape ampersands in URL attributes
      expect(escapeAttributeValue(imageUrl, 'src')).toBe(imageUrl)
      expect(escapeAttributeValue(actionUrl, 'action')).toBe(actionUrl)
      expect(escapeAttributeValue(hrefUrl, 'href')).toBe(hrefUrl)

      // Should still escape other characters in URL attributes
      expect(escapeAttributeValue('https://example.com/test"quotes', 'src')).toBe(
        'https://example.com/test&quot;quotes'
      )
    })

    it('should escape ampersands in non-URL text even in URL attribute names', () => {
      // Non-URL text in src attribute should still have ampersands escaped
      expect(escapeAttributeValue('hello & world', 'src')).toBe('hello &amp; world')
    })

    it('should handle complex strings', () => {
      const input = '<div class="test" onclick="alert(\'hello\')">'
      const expected =
        '&lt;div class=&quot;test&quot; onclick=&quot;alert(&#39;hello&#39;)&quot;&gt;'
      expect(escapeAttributeValue(input, 'title')).toBe(expected)
    })
  })

  describe('escapeContent', () => {
    it('should escape HTML entities in content', () => {
      expect(escapeContent('Hello & World')).toBe('Hello &amp; World')
      expect(escapeContent('<b>Bold</b>')).toBe('&lt;b&gt;Bold&lt;/b&gt;')
      expect(escapeContent('5 > 3')).toBe('5 &gt; 3')
    })
  })

  describe('shouldIncludeAttribute', () => {
    it('should include valid values', () => {
      expect(shouldIncludeAttribute('value')).toBe(true)
      expect(shouldIncludeAttribute(0)).toBe(true)
      expect(shouldIncludeAttribute(false)).toBe(true)
      expect(shouldIncludeAttribute('0')).toBe(true)
    })

    it('should exclude invalid values', () => {
      expect(shouldIncludeAttribute(undefined)).toBe(false)
      expect(shouldIncludeAttribute(null)).toBe(false)
      expect(shouldIncludeAttribute('')).toBe(false)
    })
  })

  describe('formatSingleAttribute', () => {
    it('should format string attributes correctly', () => {
      expect(formatSingleAttribute('fontSize', '16px')).toBe(' font-size="16px"')
      expect(formatSingleAttribute('color', '#ff0000')).toBe(' color="#ff0000"')
      expect(formatSingleAttribute('href', 'https://example.com')).toBe(
        ' href="https://example.com"'
      )
    })

    it('should format boolean attributes correctly', () => {
      expect(formatSingleAttribute('fluidOnMobile', true)).toBe(' fluid-on-mobile')
      expect(formatSingleAttribute('fluidOnMobile', false)).toBe('')
    })

    it('should escape attribute values', () => {
      expect(formatSingleAttribute('alt', 'Image with "quotes"')).toBe(
        ' alt="Image with &quot;quotes&quot;"'
      )
      expect(formatSingleAttribute('title', '<Important>')).toBe(' title="&lt;Important&gt;"')
    })
  })

  describe('formatAttributes', () => {
    it('should format multiple attributes', () => {
      const attributes = {
        fontSize: '16px',
        color: '#333',
        textAlign: 'center'
      }
      const result = formatAttributes(attributes)
      expect(result).toContain(' font-size="16px"')
      expect(result).toContain(' color="#333"')
      expect(result).toContain(' text-align="center"')
    })

    it('should handle empty attributes', () => {
      expect(formatAttributes({})).toBe('')
      expect(formatAttributes(undefined as any)).toBe('')
    })

    it('should filter out invalid values', () => {
      const attributes = {
        fontSize: '16px',
        color: '',
        width: undefined,
        height: null,
        padding: '10px'
      }
      const result = formatAttributes(attributes)
      expect(result).toContain(' font-size="16px"')
      expect(result).toContain(' padding="10px"')
      expect(result).not.toContain('color=')
      expect(result).not.toContain('width=')
      expect(result).not.toContain('height=')
    })

    it('should handle boolean attributes', () => {
      const attributes = {
        fluidOnMobile: true,
        fullWidth: false,
        fontSize: '16px'
      }
      const result = formatAttributes(attributes)
      expect(result).toContain(' fluid-on-mobile')
      expect(result).toContain(' font-size="16px"')
      expect(result).not.toContain('full-width')
    })
  })

  describe('convertBlockToMjml - Basic Blocks', () => {
    it('should convert simple mj-text block', () => {
      const block: EmailBlock = {
        id: 'text-1',
        type: 'mj-text',
        content: 'Hello World',
        attributes: {
          fontSize: '16px',
          color: '#333'
        }
      }

      const result = convertBlockToMjml(block)
      expect(result).toBe('<mj-text font-size="16px" color="#333">Hello World</mj-text>')
    })

    it('should convert mj-button block', () => {
      const block: EmailBlock = {
        id: 'button-1',
        type: 'mj-button',
        content: 'Click Me',
        attributes: {
          backgroundColor: '#007bff',
          href: 'https://example.com',
          target: '_blank'
        }
      }

      const result = convertBlockToMjml(block)
      expect(result).toBe(
        '<mj-button background-color="#007bff" href="https://example.com" target="_blank">Click Me</mj-button>'
      )
    })

    it('should convert self-closing blocks', () => {
      const block: EmailBlock = {
        id: 'image-1',
        type: 'mj-image',
        attributes: {
          src: 'https://example.com/image.jpg',
          width: '200px',
          alt: 'Test Image'
        }
      }

      const result = convertBlockToMjml(block)
      expect(result).toBe(
        '<mj-image src="https://example.com/image.jpg" width="200px" alt="Test Image" />'
      )
    })

    it('should handle blocks without attributes', () => {
      const block: EmailBlock = {
        id: 'text-1',
        type: 'mj-text',
        content: 'Simple text'
      }

      const result = convertBlockToMjml(block)
      expect(result).toBe('<mj-text>Simple text</mj-text>')
    })

    it('should handle self-closing blocks without attributes', () => {
      const block: EmailBlock = {
        id: 'break-1',
        type: 'mj-breakpoint'
      }

      const result = convertBlockToMjml(block)
      expect(result).toBe('<mj-breakpoint />')
    })
  })

  describe('convertBlockToMjml - Complex Structures', () => {
    it('should convert nested structure with proper indentation', () => {
      const block: EmailBlock = {
        id: 'section-1',
        type: 'mj-section',
        attributes: {
          backgroundColor: '#f0f0f0'
        },
        children: [
          {
            id: 'column-1',
            type: 'mj-column',
            attributes: {
              width: '50%'
            },
            children: [
              {
                id: 'text-1',
                type: 'mj-text',
                content: 'Left column',
                attributes: {
                  fontSize: '14px'
                }
              }
            ]
          }
        ]
      }

      const result = convertBlockToMjml(block)
      const expected = `<mj-section background-color="#f0f0f0">
  <mj-column width="50%">
    <mj-text font-size="14px">Left column</mj-text>
  </mj-column>
</mj-section>`

      expect(result).toBe(expected)
    })

    it('should handle multiple children at same level', () => {
      const block: EmailBlock = {
        id: 'section-1',
        type: 'mj-section',
        children: [
          {
            id: 'column-1',
            type: 'mj-column',
            children: [
              {
                id: 'text-1',
                type: 'mj-text',
                content: 'First text'
              },
              {
                id: 'text-2',
                type: 'mj-text',
                content: 'Second text'
              }
            ]
          }
        ]
      }

      const result = convertBlockToMjml(block)
      expect(result).toContain('<mj-text>First text</mj-text>')
      expect(result).toContain('<mj-text>Second text</mj-text>')
    })
  })

  describe('convertJsonToMjml - Full Email Structure', () => {
    it('should convert complete email structure', () => {
      const emailTree: EmailBlock = {
        id: 'mjml-1',
        type: 'mjml',
        children: [
          {
            id: 'body-1',
            type: 'mj-body',
            attributes: {
              backgroundColor: '#ffffff'
            },
            children: [
              {
                id: 'section-1',
                type: 'mj-section',
                attributes: {
                  paddingTop: '20px'
                },
                children: [
                  {
                    id: 'column-1',
                    type: 'mj-column',
                    children: [
                      {
                        id: 'text-1',
                        type: 'mj-text',
                        content: 'Welcome!',
                        attributes: {
                          fontSize: '24px',
                          fontWeight: 'bold',
                          textAlign: 'center'
                        }
                      }
                    ]
                  }
                ]
              }
            ]
          }
        ]
      }

      const result = convertJsonToMjml(emailTree)

      // Check structure
      expect(result).toContain('<mjml>')
      expect(result).toContain('</mjml>')
      expect(result).toContain('<mj-body background-color="#ffffff">')
      expect(result).toContain('<mj-section padding-top="20px">')
      expect(result).toContain(
        '<mj-text font-size="24px" font-weight="bold" text-align="center">Welcome!</mj-text>'
      )

      // Check proper nesting
      const lines = result.split('\n')
      expect(lines[0]).toBe('<mjml>')
      expect(lines[1]).toBe('  <mj-body background-color="#ffffff">')
      expect(lines[2]).toBe('    <mj-section padding-top="20px">')
      expect(lines[3]).toBe('      <mj-column>')
      expect(lines[4]).toBe(
        '        <mj-text font-size="24px" font-weight="bold" text-align="center">Welcome!</mj-text>'
      )
    })
  })

  describe('Attribute Conversion Edge Cases', () => {
    it('should handle special characters in content', () => {
      const block: EmailBlock = {
        id: 'text-1',
        type: 'mj-text',
        content: 'Price: $100 & free shipping!',
        attributes: {
          color: '#333'
        }
      }

      const result = convertBlockToMjml(block)
      // mj-text content should not be escaped (can contain HTML)
      expect(result).toBe('<mj-text color="#333">Price: $100 & free shipping!</mj-text>')
    })

    it('should preserve HTML content in mj-text blocks', () => {
      const block: EmailBlock = {
        id: 'text-html',
        type: 'mj-text',
        content: '<p>Hello <strong>world</strong>!</p><br><em>Welcome</em>',
        attributes: {
          color: '#333'
        }
      }

      const result = convertBlockToMjml(block)
      expect(result).toBe(
        '<mj-text color="#333"><p>Hello <strong>world</strong>!</p><br><em>Welcome</em></mj-text>'
      )
    })

    it('should preserve HTML content in mj-button blocks', () => {
      const block: EmailBlock = {
        id: 'button-html',
        type: 'mj-button',
        content: '<strong>Click</strong> <em>Here</em>',
        attributes: {
          href: 'https://example.com'
        }
      }

      const result = convertBlockToMjml(block)
      expect(result).toBe(
        '<mj-button href="https://example.com"><strong>Click</strong> <em>Here</em></mj-button>'
      )
    })

    it('should still escape HTML content for other block types', () => {
      const block: EmailBlock = {
        id: 'other-block',
        type: 'mj-preview' as any,
        content: '<script>alert("test")</script> & more',
        attributes: {}
      }

      const result = convertBlockToMjml(block)
      expect(result).toBe(
        '<mj-preview>&lt;script&gt;alert("test")&lt;/script&gt; &amp; more</mj-preview>'
      )
    })

    it('should handle quotes in attribute values', () => {
      const block: EmailBlock = {
        id: 'image-1',
        type: 'mj-image',
        attributes: {
          alt: 'Logo of "Company Name"',
          title: "It's awesome!"
        }
      }

      const result = convertBlockToMjml(block)
      expect(result).toContain('alt="Logo of &quot;Company Name&quot;"')
      expect(result).toContain('title="It&#39;s awesome!"')
    })

    it('should handle number attributes', () => {
      const block: EmailBlock = {
        id: 'text-1',
        type: 'mj-text',
        content: 'Text',
        attributes: {
          fontSize: 16,
          lineHeight: 1.5,
          padding: 0
        }
      }

      const result = convertBlockToMjml(block)
      expect(result).toContain('font-size="16"')
      expect(result).toContain('line-height="1.5"')
      expect(result).toContain('padding="0"')
    })

    it('should handle complex camelCase attributes', () => {
      const block: EmailBlock = {
        id: 'section-1',
        type: 'mj-section',
        attributes: {
          backgroundUrl: 'https://example.com/bg.jpg',
          backgroundSize: 'cover',
          backgroundRepeat: 'no-repeat',
          containerBackgroundColor: '#f8f9fa'
        }
      }

      const result = convertBlockToMjml(block)
      expect(result).toContain('background-url="https://example.com/bg.jpg"')
      expect(result).toContain('background-size="cover"')
      expect(result).toContain('background-repeat="no-repeat"')
      expect(result).toContain('container-background-color="#f8f9fa"')
    })
  })

  describe('Real-world MJML Component Tests', () => {
    it('should generate valid mj-text with all common attributes', () => {
      const block: EmailBlock = {
        id: 'text-advanced',
        type: 'mj-text',
        content: 'Advanced styled text',
        attributes: {
          align: 'center',
          backgroundColor: '#f8f9fa',
          borderRadius: '4px',
          border: '1px solid #dee2e6',
          color: '#212529',
          containerBackgroundColor: '#ffffff',
          cssClass: 'custom-text',
          fontFamily: 'Arial, sans-serif',
          fontSize: '16px',
          fontStyle: 'italic',
          fontWeight: 'bold',
          height: '100px',
          letterSpacing: '1px',
          lineHeight: '1.6',
          paddingBottom: '10px',
          paddingLeft: '15px',
          paddingRight: '15px',
          paddingTop: '10px',
          textDecoration: 'underline',
          textTransform: 'uppercase',
          verticalAlign: 'middle',
          width: '300px'
        }
      }

      const result = convertBlockToMjml(block)

      // Verify all attributes are converted correctly
      expect(result).toContain('align="center"')
      expect(result).toContain('background-color="#f8f9fa"')
      expect(result).toContain('border-radius="4px"')
      expect(result).toContain('border="1px solid #dee2e6"')
      expect(result).toContain('color="#212529"')
      expect(result).toContain('container-background-color="#ffffff"')
      expect(result).toContain('css-class="custom-text"')
      expect(result).toContain('font-family="Arial, sans-serif"')
      expect(result).toContain('font-size="16px"')
      expect(result).toContain('font-style="italic"')
      expect(result).toContain('font-weight="bold"')
      expect(result).toContain('height="100px"')
      expect(result).toContain('letter-spacing="1px"')
      expect(result).toContain('line-height="1.6"')
      expect(result).toContain('padding-bottom="10px"')
      expect(result).toContain('padding-left="15px"')
      expect(result).toContain('padding-right="15px"')
      expect(result).toContain('padding-top="10px"')
      expect(result).toContain('text-decoration="underline"')
      expect(result).toContain('text-transform="uppercase"')
      expect(result).toContain('vertical-align="middle"')
      expect(result).toContain('width="300px"')
      expect(result).toContain('>Advanced styled text</mj-text>')
    })

    it('should generate valid mj-button with all attributes', () => {
      const block: EmailBlock = {
        id: 'button-advanced',
        type: 'mj-button',
        content: 'Advanced Button',
        attributes: {
          align: 'center',
          backgroundColor: '#007bff',
          borderRadius: '25px',
          border: '2px solid #0056b3',
          color: '#ffffff',
          containerBackgroundColor: '#f8f9fa',
          cssClass: 'btn-custom',
          fontFamily: 'Helvetica, Arial',
          fontSize: '18px',
          fontStyle: 'normal',
          fontWeight: '600',
          height: '50px',
          href: 'https://example.com/action',
          innerPadding: '12px 30px',
          lineHeight: '120%',
          paddingBottom: '15px',
          paddingLeft: '20px',
          paddingRight: '20px',
          paddingTop: '15px',
          rel: 'noopener',
          target: '_blank',
          textDecoration: 'none',
          textTransform: 'capitalize',
          verticalAlign: 'middle',
          width: '200px'
        }
      }

      const result = convertBlockToMjml(block)

      // Check critical button attributes
      expect(result).toContain('background-color="#007bff"')
      expect(result).toContain('href="https://example.com/action"')
      expect(result).toContain('target="_blank"')
      expect(result).toContain('inner-padding="12px 30px"')
      expect(result).toContain('border-radius="25px"')
      expect(result).toContain('>Advanced Button</mj-button>')
    })
  })
})
