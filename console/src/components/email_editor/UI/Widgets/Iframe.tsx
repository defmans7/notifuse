import { useEffect, useRef } from 'react'
import _ from 'lodash'

const iframeStyles = {
  container: {
    position: 'relative' as const,
    height: '100%'
  },
  normal: {
    height: '200px',
    width: '100%'
  },
  actions: {
    position: 'absolute' as const,
    top: '12px',
    right: '12px',
    zIndex: 1
  }
}

interface IframeSandboxProps {
  id: string
  className?: string
  content: string
  sizeSelector: string
  style?: React.CSSProperties
}

const IframeSandbox = (props: IframeSandboxProps) => {
  const { id, className, content, sizeSelector, style } = props
  const containerRef = useRef<HTMLDivElement>(null)
  const iframeRef = useRef<HTMLIFrameElement>(null)

  const resize = () => {
    const el = document.querySelector(sizeSelector)
    const parentHeight = el ? parseInt(window.getComputedStyle(el).height) : 0

    const container = containerRef.current
    const iframe = iframeRef.current

    if (container) {
      container.style.height = parentHeight - 30 + 'px'
    }

    if (iframe) {
      iframe.style.height = parentHeight - 30 + 'px'
    }
  }

  useEffect(() => {
    // wait a bit to be sure parent element are well inserted in the dom
    // and height can be computed correctly
    const timer = window.setTimeout(() => {
      resize()
    }, 100)

    return () => clearTimeout(timer)
  }, [sizeSelector])

  return (
    <div style={iframeStyles.container} ref={containerRef}>
      <iframe
        style={{ ...style, border: 'none' }}
        title={`iframe-${id}`}
        srcDoc={content}
        id={id}
        ref={iframeRef}
        className={className}
      ></iframe>
    </div>
  )
}

export default IframeSandbox
