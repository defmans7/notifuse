import { InputNumber, Select, Space } from 'antd'
import { Fonts } from './ElementForms'
import { ColorPickerLight } from './ColorPicker'

// Font weights
const FontWeights = [
  { label: <span style={{ fontWeight: 100 }}>100</span>, value: 100 },
  { label: <span style={{ fontWeight: 200 }}>200</span>, value: 200 },
  { label: <span style={{ fontWeight: 300 }}>300</span>, value: 300 },
  { label: <span style={{ fontWeight: 400 }}>400</span>, value: 400 },
  { label: <span style={{ fontWeight: 500 }}>500</span>, value: 500 },
  { label: <span style={{ fontWeight: 600 }}>600</span>, value: 600 },
  { label: <span style={{ fontWeight: 700 }}>700</span>, value: 700 },
  { label: <span style={{ fontWeight: 800 }}>800</span>, value: 800 },
  { label: <span style={{ fontWeight: 900 }}>900</span>, value: 900 }
]

// Text transforms
const TextTransforms = [
  { label: 'None', value: 'none' },
  { label: 'Uppercase', value: 'uppercase' },
  { label: 'Capitalize', value: 'capitalize' },
  { label: 'Lowercase', value: 'lowercase' }
]

// Text decorations
const TextDecorations = [
  { label: 'None', value: 'none' },
  { label: 'Underline', value: 'underline' },
  { label: 'Line-through', value: 'line-through' }
]

interface FontStyleInputsProps {
  styles: any
  onChange: (styles: any) => void
}

export const FontStyleInputs = (props: FontStyleInputsProps) => {
  return (
    <>
      <Space.Compact style={{ width: '100%', marginBottom: '8px' }} size="small">
        <ColorPickerLight
          style={{ width: '13%' }}
          size="small"
          value={props.styles.color}
          onChange={(newColor) => {
            props.styles.color = newColor
            props.onChange(props.styles)
          }}
        />
        <Select
          size="small"
          style={{ width: '43.5%' }}
          value={props.styles.fontFamily}
          onChange={(value) => {
            props.styles.fontFamily = value
            props.onChange(props.styles)
          }}
          options={Fonts}
        />
        <Select
          size="small"
          style={{ width: '43.5%' }}
          value={props.styles.fontWeight}
          onChange={(value) => {
            props.styles.fontWeight = value
            props.onChange(props.styles)
          }}
          options={FontWeights}
        />
      </Space.Compact>

      <Space.Compact style={{ width: '100%' }} size="small">
        <InputNumber
          style={{ width: '33.33%' }}
          value={parseInt(props.styles.fontSize || '16px')}
          onChange={(value) => {
            props.styles.fontSize = value + 'px'
            props.onChange(props.styles)
          }}
          defaultValue={parseInt(props.styles.fontSize || '16px')}
          size="small"
          min={0}
          parser={(value: string | undefined) => {
            if (value === undefined) return 0
            return parseInt(value.replace('px', ''))
          }}
          formatter={(value) => value + 'px'}
        />
        <Select
          size="small"
          style={{ width: '33.33%' }}
          value={props.styles.textTransform}
          onChange={(value) => {
            props.styles.textTransform = value
            props.onChange(props.styles)
          }}
          defaultValue="none"
          options={TextTransforms}
        />
        <Select
          size="small"
          style={{ width: '33.33%' }}
          value={props.styles.textDecoration}
          onChange={(value) => {
            props.styles.textDecoration = value
            props.onChange(props.styles)
          }}
          defaultValue="none"
          options={TextDecorations}
        />
      </Space.Compact>
    </>
  )
}
