import { kebabCase } from 'lodash'
import { BlockInterface } from './Block'
import mjml2html from 'mjml-browser'
import { Liquid } from 'liquidjs'

const indentPad = (n: number) => Array(n + 1).join(' ')

const TAG_CONVERSION: any = {
  'mj-dev': 'mj-raw'
}

const lineAttributes = (attrs: any) =>
  Object.keys(attrs)
    .filter(
      (key) =>
        key !== 'passport' && attrs[key] !== undefined && attrs[key] !== null && attrs[key] !== ''
    ) // Filter out undefined/null/empty attrs
    .map((key) => `${key}="${attrs[key]}"`)
    .sort()
    .join(' ')

const objectAsKebab = (obj: any) => {
  const newObj: any = {}
  // console.log('obj', obj)
  Object.keys(obj).forEach((key: string) => {
    newObj[kebabCase(key)] = obj[key]
  })
  return newObj
}

const trackURL = (url: string, urlParams: any) => {
  // Ignore if URL is empty, a placeholder, or already tracked (basic check)
  if (!url || url.includes('{{') || url.includes('{%') || url.includes('utm_source=')) {
    return url
  }

  try {
    // Check if it's a mailto or tel link, which shouldn't be tracked
    if (url.startsWith('mailto:') || url.startsWith('tel:')) {
      return url
    }

    const newURL = new URL(url) // This might fail for relative URLs, handle gracefully
    if (!newURL.searchParams.has('utm_source') && urlParams?.utm_source) {
      newURL.searchParams.append('utm_source', urlParams.utm_source)
    }
    if (!newURL.searchParams.has('utm_medium') && urlParams?.utm_medium) {
      newURL.searchParams.append('utm_medium', urlParams.utm_medium)
    }
    if (!newURL.searchParams.has('utm_campaign') && urlParams?.utm_campaign) {
      newURL.searchParams.append('utm_campaign', urlParams.utm_campaign)
    }
    if (!newURL.searchParams.has('utm_content') && urlParams?.utm_content) {
      newURL.searchParams.append('utm_content', urlParams.utm_content)
    }
    if (!newURL.searchParams.has('utm_id') && urlParams?.utm_id) {
      newURL.searchParams.append('utm_id', urlParams.utm_id)
    }
    return newURL.toString()
  } catch (e) {
    // It might be a relative URL or invalid. Log warning and return original.
    console.warn('Could not parse URL for tracking, returning original:', url, e)
    return url
  }
}

export const treeToMjml = (
  rootStyles: any,
  block: BlockInterface,
  templateData: string,
  urlParams: any,
  indent = 0,
  parent?: BlockInterface
): string => {
  // Handle null or undefined block gracefully
  if (!block) {
    return ''
  }
  const space = indentPad(indent)
  let tagName = ''
  let attributes: any = {}
  let content = ''
  let children: BlockInterface[] = block.children || []
  let childrenMjml = ''

  // Ensure block.data exists before accessing its properties
  const blockData = block.data || {}
  const blockStyles = blockData.styles || {}

  // Determine tagName, attributes, content based on block.kind
  switch (block.kind) {
    case 'root':
      tagName = 'mjml'
      // Ensure rootStyles.body exists
      const bodyStyles = blockData.styles?.body || {}
      const bodyAttrsObj = objectAsKebab(bodyStyles)
      delete bodyAttrsObj['margin'] // Remove default margin potentially added
      const bodyAttrsStr = lineAttributes(bodyAttrsObj)
      const bodySpace = indentPad(indent + 2)

      const bodyChildren = children
        .map((child) => treeToMjml(rootStyles, child, templateData, urlParams, indent + 4, block))
        .filter((s) => s && s.trim() !== '')
        .join('\n')

      // Root MJML structure
      return `${space}<mjml>\n${bodySpace}<mj-body${bodyAttrsStr ? ' ' + bodyAttrsStr : ''}>\n${bodyChildren}\n${bodySpace}</mj-body>\n${space}</mjml>`

    case 'columns168':
    case 'columns204':
    case 'columns420':
    case 'columns816':
    case 'columns888':
    case 'columns1212':
    case 'columns6666':
    case 'oneColumn':
      tagName = 'mj-section'
      const sectionAttrs: any = {
        'text-align': blockStyles.textAlign
      }

      if (blockData.backgroundType === 'image') {
        sectionAttrs['background-url'] = blockStyles.backgroundImage
        if (blockStyles.backgroundSize) {
          sectionAttrs['background-size'] = blockStyles.backgroundSize
        }
        if (blockStyles.backgroundRepeat) {
          sectionAttrs['background-repeat'] = blockStyles.backgroundRepeat
        }
      } else if (blockData.backgroundType === 'color') {
        if (blockStyles.backgroundColor)
          sectionAttrs['background-color'] = blockStyles.backgroundColor
      }

      if (blockData.borderControl === 'all') {
        if (
          blockStyles.borderStyle &&
          blockStyles.borderStyle !== 'none' &&
          blockStyles.borderWidth &&
          blockStyles.borderColor
        ) {
          sectionAttrs['border'] =
            `${blockStyles.borderWidth} ${blockStyles.borderStyle} ${blockStyles.borderColor}`
        }
      } else if (blockData.borderControl === 'separate') {
        if (
          blockStyles.borderTopStyle &&
          blockStyles.borderTopStyle !== 'none' &&
          blockStyles.borderTopWidth &&
          blockStyles.borderTopColor
        )
          sectionAttrs['border-top'] =
            `${blockStyles.borderTopWidth} ${blockStyles.borderTopStyle} ${blockStyles.borderTopColor}`
        if (
          blockStyles.borderRightStyle &&
          blockStyles.borderRightStyle !== 'none' &&
          blockStyles.borderRightWidth &&
          blockStyles.borderRightColor
        )
          sectionAttrs['border-right'] =
            `${blockStyles.borderRightWidth} ${blockStyles.borderRightStyle} ${blockStyles.borderRightColor}`
        if (
          blockStyles.borderBottomStyle &&
          blockStyles.borderBottomStyle !== 'none' &&
          blockStyles.borderBottomWidth &&
          blockStyles.borderBottomColor
        )
          sectionAttrs['border-bottom'] =
            `${blockStyles.borderBottomWidth} ${blockStyles.borderBottomStyle} ${blockStyles.borderBottomColor}`
        if (
          blockStyles.borderLeftStyle &&
          blockStyles.borderLeftStyle !== 'none' &&
          blockStyles.borderLeftWidth &&
          blockStyles.borderLeftColor
        )
          sectionAttrs['border-left'] =
            `${blockStyles.borderLeftWidth} ${blockStyles.borderLeftStyle} ${blockStyles.borderLeftColor}`
      }

      if (blockStyles.borderRadius && blockStyles.borderRadius !== '0px') {
        sectionAttrs['border-radius'] = blockStyles.borderRadius
      }

      if (blockData.paddingControl === 'all') {
        if (blockStyles.padding && blockStyles.padding !== '0px') {
          sectionAttrs['padding'] = blockStyles.padding
        }
      } else if (blockData.paddingControl === 'separate') {
        if (blockStyles.paddingTop && blockStyles.paddingTop !== '0px')
          sectionAttrs['padding-top'] = blockStyles.paddingTop
        if (blockStyles.paddingRight && blockStyles.paddingRight !== '0px')
          sectionAttrs['padding-right'] = blockStyles.paddingRight
        if (blockStyles.paddingBottom && blockStyles.paddingBottom !== '0px')
          sectionAttrs['padding-bottom'] = blockStyles.paddingBottom
        if (blockStyles.paddingLeft && blockStyles.paddingLeft !== '0px')
          sectionAttrs['padding-left'] = blockStyles.paddingLeft
      }

      attributes = objectAsKebab(sectionAttrs)

      // Handle mj-group wrapping
      if (blockData.columnsOnMobile === true && children.length > 0) {
        const groupChildren = children
          .map((child) => treeToMjml(rootStyles, child, templateData, urlParams, indent + 4, block)) // Children of group indented further
          .filter((s) => s && s.trim() !== '')
          .join('\n')

        const groupSpace = indentPad(indent + 2)
        // Ensure group only renders if it has content
        if (groupChildren.trim() !== '') {
          childrenMjml = `${groupSpace}<mj-group>\n${groupChildren}\n${groupSpace}</mj-group>`
        } else {
          childrenMjml = '' // No group needed if children render empty
        }
        children = [] // Prevent default children processing below for section
      }

      break // Go to common children processing

    case 'column':
      tagName = 'mj-column'
      const columnAttrs: any = {
        'vertical-align': blockStyles.verticalAlign
      }

      // Calculate width based on parent kind
      if (parent && parent.children) {
        const parentKind = parent.kind
        const index = parent.children.findIndex((c) => c.id === block.id)
        if (index !== -1) {
          switch (parentKind) {
            case 'columns168':
              columnAttrs['width'] = index === 0 ? '66.66%' : '33.33%'
              break
            case 'columns204':
              columnAttrs['width'] = index === 0 ? '83.33%' : '16.66%'
              break
            case 'columns420':
              columnAttrs['width'] = index === 0 ? '16.66%' : '83.33%'
              break
            case 'columns816':
              columnAttrs['width'] = index === 0 ? '33.33%' : '66.66%'
              break
            // Add cases for columns888 (25%), columns1212 (50%), columns6666 (25%) - assuming equal division
            case 'columns888':
              columnAttrs['width'] = '25%'
              break
            case 'columns1212':
              columnAttrs['width'] = '50%'
              break
            case 'columns6666':
              columnAttrs['width'] = '25%'
              break
            // oneColumn case: width is implicitly 100% by MJML default for single column
            // default: equal width assumed if not specified (handled by MJML)
          }
        }
      }

      if (blockStyles.backgroundColor) {
        columnAttrs['background-color'] = blockStyles.backgroundColor
      }

      if (blockData.borderControl === 'all') {
        if (
          blockStyles.borderStyle &&
          blockStyles.borderStyle !== 'none' &&
          blockStyles.borderWidth &&
          blockStyles.borderColor
        ) {
          columnAttrs['border'] =
            `${blockStyles.borderWidth} ${blockStyles.borderStyle} ${blockStyles.borderColor}`
        }
      } else if (blockData.borderControl === 'separate') {
        if (
          blockStyles.borderTopStyle &&
          blockStyles.borderTopStyle !== 'none' &&
          blockStyles.borderTopWidth &&
          blockStyles.borderTopColor
        )
          columnAttrs['border-top'] =
            `${blockStyles.borderTopWidth} ${blockStyles.borderTopStyle} ${blockStyles.borderTopColor}`
        if (
          blockStyles.borderRightStyle &&
          blockStyles.borderRightStyle !== 'none' &&
          blockStyles.borderRightWidth &&
          blockStyles.borderRightColor
        )
          columnAttrs['border-right'] =
            `${blockStyles.borderRightWidth} ${blockStyles.borderRightStyle} ${blockStyles.borderRightColor}`
        if (
          blockStyles.borderBottomStyle &&
          blockStyles.borderBottomStyle !== 'none' &&
          blockStyles.borderBottomWidth &&
          blockStyles.borderBottomColor
        )
          columnAttrs['border-bottom'] =
            `${blockStyles.borderBottomWidth} ${blockStyles.borderBottomStyle} ${blockStyles.borderBottomColor}`
        if (
          blockStyles.borderLeftStyle &&
          blockStyles.borderLeftStyle !== 'none' &&
          blockStyles.borderLeftWidth &&
          blockStyles.borderLeftColor
        )
          columnAttrs['border-left'] =
            `${blockStyles.borderLeftWidth} ${blockStyles.borderLeftStyle} ${blockStyles.borderLeftColor}`
      }

      if (blockStyles.borderRadius && blockStyles.borderRadius !== '0px') {
        columnAttrs['border-radius'] = blockStyles.borderRadius
      }

      if (blockData.paddingControl === 'all') {
        if (blockStyles.padding && blockStyles.padding !== '0px') {
          columnAttrs['padding'] = blockStyles.padding
        }
      } else if (blockData.paddingControl === 'separate') {
        if (blockStyles.paddingTop && blockStyles.paddingTop !== '0px')
          columnAttrs['padding-top'] = blockStyles.paddingTop
        if (blockStyles.paddingRight && blockStyles.paddingRight !== '0px')
          columnAttrs['padding-right'] = blockStyles.paddingRight
        if (blockStyles.paddingBottom && blockStyles.paddingBottom !== '0px')
          columnAttrs['padding-bottom'] = blockStyles.paddingBottom
        if (blockStyles.paddingLeft && blockStyles.paddingLeft !== '0px')
          columnAttrs['padding-left'] = blockStyles.paddingLeft
      }

      attributes = objectAsKebab(columnAttrs)
      break // Go to common children processing

    case 'text':
    case 'heading':
      tagName = 'mj-text'
      const textAttrs: any = {
        align: blockData.align,
        padding: 0 // Override default mjml padding
      }

      if (blockData.backgroundColor) {
        textAttrs['container-background-color'] = blockData.backgroundColor
      }

      attributes = objectAsKebab(textAttrs)

      // --- Content Generation Logic ---
      blockData.editorData?.forEach((line: any) => {
        // Ensure line and line.children exist
        if (!line || !line.children) return

        let lineContent = ''
        line.children.forEach((part: any) => {
          // Ensure part exists
          if (!part) return

          let partContent = part.text || '' // Handle potentially missing text

          // Sanitize text to prevent potential XSS if data is untrusted
          // Basic sanitization: escape '<' and '>'
          partContent = partContent.replace(/</g, '&lt;').replace(/>/g, '&gt;')

          if (
            part.bold ||
            part.italic ||
            part.underlined ||
            part.fontSize ||
            part.fontColor ||
            part.fontFamily ||
            part.hyperlink
          ) {
            const spanStyles = []
            if (part.bold) spanStyles.push('font-weight: bold')
            if (part.italic) spanStyles.push('font-style: italic')
            if (part.underlined) spanStyles.push('text-decoration: underline')
            // Add !important for better email client compatibility, especially for font styles
            if (part.fontSize) spanStyles.push(`font-size: ${part.fontSize} !important`)
            if (part.fontColor) spanStyles.push(`color: ${part.fontColor} !important`)
            if (part.fontFamily) spanStyles.push(`font-family: ${part.fontFamily} !important`)

            if (part.hyperlink && part.hyperlink.url) {
              const hyperlinkStyles = []
              const hStyles = blockData.hyperlinkStyles || {} // Default if styles missing
              // Prioritize inline styles from span if they exist, otherwise use hyperlink styles
              if (part.fontColor || hStyles.color)
                hyperlinkStyles.push(`color: ${part.fontColor || hStyles.color} !important`)
              if (part.fontFamily || hStyles.fontFamily)
                hyperlinkStyles.push(
                  `font-family: ${part.fontFamily || hStyles.fontFamily} !important`
                )
              if (part.fontSize || hStyles.fontSize)
                hyperlinkStyles.push(`font-size: ${part.fontSize || hStyles.fontSize} !important`)
              if (part.fontStyle || hStyles.fontStyle)
                hyperlinkStyles.push(
                  `font-style: ${part.italic ? 'italic' : hStyles.fontStyle || 'normal'} !important`
                )
              if (part.fontWeight || hStyles.fontWeight)
                hyperlinkStyles.push(
                  `font-weight: ${part.bold ? 'bold' : hStyles.fontWeight || 'normal'} !important`
                )
              // Apply underline from span OR hyperlink setting, default ON for links
              const underline =
                part.underlined !== undefined ? part.underlined : hStyles.textDecoration !== 'none'
              hyperlinkStyles.push(
                `text-decoration: ${underline ? 'underline' : 'none'} !important`
              )

              const finalURL =
                part.hyperlink.disable_tracking !== true // Check explicitly for false/undefined
                  ? trackURL(part.hyperlink.url, urlParams)
                  : part.hyperlink.url

              // Add rel="noopener noreferrer" for security
              lineContent += `<a style="${hyperlinkStyles.join('; ')}" href="${finalURL}" target="_blank" rel="noopener noreferrer">${partContent}</a>`
            } else {
              // Only add span if styles are present
              if (spanStyles.length > 0) {
                lineContent += `<span style="${spanStyles.join('; ')}">${partContent}</span>`
              } else {
                lineContent += partContent // Add text directly if no span styles
              }
            }
          } else {
            lineContent += partContent
          }
        })

        const lineStyles = []
        // Ensure rootStyles and rootStyles[line.type] exist
        const lineTypeStyle = rootStyles && rootStyles[line.type] ? rootStyles[line.type] : {}

        // Add !important to styles for better compatibility
        if (lineTypeStyle.color) lineStyles.push(`color: ${lineTypeStyle.color} !important`)
        if (lineTypeStyle.fontFamily)
          lineStyles.push(`font-family: ${lineTypeStyle.fontFamily} !important`)
        if (lineTypeStyle.fontSize)
          lineStyles.push(`font-size: ${lineTypeStyle.fontSize} !important`)
        if (lineTypeStyle.fontStyle)
          lineStyles.push(`font-style: ${lineTypeStyle.fontStyle} !important`)
        if (lineTypeStyle.fontWeight)
          lineStyles.push(`font-weight: ${lineTypeStyle.fontWeight} !important`)
        if (lineTypeStyle.lineHeight)
          lineStyles.push(`line-height: ${lineTypeStyle.lineHeight} !important`)
        if (lineTypeStyle.letterSpacing)
          lineStyles.push(`letter-spacing: ${lineTypeStyle.letterSpacing} !important`)
        if (lineTypeStyle.textDecoration)
          lineStyles.push(`text-decoration: ${lineTypeStyle.textDecoration} !important`)
        if (lineTypeStyle.textTransform)
          lineStyles.push(`text-transform: ${lineTypeStyle.textTransform} !important`)

        // Padding
        if (lineTypeStyle.paddingControl === 'all') {
          if (lineTypeStyle.padding && lineTypeStyle.padding !== '0px')
            lineStyles.push(`padding: ${lineTypeStyle.padding} !important`)
        } else {
          if (lineTypeStyle.paddingTop && lineTypeStyle.paddingTop !== '0px')
            lineStyles.push(`padding-top: ${lineTypeStyle.paddingTop} !important`)
          if (lineTypeStyle.paddingRight && lineTypeStyle.paddingRight !== '0px')
            lineStyles.push(`padding-right: ${lineTypeStyle.paddingRight} !important`)
          if (lineTypeStyle.paddingBottom && lineTypeStyle.paddingBottom !== '0px')
            lineStyles.push(`padding-bottom: ${lineTypeStyle.paddingBottom} !important`)
          if (lineTypeStyle.paddingLeft && lineTypeStyle.paddingLeft !== '0px')
            lineStyles.push(`padding-left: ${lineTypeStyle.paddingLeft} !important`)
        }

        // Margin
        if (lineTypeStyle.marginControl === 'all') {
          if (lineTypeStyle.margin && lineTypeStyle.margin !== '0px')
            lineStyles.push(`margin: ${lineTypeStyle.margin} !important`)
          else lineStyles.push('margin: 0px !important') // Ensure default browser margins are reset if not specified
        } else {
          let marginStyles = ''
          if (lineTypeStyle.marginTop && lineTypeStyle.marginTop !== '0px')
            marginStyles += `margin-top: ${lineTypeStyle.marginTop} !important;`
          if (lineTypeStyle.marginRight && lineTypeStyle.marginRight !== '0px')
            marginStyles += `margin-right: ${lineTypeStyle.marginRight} !important;`
          if (lineTypeStyle.marginBottom && lineTypeStyle.marginBottom !== '0px')
            marginStyles += `margin-bottom: ${lineTypeStyle.marginBottom} !important;`
          if (lineTypeStyle.marginLeft && lineTypeStyle.marginLeft !== '0px')
            marginStyles += `margin-left: ${lineTypeStyle.marginLeft} !important;`

          if (marginStyles) {
            lineStyles.push(marginStyles.slice(0, -1)) // Remove trailing semicolon before adding to array (though it's harmless)
          } else {
            lineStyles.push('margin: 0px !important') // Reset if no specific margins set
          }
        }

        // Determine tag type (h1, h2, h3, p)
        const tag =
          line.type === 'paragraph' || !['h1', 'h2', 'h3'].includes(line.type) ? 'p' : line.type
        // Only add content if lineContent is not empty
        if (lineContent.trim()) {
          // Add newline after each block element for readability
          content += `<${tag}${lineStyles.length > 0 ? ` style="${lineStyles.join('; ')}"` : ''}>${lineContent}</${tag}>\n`
        }
      })

      // Remove trailing newline from content before Liquid processing
      content = content.trimEnd()

      // --- Liquid Processing ---
      if (
        templateData &&
        templateData.trim() !== '' &&
        (content.includes('{{') || content.includes('{%'))
      ) {
        try {
          const engine = new Liquid()
          let jsonData = {}
          try {
            jsonData = JSON.parse(templateData)
          } catch (jsonError) {
            console.error('Invalid JSON in templateData:', templateData, jsonError)
            // Add error as HTML comment
            content += `\n<!-- Invalid template data provided -->`
          }
          content = engine.parseAndRenderSync(content, jsonData || {})
        } catch (e: any) {
          console.error('Liquid rendering error in text block:', e)
          console.error('Content:', content)
          console.error('Data:', templateData)
          // Add error as HTML comment
          content += `\n<!-- Liquid error: ${e.message} -->`
        }
      }
      // --- End Content Generation ---

      children = [] // mj-text handles content directly
      break

    case 'image':
      tagName = 'mj-image'
      // Ensure image and wrapper data exist
      const imgData = blockData.image || {}
      const imgWrapData = blockData.wrapper || {}
      const imageAttrs: any = {
        align: imgWrapData.align,
        src: imgData.src,
        alt: imgData.alt || '', // Ensure alt text is always present, even if empty
        height: imgData.height === 'auto' ? undefined : imgData.height, // Don't include height="auto"
        // MJML width is unitless (pixels), remove 'px'. Use undefined if '100%'.
        width: imgData.width === '100%' ? undefined : imgData.width?.replace?.('px', ''),
        'fluid-on-mobile': imgData.fullWidthOnMobile === true ? 'true' : undefined, // Only include if true
        padding: '0' // Reset default padding initially
      }

      if (imgData.href) {
        imageAttrs['href'] =
          imgData.disable_tracking !== true ? trackURL(imgData.href, urlParams) : imgData.href
        imageAttrs['target'] = '_blank' // Open image links in new tab
        imageAttrs['rel'] = 'noopener noreferrer' // Security best practice
      }

      if (imgData.borderRadius && imgData.borderRadius !== '0px') {
        imageAttrs['border-radius'] = imgData.borderRadius
      }

      // Padding from wrapper
      if (imgWrapData.paddingControl === 'all') {
        if (imgWrapData.padding && imgWrapData.padding !== '0px')
          imageAttrs['padding'] = imgWrapData.padding
      } else if (imgWrapData.paddingControl === 'separate') {
        delete imageAttrs['padding']
        if (imgWrapData.paddingTop && imgWrapData.paddingTop !== '0px')
          imageAttrs['padding-top'] = imgWrapData.paddingTop
        if (imgWrapData.paddingRight && imgWrapData.paddingRight !== '0px')
          imageAttrs['padding-right'] = imgWrapData.paddingRight
        if (imgWrapData.paddingBottom && imgWrapData.paddingBottom !== '0px')
          imageAttrs['padding-bottom'] = imgWrapData.paddingBottom
        if (imgWrapData.paddingLeft && imgWrapData.paddingLeft !== '0px')
          imageAttrs['padding-left'] = imgWrapData.paddingLeft
      }

      // Border from wrapper
      if (imgWrapData.borderControl === 'all') {
        if (
          imgWrapData.borderStyle &&
          imgWrapData.borderStyle !== 'none' &&
          imgWrapData.borderWidth &&
          imgWrapData.borderColor
        ) {
          imageAttrs['border'] =
            `${imgWrapData.borderWidth} ${imgWrapData.borderStyle} ${imgWrapData.borderColor}`
        }
      } else if (imgWrapData.borderControl === 'separate') {
        if (
          imgWrapData.borderTopStyle &&
          imgWrapData.borderTopStyle !== 'none' &&
          imgWrapData.borderTopWidth &&
          imgWrapData.borderTopColor
        )
          imageAttrs['border-top'] =
            `${imgWrapData.borderTopWidth} ${imgWrapData.borderTopStyle} ${imgWrapData.borderTopColor}`
        if (
          imgWrapData.borderRightStyle &&
          imgWrapData.borderRightStyle !== 'none' &&
          imgWrapData.borderRightWidth &&
          imgWrapData.borderRightColor
        )
          imageAttrs['border-right'] =
            `${imgWrapData.borderRightWidth} ${imgWrapData.borderRightStyle} ${imgWrapData.borderRightColor}`
        if (
          imgWrapData.borderBottomStyle &&
          imgWrapData.borderBottomStyle !== 'none' &&
          imgWrapData.borderBottomWidth &&
          imgWrapData.borderBottomColor
        )
          imageAttrs['border-bottom'] =
            `${imgWrapData.borderBottomWidth} ${imgWrapData.borderBottomStyle} ${imgWrapData.borderBottomColor}`
        if (
          imgWrapData.borderLeftStyle &&
          imgWrapData.borderLeftStyle !== 'none' &&
          imgWrapData.borderLeftWidth &&
          imgWrapData.borderLeftColor
        )
          imageAttrs['border-left'] =
            `${imgWrapData.borderLeftWidth} ${imgWrapData.borderLeftStyle} ${imgWrapData.borderLeftColor}`
      }

      // Container background color from wrapper (applied to the container, not the image element itself)
      if (imgWrapData.backgroundColor) {
        imageAttrs['container-background-color'] = imgWrapData.backgroundColor
      }

      attributes = objectAsKebab(imageAttrs)
      children = [] // mj-image is self-contained
      break

    case 'button':
      tagName = 'mj-button'
      // Ensure button and wrapper data exist
      const btnData = blockData.button || {}
      const btnWrapData = blockData.wrapper || {}
      const buttonAttrs: any = {
        align: btnWrapData.align,
        // Track URL only if href exists and tracking is not disabled
        href: btnData.href
          ? btnData.disable_tracking !== true
            ? trackURL(btnData.href, urlParams)
            : btnData.href
          : undefined,
        target: btnData.href ? '_blank' : undefined, // Add target blank if href exists
        rel: btnData.href ? 'noopener noreferrer' : undefined, // Security best practice
        'background-color': btnData.backgroundColor,
        'font-family': btnData.fontFamily,
        'font-size': btnData.fontSize?.replace('px', ''), // Font size is unitless (pixels) in MJML
        'font-weight': btnData.fontWeight,
        'font-style': btnData.fontStyle === 'normal' ? undefined : btnData.fontStyle, // Don't include 'normal'
        color: btnData.color,
        padding: '0', // Reset default padding initially
        // Ensure inner padding has units (px), provide defaults
        'inner-padding': `${btnData.innerVerticalPadding || '10'}px ${btnData.innerHorizontalPadding || '25'}px`,
        'text-transform': btnData.textTransform === 'none' ? undefined : btnData.textTransform,
        'border-radius': btnData.borderRadius === '0px' ? undefined : btnData.borderRadius,
        // Width is unitless (pixels) or %
        width: btnData.width === 'auto' ? undefined : btnData.width?.replace('px', ''),
        'vertical-align': btnWrapData.verticalAlign // Use wrapper's vertical align
      }

      // Padding from wrapper
      if (btnWrapData.paddingControl === 'all') {
        if (btnWrapData.padding && btnWrapData.padding !== '0px')
          buttonAttrs['padding'] = btnWrapData.padding
      } else if (btnWrapData.paddingControl === 'separate') {
        delete buttonAttrs['padding'] // Remove global padding if specific ones are set
        if (btnWrapData.paddingTop && btnWrapData.paddingTop !== '0px')
          buttonAttrs['padding-top'] = btnWrapData.paddingTop
        if (btnWrapData.paddingRight && btnWrapData.paddingRight !== '0px')
          buttonAttrs['padding-right'] = btnWrapData.paddingRight
        if (btnWrapData.paddingBottom && btnWrapData.paddingBottom !== '0px')
          buttonAttrs['padding-bottom'] = btnWrapData.paddingBottom
        if (btnWrapData.paddingLeft && btnWrapData.paddingLeft !== '0px')
          buttonAttrs['padding-left'] = btnWrapData.paddingLeft
      }

      // Border from button itself
      if (btnData.borderControl === 'all') {
        if (
          btnData.borderStyle &&
          btnData.borderStyle !== 'none' &&
          btnData.borderWidth &&
          btnData.borderColor
        ) {
          buttonAttrs['border'] =
            `${btnData.borderWidth} ${btnData.borderStyle} ${btnData.borderColor}`
        }
      } else if (btnData.borderControl === 'separate') {
        if (
          btnData.borderTopStyle &&
          btnData.borderTopStyle !== 'none' &&
          btnData.borderTopWidth &&
          btnData.borderTopColor
        )
          buttonAttrs['border-top'] =
            `${btnData.borderTopWidth} ${btnData.borderTopStyle} ${btnData.borderTopColor}`
        if (
          btnData.borderRightStyle &&
          btnData.borderRightStyle !== 'none' &&
          btnData.borderRightWidth &&
          btnData.borderRightColor
        )
          buttonAttrs['border-right'] =
            `${btnData.borderRightWidth} ${btnData.borderRightStyle} ${btnData.borderRightColor}`
        if (
          btnData.borderBottomStyle &&
          btnData.borderBottomStyle !== 'none' &&
          btnData.borderBottomWidth &&
          btnData.borderBottomColor
        )
          buttonAttrs['border-bottom'] =
            `${btnData.borderBottomWidth} ${btnData.borderBottomStyle} ${btnData.borderBottomColor}`
        if (
          btnData.borderLeftStyle &&
          btnData.borderLeftStyle !== 'none' &&
          btnData.borderLeftWidth &&
          btnData.borderLeftColor
        )
          buttonAttrs['border-left'] =
            `${btnData.borderLeftWidth} ${btnData.borderLeftStyle} ${btnData.borderLeftColor}`
      }

      // Container background color from wrapper
      if (btnWrapData.backgroundColor) {
        buttonAttrs['container-background-color'] = btnWrapData.backgroundColor
      }

      attributes = objectAsKebab(buttonAttrs)
      // Sanitize button text
      content = (btnData.text || '').replace(/</g, '&lt;').replace(/>/g, '&gt;')
      children = [] // mj-button uses content
      break

    case 'divider':
      tagName = 'mj-divider'
      const divData = blockData || {} // Use blockData directly
      const dividerAttrs: any = {
        align: divData.align,
        'border-color': divData.borderColor,
        'border-style': divData.borderStyle,
        'border-width': divData.borderWidth,
        // Width is unitless (pixels) or %
        width: divData.width === '100%' ? undefined : divData.width?.replace('px', ''),
        padding: '0' // Reset default padding initially
      }

      if (divData.backgroundColor) {
        dividerAttrs['container-background-color'] = divData.backgroundColor
      }

      // Padding
      if (divData.paddingControl === 'all') {
        if (divData.padding && divData.padding !== '0px') dividerAttrs['padding'] = divData.padding
      } else if (divData.paddingControl === 'separate') {
        delete dividerAttrs['padding'] // Remove global padding if specific ones are set
        if (divData.paddingTop && divData.paddingTop !== '0px')
          dividerAttrs['padding-top'] = divData.paddingTop
        if (divData.paddingRight && divData.paddingRight !== '0px')
          dividerAttrs['padding-right'] = divData.paddingRight
        if (divData.paddingBottom && divData.paddingBottom !== '0px')
          dividerAttrs['padding-bottom'] = divData.paddingBottom
        if (divData.paddingLeft && divData.paddingLeft !== '0px')
          dividerAttrs['padding-left'] = divData.paddingLeft
      }

      attributes = objectAsKebab(dividerAttrs)
      children = [] // mj-divider is self-contained
      break

    case 'openTracking':
      // Output the tracking pixel using mj-image
      tagName = 'mj-image'
      attributes = {
        src: '{{ open_tracking_pixel_src }}', // Placeholder replaced server-side
        alt: '',
        height: '1px',
        width: '1px',
        padding: '0', // Ensure no extra space
        // Add styles for better compatibility, hidden visually but loaded
        style: 'display:block; max-height:1px; max-width:1px; visibility:hidden; mso-hide:all;'
      }
      children = []
      break

    case 'liquid':
      // Process liquid content and return it directly (will be inserted raw by parent)
      try {
        const engine = new Liquid()
        // Ensure templateData is valid JSON before parsing
        let jsonData = {}
        if (templateData && templateData.trim() !== '') {
          try {
            jsonData = JSON.parse(templateData)
          } catch (jsonError) {
            console.error('Invalid JSON in templateData for Liquid block:', templateData, jsonError)
            // Return error comment instead of throwing, allowing rest of template to render
            return `<!-- Invalid template data provided for Liquid block -->`
          }
        }
        const liquidCode = blockData.liquidCode || ''
        // Render and trim whitespace which might interfere with layout
        // Return the raw rendered output
        return engine.parseAndRenderSync(liquidCode, jsonData || {})
      } catch (e: any) {
        console.error('Liquid rendering error in liquid block:', e)
        console.error('Code:', blockData.liquidCode)
        console.error('Data:', templateData)
        // Return error comment directly, it will be placed raw in the MJML
        return `<!-- Liquid error: ${e.message} -->`
      }

    default:
      console.warn('MJML conversion not implemented for block kind:', block.kind, block)
      // Return an empty string or a comment to avoid breaking the structure
      return `${space}<!-- MJML Not Implemented: ${block.kind} -->`
  }

  // --- Common Processing for Children (if not handled above) ---
  // childrenMjml might already be set (e.g., for mj-group)
  if (children.length > 0 && !childrenMjml) {
    childrenMjml = children
      .map((child) => treeToMjml(rootStyles, child, templateData, urlParams, indent + 2, block))
      .filter((s) => s && s.trim() !== '')
      .join('\n')
  }

  // --- Assemble MJML String ---
  // Only proceed if tagName is defined
  if (!tagName) {
    // This case should ideally not be reached if all block kinds are handled
    // or return comments/empty strings. Return children if they exist.
    return childrenMjml || ''
  }

  const attrString = lineAttributes(attributes)
  const openTag = `${space}<${tagName}${attrString ? ' ' + attrString : ''}>`
  const closeTag = `</${tagName}>`

  // Handle content: Ensure content is not just whitespace before rendering inside tags
  const trimmedContent = content?.trim()
  // Handle children: Ensure childrenMjml is not just whitespace
  const trimmedChildrenMjml = childrenMjml?.trim()

  if (trimmedContent) {
    // Indent multi-line content properly. Avoid adding extra newlines if content is single line.
    const needsIndentation = content.includes('\n')
    const indentedContent = needsIndentation
      ? content
          .split('\n')
          .map((line) => `${space}  ${line}`)
          .join('\n')
      : `${space}  ${content}` // Single line content just gets padding

    return `${openTag}\n${indentedContent}\n${space}${closeTag}`
  } else if (trimmedChildrenMjml) {
    // Children are already indented from recursive calls
    return `${openTag}\n${childrenMjml}\n${space}${closeTag}`
  } else {
    // Self-closing style tags are not standard in MJML, use open/close tags for elements like mj-image, mj-divider
    return `${openTag}${closeTag}`
  }
}

// Updated Export Functions
export const ExportHTML = (
  editorData: BlockInterface | null,
  urlParams: any,
  templateData: string
): { html: string; errors: any[] } => {
  if (!editorData) {
    console.error('ExportHTML called with null editorData')
    return { html: '', errors: [{ message: 'No editor data provided.' }] }
  }
  // Ensure root styles are available, provide default empty object if not
  const rootStyles = editorData.data?.styles || {}
  const mjml = treeToMjml(rootStyles, editorData, templateData || '', urlParams || {}) // Provide defaults for safety
  // Add basic validation or error handling if needed before passing to mjml2html
  try {
    const result = mjml2html(mjml, {
      // MJML options if needed, e.g., validationLevel
      // validationLevel: 'strict'
    })
    return result // mjml2html returns { html, errors }
  } catch (e: any) {
    console.error('Error converting MJML to HTML:', e)
    console.error('Generated MJML:\n', mjml) // Log MJML with newline for readability
    // Escape MJML for display in HTML error message
    const escapedMjml = mjml.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
    return {
      html: `<html><body><h1>Error during HTML generation</h1><pre>${e.message}</pre><hr><h2>MJML Source:</h2><pre>${escapedMjml}</pre></body></html>`,
      errors: [{ message: `MJML Conversion Error: ${e.message}` }]
    }
  }
}

export const ExportMJML = (
  editorData: BlockInterface | null,
  urlParams: any,
  templateData: string
): string => {
  if (!editorData) {
    console.error('ExportMJML called with null editorData')
    return '<!-- No editor data provided -->'
  }
  // Ensure root styles are available, provide default empty object if not
  const rootStyles = editorData.data?.styles || {}
  return treeToMjml(rootStyles, editorData, templateData || '', urlParams || {}) // Provide defaults for safety
}
