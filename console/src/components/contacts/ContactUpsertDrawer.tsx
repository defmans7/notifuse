import React from 'react'
import {
  Button,
  Drawer,
  Form,
  Input,
  Space,
  Select,
  Typography,
  Divider,
  Alert,
  InputNumber,
  DatePicker,
  message
} from 'antd'
import type { InputProps } from 'antd/es/input'
import type { TextAreaProps } from 'antd/es/input/TextArea'
import type { SelectProps, DefaultOptionType } from 'antd/es/select'
import type { DatePickerProps } from 'antd/es/date-picker'
import type { InputNumberProps } from 'antd/es/input-number'
import { CountriesFormOptions, TimezonesFormOptions } from '../utils/countries_timezones'
import { Languages } from '../utils/languages'
import { Contact } from '../../services/api/contacts'

const { Option } = Select
const { Text } = Typography
const { TextArea } = Input

// Custom form input components
const NullableInput: React.FC<InputProps & { name: string }> = ({ name, ...props }) => {
  const form = Form.useFormInstance()
  const value = Form.useWatch(name, form)

  return (
    <Space.Compact style={{ width: '100%' }}>
      <Input {...props} />
      <Button
        type={value === null ? 'primary' : 'default'}
        onClick={() => form.setFieldValue(name, null)}
        style={{ padding: '0 8px' }}
      >
        Null
      </Button>
    </Space.Compact>
  )
}

const NullableTextArea: React.FC<TextAreaProps & { name: string }> = ({ name, ...props }) => {
  const form = Form.useFormInstance()
  const value = Form.useWatch(name, form)

  return (
    <Space.Compact style={{ width: '100%' }}>
      <TextArea {...props} style={{ width: '100%', ...props.style }} />
      <Button
        type={value === null ? 'primary' : 'default'}
        onClick={() => form.setFieldValue(name, null)}
        style={{ padding: '0 8px' }}
      >
        Null
      </Button>
    </Space.Compact>
  )
}

const NullableInputNumber: React.FC<InputNumberProps & { name: string }> = ({ name, ...props }) => {
  const form = Form.useFormInstance()
  const value = Form.useWatch(name, form)

  return (
    <Space.Compact style={{ width: '100%' }}>
      <InputNumber {...props} style={{ width: '100%', ...props.style }} />
      <Button
        type={value === null ? 'primary' : 'default'}
        onClick={() => form.setFieldValue(name, null)}
        style={{ padding: '0 8px' }}
      >
        Null
      </Button>
    </Space.Compact>
  )
}

const NullableDatePicker: React.FC<DatePickerProps & { name: string }> = ({ name, ...props }) => {
  const form = Form.useFormInstance()
  const value = Form.useWatch(name, form)

  return (
    <Space.Compact style={{ width: '100%' }}>
      <DatePicker {...props} style={{ width: '100%', ...props.style }} />
      <Button
        type={value === null ? 'primary' : 'default'}
        onClick={() => form.setFieldValue(name, null)}
        style={{ padding: '0 8px' }}
      >
        Null
      </Button>
    </Space.Compact>
  )
}

const NullableSelect: React.FC<SelectProps & { name: string }> = ({ name, ...props }) => {
  const form = Form.useFormInstance()
  const value = Form.useWatch(name, form)

  return (
    <Space.Compact style={{ width: '100%' }}>
      <Select {...props} style={{ width: '100%', ...props.style }} />
      <Button
        type={value === null ? 'primary' : 'default'}
        onClick={() => form.setFieldValue(name, null)}
        style={{ padding: '0 8px' }}
      >
        Null
      </Button>
    </Space.Compact>
  )
}

const optionalFields = [
  { key: 'first_name', label: 'First Name' },
  { key: 'last_name', label: 'Last Name' },
  { key: 'phone', label: 'Phone' },
  { key: 'country', label: 'Country' },
  { key: 'external_id', label: 'External ID' },
  { key: 'timezone', label: 'Timezone' },
  { key: 'language', label: 'Language' },
  { key: 'address_line_1', label: 'Address Line 1' },
  { key: 'address_line_2', label: 'Address Line 2' },
  { key: 'postcode', label: 'Postcode' },
  { key: 'state', label: 'State' },
  { key: 'job_title', label: 'Job Title' },
  { key: 'lifetime_value', label: 'Lifetime Value' },
  { key: 'orders_count', label: 'Orders Count' },
  { key: 'last_order_at', label: 'Last Order At' },
  { key: 'custom_string_1', label: 'Custom String 1' },
  { key: 'custom_string_2', label: 'Custom String 2' },
  { key: 'custom_string_3', label: 'Custom String 3' },
  { key: 'custom_string_4', label: 'Custom String 4' },
  { key: 'custom_string_5', label: 'Custom String 5' },
  { key: 'custom_number_1', label: 'Custom Number 1' },
  { key: 'custom_number_2', label: 'Custom Number 2' },
  { key: 'custom_number_3', label: 'Custom Number 3' },
  { key: 'custom_number_4', label: 'Custom Number 4' },
  { key: 'custom_number_5', label: 'Custom Number 5' },
  { key: 'custom_datetime_1', label: 'Custom Date 1' },
  { key: 'custom_datetime_2', label: 'Custom Date 2' },
  { key: 'custom_datetime_3', label: 'Custom Date 3' },
  { key: 'custom_datetime_4', label: 'Custom Date 4' },
  { key: 'custom_datetime_5', label: 'Custom Date 5' },
  { key: 'custom_json_1', label: 'Custom JSON 1', type: 'json' },
  { key: 'custom_json_2', label: 'Custom JSON 2', type: 'json' },
  { key: 'custom_json_3', label: 'Custom JSON 3', type: 'json' },
  { key: 'custom_json_4', label: 'Custom JSON 4', type: 'json' },
  { key: 'custom_json_5', label: 'Custom JSON 5', type: 'json' }
]

interface ContactUpsertDrawerProps {
  workspaceId: string
  contact?: Contact
  onSuccess?: () => void
  buttonProps?: {
    type?: 'primary' | 'default' | 'dashed' | 'link' | 'text'
    icon?: React.ReactNode
    buttonContent?: React.ReactNode
    className?: string
    style?: React.CSSProperties
    size?: 'large' | 'middle' | 'small'
    disabled?: boolean
    loading?: boolean
    danger?: boolean
    ghost?: boolean
    block?: boolean
  }
}

export function ContactUpsertDrawer({
  workspaceId,
  contact,
  onSuccess,
  buttonProps
}: ContactUpsertDrawerProps) {
  const [drawerVisible, setDrawerVisible] = React.useState(false)
  const [selectedFields, setSelectedFields] = React.useState<string[]>([])
  const [selectedFieldToAdd, setSelectedFieldToAdd] = React.useState<string | null>(null)
  const [form] = Form.useForm()

  React.useEffect(() => {
    if (drawerVisible && contact) {
      // Pre-fill form with contact data
      const fieldsToShow = Object.keys(contact).filter(
        (key) =>
          key !== 'email' &&
          key !== 'workspace_id' &&
          contact[key as keyof Contact] !== undefined &&
          optionalFields.some((field) => field.key === key) // Only include fields that are in our optionalFields array
      )
      setSelectedFields(fieldsToShow)

      // Format JSON fields for display
      const formattedValues = { ...contact }
      fieldsToShow.forEach((field) => {
        if (field.startsWith('custom_json_')) {
          try {
            formattedValues[field as keyof Contact] = JSON.stringify(
              contact[field as keyof Contact],
              null,
              2
            )
          } catch (e) {
            console.error(`Error formatting JSON for field ${field}:`, e)
          }
        }
      })

      form.setFieldsValue(formattedValues)
    }
  }, [contact, form, drawerVisible])

  const handleAddField = () => {
    if (selectedFieldToAdd && !selectedFields.includes(selectedFieldToAdd)) {
      setSelectedFields([...selectedFields, selectedFieldToAdd])
      setSelectedFieldToAdd(null)
    }
  }

  const handleRemoveField = (field: string) => {
    setSelectedFields(selectedFields.filter((f) => f !== field))
    form.setFieldValue(field, undefined)
  }

  const handleSubmit = (values: any) => {
    const contactData = {
      ...values,
      workspace_id: workspaceId
    }

    // Parse JSON fields before submission
    selectedFields.forEach((field) => {
      if (field.startsWith('custom_json_')) {
        try {
          contactData[field] = JSON.parse(values[field])
        } catch (e) {
          message.error(`Invalid JSON in field ${field}`)
          return
        }
      }
    })

    // Here you would implement the API call to create/update contact
    console.log('Contact data to send:', contactData)

    setDrawerVisible(false)
    form.resetFields()
    setSelectedFields([])
    onSuccess?.()
  }

  const handleClose = () => {
    setDrawerVisible(false)
    form.resetFields()
    setSelectedFields([])
  }

  // Separate buttonContent from other props
  const { buttonContent, ...otherButtonProps } = buttonProps || {}
  const defaultButtonProps = {
    type: 'primary' as const,
    ...otherButtonProps
  }

  const renderFieldInput = (field: string, fieldInfo: (typeof optionalFields)[0]) => {
    if (field.startsWith('custom_json_')) {
      return (
        <NullableTextArea
          name={field}
          rows={4}
          placeholder={`Enter ${fieldInfo.label.toLowerCase()}`}
          style={{ fontFamily: 'monospace' }}
        />
      )
    }

    if (
      field === 'lifetime_value' ||
      field === 'orders_count' ||
      field === 'custom_number_1' ||
      field === 'custom_number_2' ||
      field === 'custom_number_3' ||
      field === 'custom_number_4' ||
      field === 'custom_number_5'
    ) {
      return (
        <NullableInputNumber name={field} placeholder={`Enter ${fieldInfo.label.toLowerCase()}`} />
      )
    }

    if (
      field === 'last_order_at' ||
      field === 'custom_datetime_1' ||
      field === 'custom_datetime_2' ||
      field === 'custom_datetime_3' ||
      field === 'custom_datetime_4' ||
      field === 'custom_datetime_5'
    ) {
      return <NullableDatePicker name={field} showTime format="YYYY-MM-DD HH:mm:ss" />
    }

    if (field === 'timezone') {
      return (
        <NullableSelect
          name={field}
          placeholder="Select timezone"
          options={TimezonesFormOptions}
          showSearch
          filterOption={(input: string, option: DefaultOptionType | undefined) =>
            String(option?.label ?? '')
              .toLowerCase()
              .includes(input.toLowerCase())
          }
        />
      )
    }

    if (field === 'country') {
      return (
        <NullableSelect
          name={field}
          placeholder="Select country"
          options={CountriesFormOptions}
          showSearch
          filterOption={(input: string, option: DefaultOptionType | undefined) =>
            String(option?.label ?? '')
              .toLowerCase()
              .includes(input.toLowerCase())
          }
        />
      )
    }

    if (field === 'language') {
      return (
        <NullableSelect
          name={field}
          placeholder="Select language"
          options={Languages}
          showSearch
          filterOption={(input: string, option: DefaultOptionType | undefined) =>
            String(option?.label ?? '')
              .toLowerCase()
              .includes(input.toLowerCase())
          }
        />
      )
    }

    return <NullableInput name={field} placeholder={`Enter ${fieldInfo.label.toLowerCase()}`} />
  }

  return (
    <>
      <Button onClick={() => setDrawerVisible(true)} {...defaultButtonProps}>
        {buttonContent || (contact ? 'Update Contact' : 'Insert Contact')}
      </Button>

      <Drawer
        title={contact ? 'Update Contact' : 'Insert Contact'}
        width={500}
        open={drawerVisible}
        onClose={handleClose}
        extra={
          <Space>
            <Button onClick={handleClose}>Cancel</Button>
            <Button type="primary" onClick={() => form.submit()}>
              Save
            </Button>
          </Space>
        }
      >
        <Alert
          description="If a contact with this email already exists, the provided fields will be overwritten. Fields not included in the form will remain unchanged."
          type="info"
          showIcon
          style={{ marginBottom: '16px' }}
        />
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item
            name="email"
            label="Email"
            rules={[
              { required: true, message: 'Email is required' },
              { type: 'email', message: 'Please enter a valid email' }
            ]}
          >
            <Input placeholder="Enter email address" />
          </Form.Item>

          {selectedFields.map((field) => {
            const fieldInfo = optionalFields.find((f) => f.key === field)
            if (!fieldInfo) return null // Skip rendering if fieldInfo is undefined

            return (
              <Form.Item
                key={field}
                name={field}
                label={
                  <Space>
                    <span>{fieldInfo.label}</span>
                    <Button
                      type="text"
                      size="small"
                      danger
                      onClick={() => handleRemoveField(field)}
                      style={{ marginLeft: 'auto' }}
                    >
                      Remove
                    </Button>
                  </Space>
                }
              >
                {renderFieldInput(field, fieldInfo)}
              </Form.Item>
            )
          })}

          <Divider />

          <div>
            <Text strong>Add an optional field</Text>
            <div className="mt-2 flex gap-2">
              <Select
                placeholder="Select a field"
                style={{ width: '80%' }}
                value={selectedFieldToAdd}
                onChange={setSelectedFieldToAdd}
              >
                {optionalFields
                  .filter((field) => !selectedFields.includes(field.key))
                  .map((field) => (
                    <Option key={field.key} value={field.key}>
                      {field.label}
                    </Option>
                  ))}
              </Select>
              <Button type="primary" onClick={handleAddField} disabled={!selectedFieldToAdd}>
                Add
              </Button>
            </div>
          </div>
        </Form>
      </Drawer>
    </>
  )
}
