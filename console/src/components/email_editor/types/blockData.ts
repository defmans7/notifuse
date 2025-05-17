// Common types used across multiple blocks
export interface BaseStyles {
  paddingTop?: string
  paddingRight?: string
  paddingBottom?: string
  paddingLeft?: string
  padding?: string
  borderStyle?: string
  borderWidth?: string
  borderColor?: string
  borderRadius?: string
}

export interface WrapperStyles {
  align: string
  paddingControl: 'all' | 'separate'
  padding?: string
  paddingTop?: string
  paddingRight?: string
  paddingBottom?: string
  paddingLeft?: string
}

// Button block data
export interface ButtonBlockData {
  button: {
    text: string
    href: string
    backgroundColor: string
    fontFamily: string
    fontSize: string
    fontWeight: number
    fontStyle: string
    color: string
    innerVerticalPadding: string
    innerHorizontalPadding: string
    width: string
    textTransform: string
    borderRadius: string
    disable_tracking: boolean
    borderControl: 'all' | 'separate'
  }
  wrapper: WrapperStyles
}

// Image block data
export interface ImageBlockData {
  image: {
    src: string
    alt: string
    href: string
    width: string
    borderControl: 'all' | 'separate'
  }
  wrapper: WrapperStyles
}

// Column block data
export interface ColumnBlockData {
  styles: {
    verticalAlign: 'top' | 'middle' | 'bottom'
    backgroundColor?: string
    minHeight?: string
  } & BaseStyles
  paddingControl: 'all' | 'separate'
  borderControl: 'all' | 'separate'
}

// Divider block data
export interface DividerBlockData {
  align: 'left' | 'center' | 'right'
  borderColor: string
  borderStyle: string
  borderWidth: string
  backgroundColor?: string
  width: string
  paddingControl: 'all' | 'separate'
  padding?: string
  paddingTop?: string
  paddingRight?: string
  paddingBottom?: string
  paddingLeft?: string
}

// Section block data
export interface SectionBlockData {
  columnsOnMobile: boolean
  stackColumnsAtWidth: number
  backgroundType: 'color' | 'image'
  paddingControl: 'all' | 'separate'
  borderControl: 'all' | 'separate'
  styles: {
    textAlign: 'left' | 'center' | 'right' | 'justify'
    backgroundRepeat?: 'repeat' | 'no-repeat' | 'repeat-x' | 'repeat-y'
    padding?: string
    borderWidth?: string
    borderStyle?: string
    borderColor?: string
    backgroundColor?: string
    backgroundImage?: string
    backgroundSize?: 'cover' | 'contain'
  } & BaseStyles
}

// Text block data
export interface TextBlockData {
  align: 'left' | 'center' | 'right'
  width: string
  hyperlinkStyles: {
    color: string
    textDecoration: string
    fontFamily: string
    fontSize: string
    fontWeight: number
    fontStyle: string
    textTransform: string
  }
  editorData: Array<{
    type: string
    children: Array<{
      text: string
    }>
  }>
  backgroundColor?: string
  paddingControl?: 'all' | 'separate'
  padding?: string
  paddingTop?: string
  paddingRight?: string
  paddingBottom?: string
  paddingLeft?: string
}

// Heading block data
export interface HeadingBlockData {
  type: 'h1' | 'h2' | 'h3'
  align: 'left' | 'center' | 'right'
  width: string
  editorData: Array<{
    type: string
    children: Array<{
      text: string
    }>
  }>
  backgroundColor?: string
  paddingControl?: 'all' | 'separate'
  padding?: string
  paddingTop?: string
  paddingRight?: string
  paddingBottom?: string
  paddingLeft?: string
}

// Liquid block data
export interface LiquidBlockData {
  liquidCode: string
}

// Root block data
export interface RootBlockData {
  styles: {
    body: {
      width: string
      margin: string
      backgroundColor: string
    }
    h1: {
      color: string
      fontSize: string
      fontStyle: string
      fontWeight: number
      paddingControl: 'all' | 'separate'
      padding?: string
      paddingTop?: string
      paddingRight?: string
      paddingBottom?: string
      paddingLeft?: string
      margin: string
      fontFamily: string
    }
    h2: {
      color: string
      fontSize: string
      fontStyle: string
      fontWeight: number
      paddingControl: 'all' | 'separate'
      padding?: string
      paddingTop?: string
      paddingRight?: string
      paddingBottom?: string
      paddingLeft?: string
      margin: string
      fontFamily: string
    }
    h3: {
      color: string
      fontSize: string
      fontStyle: string
      fontWeight: number
      paddingControl: 'all' | 'separate'
      padding?: string
      paddingTop?: string
      paddingRight?: string
      paddingBottom?: string
      paddingLeft?: string
      margin: string
      fontFamily: string
    }
    paragraph: {
      color: string
      fontSize: string
      fontStyle: string
      fontWeight: number
      paddingControl: 'all' | 'separate'
      padding?: string
      paddingTop?: string
      paddingRight?: string
      paddingBottom?: string
      paddingLeft?: string
      margin: string
      fontFamily: string
    }
    hyperlink: {
      color: string
      textDecoration: string
      fontFamily: string
      fontSize: string
      fontWeight: number
      fontStyle: string
      textTransform: string
    }
  }
}

// Column layout block data (for all column variations)
export interface ColumnLayoutBlockData extends SectionBlockData {
  columns: number[]
}

// One column block data
export interface OneColumnBlockData extends SectionBlockData {
  columns: [24]
}

// Column variations (168, 204, 420, 816, 888, 1212, 6666)
export interface Columns168BlockData extends ColumnLayoutBlockData {
  columns: [16, 8]
}

export interface Columns204BlockData extends ColumnLayoutBlockData {
  columns: [20, 4]
}

export interface Columns420BlockData extends ColumnLayoutBlockData {
  columns: [4, 20]
}

export interface Columns816BlockData extends ColumnLayoutBlockData {
  columns: [8, 16]
}

export interface Columns888BlockData extends ColumnLayoutBlockData {
  columns: [8, 8, 8]
}

export interface Columns1212BlockData extends ColumnLayoutBlockData {
  columns: [12, 12]
}

export interface Columns6666BlockData extends ColumnLayoutBlockData {
  columns: [6, 6, 6, 6]
}
