import React from 'react'
import Prismjs from 'prismjs'

import '../styles/prism.css' /* Default theme */
import '../styles/prism-line-numbers.css' /* Line numbers plugin */

// Import the markup-templating dependency (required for Liquid)
import 'prismjs/components/prism-markup-templating'

// Import all needed languages
import 'prismjs/components/prism-xml-doc'
import 'prismjs/components/prism-liquid'
import 'prismjs/components/prism-json'
import 'prismjs/components/prism-bash'
// import 'prismjs/components/prism-typescript';
// import 'prismjs/components/prism-jsx';
// import 'prismjs/components/prism-tsx';

// Import all needed plugins
import 'prismjs/plugins/line-numbers/prism-line-numbers'

export function usePrismjs<T extends HTMLElement>(
  target: React.RefObject<T>,
  plugins: string[] = []
) {
  React.useLayoutEffect(() => {
    if (target.current) {
      if (plugins.length > 0) {
        target.current.classList.add(...plugins)
      }
      // Highlight all <pre><code>...</code></pre> blocks contained by this element
      Prismjs.highlightAllUnder(target.current)
    }
  }, [target, plugins])
}
