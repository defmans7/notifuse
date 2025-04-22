import React, { useState, useRef, useEffect } from 'react'
import {
  Button,
  Drawer,
  Upload,
  Form,
  Select,
  Progress,
  Space,
  Typography,
  Alert,
  message,
  Modal,
  Tag
} from 'antd'
import {
  UploadOutlined,
  PauseCircleOutlined,
  PlayCircleOutlined,
  CloseCircleOutlined,
  UserAddOutlined
} from '@ant-design/icons'
import type { UploadProps } from 'antd'
import Papa from 'papaparse'
import type { ParseResult } from 'papaparse'
import { Contact } from '../../services/api/contacts'
import { contactsApi } from '../../services/api/contacts'
import { contactListApi } from '../../services/api/contact_list'
import { List } from '../../services/api/types'

const { Text } = Typography
const { Option } = Select
const { Dragger } = Upload

// Batch size for processing
const BATCH_SIZE = 25
const PREVIEW_ROWS = 15
const PROGRESS_SAVE_INTERVAL = 10000 // 10 seconds

// Function to generate a unique storage key for each workspace+file combination
const getProgressStorageKey = (workspaceId: string, fileName: string): string => {
  return `csv_upload_progress_${workspaceId}_${fileName}`
}

export interface ContactsCsvUploadDrawerProps {
  workspaceId: string
  lists?: List[]
  onSuccess?: () => void
  isVisible: boolean
  onClose: () => void
}

// Create context is moved to ContactsCsvUploadProvider.tsx

interface CsvData {
  headers: string[]
  rows: any[][]
  preview: any[][]
}

interface SavedProgress {
  fileName: string
  currentRow: number
  totalRows: number
  currentBatch: number
  totalBatches: number
  mappings: Record<string, string>
  selectedListIds: string[]
  timestamp: number
}

// Define contact fields for mapping
const contactFields = [
  { key: 'email', label: 'Email', required: true },
  { key: 'external_id', label: 'External ID' },
  { key: 'first_name', label: 'First Name' },
  { key: 'last_name', label: 'Last Name' },
  { key: 'phone', label: 'Phone' },
  { key: 'country', label: 'Country' },
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
  { key: 'custom_json_1', label: 'Custom JSON 1' },
  { key: 'custom_json_2', label: 'Custom JSON 2' },
  { key: 'custom_json_3', label: 'Custom JSON 3' },
  { key: 'custom_json_4', label: 'Custom JSON 4' },
  { key: 'custom_json_5', label: 'Custom JSON 5' }
]

export function ContactsCsvUploadDrawer({
  workspaceId,
  lists = [],
  onSuccess,
  isVisible,
  onClose
}: ContactsCsvUploadDrawerProps) {
  const [form] = Form.useForm()
  const [csvData, setCsvData] = useState<CsvData | null>(null)
  const [fileName, setFileName] = useState<string>('')
  const [uploading, setUploading] = useState<boolean>(false)
  const [uploadProgress, setUploadProgress] = useState<number>(0)
  const [currentBatch, setCurrentBatch] = useState<number>(0)
  const [totalBatches, setTotalBatches] = useState<number>(0)
  const [currentRow, setCurrentRow] = useState<number>(0)
  const [totalRows, setTotalRows] = useState<number>(0)
  const [paused, setPaused] = useState<boolean>(false)
  const [processingCancelled, setProcessingCancelled] = useState<boolean>(false)
  const [uploadError, setUploadError] = useState<string | null>(null)
  const [savedProgressExists, setSavedProgressExists] = useState<boolean>(false)
  const [selectedListIds, setSelectedListIds] = useState<string[]>([])
  const [uploadComplete, setUploadComplete] = useState<boolean>(false)
  const [successCount, setSuccessCount] = useState<number>(0)
  const [failureCount, setFailureCount] = useState<number>(0)
  const [errorDetails, setErrorDetails] = useState<
    Array<{ line: number; email: string; error: string }>
  >([])
  const progressSaveInterval = useRef<NodeJS.Timeout | null>(null)
  const uploadRef = useRef<{
    abort: () => void
    resume: () => void
    isPaused: boolean
  }>({
    abort: () => {},
    resume: () => {},
    isPaused: false
  })

  // Check for saved progress
  const checkForSavedProgress = (filename: string) => {
    try {
      const savedData = localStorage.getItem(getProgressStorageKey(workspaceId, filename))
      if (savedData) {
        const savedProgress: SavedProgress = JSON.parse(savedData)

        // Check if the filename matches and it's recent (within 7 days)
        const isRecent = Date.now() - savedProgress.timestamp < 7 * 24 * 60 * 60 * 1000

        if (savedProgress.fileName === filename && isRecent) {
          setSavedProgressExists(true)
          return savedProgress
        }
      }
    } catch (error) {
      console.error('Error checking for saved progress:', error)
    }

    setSavedProgressExists(false)
    return null
  }

  // Save progress to localStorage
  const saveProgress = () => {
    if (!fileName || !csvData || !form) return

    try {
      const mappings = form.getFieldValue('mappings') || {}
      const selectedListIds = form.getFieldValue('selectedListIds') || []

      const progressData: SavedProgress = {
        fileName,
        currentRow,
        totalRows,
        currentBatch,
        totalBatches,
        mappings,
        selectedListIds,
        timestamp: Date.now()
      }

      localStorage.setItem(
        getProgressStorageKey(workspaceId, fileName),
        JSON.stringify(progressData)
      )
    } catch (error) {
      console.error('Error saving progress:', error)
    }
  }

  // Clear saved progress
  const clearSavedProgress = () => {
    try {
      localStorage.removeItem(getProgressStorageKey(workspaceId, fileName))
      setSavedProgressExists(false)
    } catch (error) {
      console.error('Error clearing saved progress:', error)
    }
  }

  // Start progress auto-save interval
  const startProgressSaveInterval = () => {
    if (progressSaveInterval.current) {
      clearInterval(progressSaveInterval.current)
    }

    progressSaveInterval.current = setInterval(() => {
      if (uploading && !processingCancelled) {
        saveProgress()
      }
    }, PROGRESS_SAVE_INTERVAL)
  }

  // Stop progress auto-save interval
  const stopProgressSaveInterval = () => {
    if (progressSaveInterval.current) {
      clearInterval(progressSaveInterval.current)
      progressSaveInterval.current = null
    }
  }

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      stopProgressSaveInterval()
    }
  }, [])

  // Handle restore progress
  const handleRestoreProgress = (savedProgress: SavedProgress) => {
    if (!csvData) return

    // Restore form mappings
    form.setFieldsValue({
      mappings: savedProgress.mappings,
      selectedListIds: savedProgress.selectedListIds || []
    })

    // Set state for resuming
    setCurrentRow(savedProgress.currentRow)
    setTotalRows(savedProgress.totalRows)
    setCurrentBatch(savedProgress.currentBatch)
    setTotalBatches(savedProgress.totalBatches)
    setUploadProgress(Math.round((savedProgress.currentRow / savedProgress.totalRows) * 100))
    setSelectedListIds(savedProgress.selectedListIds || [])

    message.success('Previous upload progress restored')
  }

  const handleCloseDrawer = () => {
    if (uploading) {
      Modal.confirm({
        title: 'Cancel Upload?',
        content:
          'Are you sure you want to cancel the upload process? Progress will be saved and you can resume later.',
        onOk: () => {
          if (uploading && !processingCancelled) {
            saveProgress()
          }
          cancelUpload()
          onClose()
        }
      })
    } else {
      onClose()
    }
  }

  const beforeUpload = (file: File) => {
    const isCsv = file.type === 'text/csv' || file.name.endsWith('.csv')
    if (!isCsv) {
      message.error('You can only upload CSV files!')
      return Upload.LIST_IGNORE
    }

    setFileName(file.name)

    // Check for saved progress
    const savedProgress = checkForSavedProgress(file.name)

    // Parse CSV file
    Papa.parse<string[]>(file, {
      header: false,
      complete: (results: ParseResult<string[]>) => {
        if (results.data && results.data.length > 0) {
          const headers = results.data[0]
          const rows = results.data.slice(1)
          const preview = rows.slice(0, PREVIEW_ROWS)

          setCsvData({
            headers,
            rows,
            preview
          })

          // Set total rows
          setTotalRows(rows.length)

          // Auto-map fields if column names match contact field names
          const initialMappings: Record<string, string> = {}
          headers.forEach((header) => {
            const matchingField = contactFields.find(
              (field) =>
                field.key.toLowerCase() === header.toLowerCase() ||
                field.label.toLowerCase() === header.toLowerCase()
            )
            if (matchingField) {
              initialMappings[matchingField.key] = header
            }
          })

          // If we have saved progress and the user hasn't chosen to restore yet,
          // we'll wait for their decision before setting field values
          if (!savedProgress || savedProgressExists) {
            form.setFieldsValue({ mappings: initialMappings })
          }

          // If there's saved progress, show modal to ask user
          if (savedProgress && !savedProgressExists) {
            Modal.confirm({
              title: 'Resume Previous Upload',
              content: `A previous upload for "${file.name}" was found (${new Date(savedProgress.timestamp).toLocaleString()}). Would you like to resume from where you left off?`,
              okText: 'Resume',
              cancelText: 'Start New',
              onOk: () => {
                handleRestoreProgress(savedProgress)
              },
              onCancel: () => {
                // Start fresh - clear saved progress
                clearSavedProgress()
                form.setFieldsValue({ mappings: initialMappings })
              }
            })
          }
        } else {
          message.error('The CSV file appears to be empty or invalid.')
        }
      },
      error: (error: Error) => {
        message.error(`Error parsing CSV: ${error.message}`)
      }
    })

    return false // Prevent default upload behavior
  }

  const uploadProps: UploadProps = {
    name: 'file',
    multiple: false,
    accept: '.csv',
    showUploadList: false,
    beforeUpload
  }

  const startUpload = async () => {
    try {
      const mappings = form.getFieldValue('mappings')

      // Validate email mapping is set
      if (!mappings.email) {
        message.error('Email field mapping is required')
        return
      }

      if (!csvData) {
        message.error('No CSV data available')
        return
      }

      setUploading(true)
      setProcessingCancelled(false)
      setPaused(false)
      setUploadError(null)
      setSuccessCount(0)
      setFailureCount(0)
      setErrorDetails([])
      setUploadComplete(false)

      // Calculate total rows and batches if starting fresh
      let startRow = currentRow
      if (startRow === 0) {
        const totalBatches = Math.ceil(csvData.rows.length / BATCH_SIZE)
        setTotalRows(csvData.rows.length)
        setTotalBatches(totalBatches)
      }

      // Start auto-saving progress
      startProgressSaveInterval()

      // Process in batches
      let rowIndex = startRow
      let batch = currentBatch || 1
      let successCount = 0
      let failureCount = 0
      let errors: Array<{ line: number; email: string; error: string }> = []

      const processNextBatch = async () => {
        if (processingCancelled) {
          return
        }

        setCurrentBatch(batch)
        const end = Math.min(rowIndex + BATCH_SIZE, totalRows)
        const batchRows = csvData.rows.slice(rowIndex, end)

        const contacts: Partial<Contact>[] = batchRows.map((row) => {
          const contact: Partial<Contact> = {}

          // Map CSV columns to contact fields
          Object.entries(mappings).forEach(([contactField, csvColumn]) => {
            if (csvColumn) {
              const columnIndex = csvData.headers.indexOf(csvColumn as string)
              if (columnIndex !== -1) {
                let value = row[columnIndex]

                // Handle special field types
                if (contactField.startsWith('custom_json_') && value) {
                  try {
                    value = JSON.parse(value)
                  } catch (e) {
                    // Set to null if not valid JSON
                    value = null
                  }
                } else if (
                  contactField.startsWith('custom_number_') ||
                  contactField === 'lifetime_value' ||
                  contactField === 'orders_count'
                ) {
                  if (value && value.trim() !== '') {
                    value = Number(value)
                  } else {
                    value = null
                  }
                }

                ;(contact as any)[contactField] = value !== '' ? value : null
              }
            }
          })

          return contact
        })

        // Filter out contacts without email
        const validContacts = contacts.filter((contact) => contact.email)

        // Process each contact
        for (let i = 0; i < validContacts.length; i++) {
          const contact = validContacts[i]
          const csvLineNumber = rowIndex + i + 2 // +2 because headers are line 1, and rowIndex is 0-based
          if (processingCancelled) break

          // Check if paused
          while (uploadRef.current.isPaused && !processingCancelled) {
            await new Promise((resolve) => setTimeout(resolve, 100))
          }

          try {
            await contactsApi.upsert({
              workspace_id: workspaceId,
              contact
            })

            // Add contact to selected lists if any
            const listsToAdd = selectedListIds || []
            for (const listId of listsToAdd) {
              if (processingCancelled) break

              try {
                await contactListApi.addContact({
                  workspace_id: workspaceId,
                  email: contact.email!,
                  list_id: listId,
                  status: 'active'
                })
              } catch (listError) {
                console.error(`Error adding contact to list ${listId}:`, listError)
                // Continue with next list - don't interrupt the process
              }
            }

            successCount++
            setSuccessCount(successCount)
          } catch (error) {
            console.error('Error upserting contact:', error)
            failureCount++
            setFailureCount(failureCount)

            // Store error details (limit to 100)
            if (errors.length < 100) {
              errors.push({
                line: csvLineNumber,
                email: contact.email || 'Unknown',
                error: error instanceof Error ? error.message : 'Unknown error'
              })
              setErrorDetails(errors)
            }
            // Continue with next contact
          }
        }

        rowIndex = end
        setCurrentRow(rowIndex)
        const progress = Math.min(Math.round((rowIndex / totalRows) * 100), 100)
        setUploadProgress(progress)

        // Save progress after each batch
        saveProgress()

        if (rowIndex < totalRows && !processingCancelled) {
          batch++
          await processNextBatch()
        } else {
          stopProgressSaveInterval()
          setUploading(false)
          setUploadComplete(true)
          if (!processingCancelled) {
            message.success(
              `Upload completed with ${successCount} successful contacts and ${failureCount} failures`
            )
            // Clear saved progress when complete
            clearSavedProgress()
            onSuccess?.()
          }
        }
      }

      // Setup abort and resume controls
      uploadRef.current = {
        abort: () => {
          setProcessingCancelled(true)
          saveProgress() // Save progress on abort
        },
        resume: () => {
          uploadRef.current.isPaused = false
          setPaused(false)
        },
        isPaused: false
      }

      await processNextBatch()
    } catch (error) {
      stopProgressSaveInterval()
      setUploading(false)
      setUploadError(`Upload failed: ${error instanceof Error ? error.message : 'Unknown error'}`)
      message.error('Upload failed. Please try again.')
    }
  }

  const pauseUpload = () => {
    uploadRef.current.isPaused = true
    setPaused(true)
    saveProgress() // Save progress on pause
  }

  const resumeUpload = () => {
    uploadRef.current.resume()
  }

  const cancelUpload = () => {
    saveProgress() // Save progress before cancelling
    uploadRef.current.abort()
    setUploading(false)
    stopProgressSaveInterval()
  }

  return (
    <Drawer
      title="Import Contacts from CSV"
      placement="right"
      onClose={handleCloseDrawer}
      open={isVisible}
      width="90%"
      maskClosable={false}
      styles={{
        body: {
          padding: '24px'
        }
      }}
      extra={
        <Space>
          {!uploading && !uploadComplete && csvData && (
            <Button type="primary" onClick={startUpload}>
              {currentRow > 0 ? 'Resume Upload' : 'Start Upload'}
            </Button>
          )}
          {uploading && !paused && (
            <Button icon={<PauseCircleOutlined />} onClick={pauseUpload}>
              Pause
            </Button>
          )}
          {uploading && paused && (
            <Button icon={<PlayCircleOutlined />} onClick={resumeUpload}>
              Resume
            </Button>
          )}
          {uploading && (
            <Button danger icon={<CloseCircleOutlined />} onClick={cancelUpload}>
              Cancel
            </Button>
          )}
          {uploadComplete && (
            <Button type="primary" onClick={onClose}>
              Close
            </Button>
          )}
        </Space>
      }
      footer={null}
    >
      {!csvData && !uploading && !uploadComplete && (
        <Dragger {...uploadProps}>
          <p className="ant-upload-drag-icon">
            <UploadOutlined />
          </p>
          <p className="ant-upload-text">Click or drag a CSV file to this area to upload</p>
          <p className="ant-upload-hint">The CSV file should have headers in the first row.</p>
        </Dragger>
      )}

      {savedProgressExists && csvData && !uploading && !uploadComplete && (
        <Alert
          description={`You're continuing a previous upload session of "${fileName}". The upload will resume from row ${currentRow + 1} of ${totalRows}.`}
          type="info"
          showIcon
          style={{ marginBottom: 24 }}
          action={
            <Button size="small" danger onClick={clearSavedProgress}>
              Start Fresh
            </Button>
          }
        />
      )}

      {uploading && (
        <div style={{ marginTop: 24, marginBottom: 24 }}>
          <Progress percent={uploadProgress} status={paused ? 'exception' : undefined} />
          <div style={{ marginTop: 12, textAlign: 'center' }}>
            <Text>
              Processing batch {currentBatch} of {totalBatches} ({currentRow.toLocaleString()} of{' '}
              {totalRows.toLocaleString()} rows)
              {paused && ' (Paused)'}
            </Text>
          </div>
        </div>
      )}

      {uploadError && (
        <Alert
          message="Upload Error"
          description={uploadError}
          type="error"
          style={{ marginTop: 16, marginBottom: 24 }}
        />
      )}

      {uploadComplete && (
        <div style={{ marginTop: 24, marginBottom: 24 }}>
          <Alert
            message="Upload Complete"
            description={`Processed ${totalRows.toLocaleString()} contacts: ${successCount.toLocaleString()} successful, ${failureCount.toLocaleString()} failed`}
            type={failureCount === 0 ? 'success' : 'warning'}
            showIcon
            style={{ marginBottom: 24 }}
          />

          {failureCount > 0 && (
            <>
              <Typography.Title level={4}>
                Errors ({Math.min(failureCount, 100)} of {failureCount.toLocaleString()} shown)
              </Typography.Title>
              <div style={{ overflowY: 'auto', maxHeight: '400px' }}>
                <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                  <thead>
                    <tr>
                      <th
                        style={{
                          border: '1px solid #f0f0f0',
                          padding: '8px',
                          background: '#fafafa',
                          textAlign: 'left'
                        }}
                      >
                        Line
                      </th>
                      <th
                        style={{
                          border: '1px solid #f0f0f0',
                          padding: '8px',
                          background: '#fafafa',
                          textAlign: 'left'
                        }}
                      >
                        Email
                      </th>
                      <th
                        style={{
                          border: '1px solid #f0f0f0',
                          padding: '8px',
                          background: '#fafafa',
                          textAlign: 'left'
                        }}
                      >
                        Error
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {errorDetails.map((error, index) => (
                      <tr key={index}>
                        <td style={{ border: '1px solid #f0f0f0', padding: '8px' }}>
                          {error.line}
                        </td>
                        <td style={{ border: '1px solid #f0f0f0', padding: '8px' }}>
                          {error.email}
                        </td>
                        <td style={{ border: '1px solid #f0f0f0', padding: '8px' }}>
                          {error.error}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </>
          )}
        </div>
      )}

      {csvData && !uploading && !uploadComplete && (
        <Form
          form={form}
          layout="horizontal"
          labelCol={{ span: 6 }}
          wrapperCol={{ span: 18 }}
          labelAlign="left"
          className="csv-mapping-form"
          style={
            {
              '--form-item-margin-bottom': '12px',
              '--label-font-weight': '500'
            } as React.CSSProperties
          }
        >
          {fileName && (
            <div style={{ marginBottom: 24 }}>
              <Text strong>File: </Text>
              <Text>{fileName}</Text>
              <Text style={{ marginLeft: 8 }}>({csvData.rows.length.toLocaleString()} rows)</Text>
              {currentRow > 0 && (
                <Text type="success" style={{ marginLeft: 8 }}>
                  Will resume from row {currentRow.toLocaleString()}
                </Text>
              )}
            </div>
          )}

          {lists && lists.length > 0 && (
            <div
              style={{
                background: '#f8f8f8',
                padding: '16px',
                borderRadius: '4px',
                marginBottom: '24px'
              }}
            >
              <Form.Item
                name="selectedListIds"
                label={
                  <span>
                    <UserAddOutlined /> Add to Lists
                  </span>
                }
                help="Contacts will be added to these lists on import (optional)"
                initialValue={selectedListIds}
              >
                <Select
                  mode="multiple"
                  placeholder="Select lists to add contacts to"
                  style={{ width: '100%' }}
                  allowClear
                  onChange={(values) => setSelectedListIds(values)}
                  optionFilterProp="children"
                  tagRender={(props) => {
                    const { label, closable, onClose } = props
                    return (
                      <Tag
                        color="blue"
                        closable={closable}
                        onClose={onClose}
                        style={{ marginRight: 3 }}
                      >
                        {label}
                      </Tag>
                    )
                  }}
                >
                  {lists.map((list) => (
                    <Option key={list.id} value={list.id}>
                      {list.name}
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </div>
          )}

          <div
            style={{
              background: '#f8f8f8',
              padding: '16px',
              borderRadius: '4px',
              marginBottom: '12px'
            }}
          >
            <p style={{ marginBottom: '16px' }}>
              <Text>
                Map your CSV columns to contact fields. The <Text strong>Email</Text> field is
                required.
              </Text>
            </p>

            {csvData.headers.map((header, headerIndex) => {
              // Check if this header is currently mapped to the email field
              const mappings = form.getFieldValue('mappings') || {}
              const isEmailMapped = mappings.email === header

              // Find contact field this header is mapped to (if any)
              const mappedToField = Object.entries(mappings).find(
                ([_, value]) => value === header
              )?.[0]

              // Get up to 5 sample values for this column
              const sampleValues = csvData.preview
                .slice(0, 5)
                .map((row) => row[headerIndex])
                .filter((val) => val !== undefined && val !== null && val !== '')

              return (
                <div key={header} style={{ marginBottom: '16px' }}>
                  <Form.Item
                    label={
                      <span>
                        <Text strong>{header}</Text>
                        {isEmailMapped && (
                          <Tag color="red" style={{ marginLeft: 8 }}>
                            Email (Required)
                          </Tag>
                        )}
                      </span>
                    }
                    style={{ marginBottom: '8px' }}
                  >
                    <div
                      style={{ display: 'flex', flexDirection: 'row', alignItems: 'flex-start' }}
                    >
                      <Select
                        placeholder="Select field to map to"
                        value={mappedToField}
                        style={{ width: '200px', marginRight: '12px' }}
                        status={header === mappings.email ? '' : ''}
                        onChange={(value) => {
                          const currentMappings = { ...mappings }

                          // Remove this column from any existing mappings
                          Object.keys(currentMappings).forEach((key) => {
                            if (currentMappings[key] === header) {
                              delete currentMappings[key]
                            }
                          })

                          // Add the new mapping if a field is selected
                          if (value) {
                            currentMappings[value] = header
                          }

                          form.setFieldsValue({ mappings: currentMappings })
                        }}
                        allowClear
                      >
                        <Select.OptGroup label="Required Fields">
                          <Option
                            key="email"
                            value="email"
                            disabled={mappings.email && mappings.email !== header}
                          >
                            Email
                          </Option>
                        </Select.OptGroup>

                        <Select.OptGroup label="Basic Information">
                          {contactFields
                            .filter(
                              (field) => field.key !== 'email' && !field.key.startsWith('custom_')
                            )
                            .map((field) => (
                              <Option
                                key={field.key}
                                value={field.key}
                                disabled={mappings[field.key] && mappings[field.key] !== header}
                              >
                                {field.label}
                              </Option>
                            ))}
                        </Select.OptGroup>

                        <Select.OptGroup label="Custom Fields">
                          {contactFields
                            .filter((field) => field.key.startsWith('custom_'))
                            .map((field) => (
                              <Option
                                key={field.key}
                                value={field.key}
                                disabled={mappings[field.key] && mappings[field.key] !== header}
                              >
                                {field.label}
                              </Option>
                            ))}
                        </Select.OptGroup>
                      </Select>

                      <div
                        style={{
                          minWidth: '300px',
                          background: 'white',
                          border: '1px solid #f0f0f0',
                          borderRadius: '4px',
                          padding: '4px 8px'
                        }}
                      >
                        {sampleValues.length > 0 ? (
                          <>
                            <Text
                              type="secondary"
                              style={{ fontSize: '12px', display: 'block', marginBottom: '4px' }}
                            >
                              Sample values:
                            </Text>
                            {sampleValues.map((value, i) => (
                              <div
                                key={i}
                                style={{
                                  fontSize: '13px',
                                  color: '#333',
                                  whiteSpace: 'nowrap',
                                  overflow: 'hidden',
                                  textOverflow: 'ellipsis',
                                  marginBottom: '2px'
                                }}
                              >
                                {String(value).substring(0, 40)}
                                {String(value).length > 40 ? '...' : ''}
                              </div>
                            ))}
                          </>
                        ) : (
                          <Text type="secondary" style={{ fontSize: '12px' }}>
                            No sample values available
                          </Text>
                        )}
                      </div>
                    </div>
                  </Form.Item>
                </div>
              )
            })}

            <Form.Item
              hidden
              name={['mappings', 'email']}
              rules={[{ required: true, message: 'Email mapping is required' }]}
            >
              <input type="hidden" />
            </Form.Item>
          </div>
        </Form>
      )}
    </Drawer>
  )
}
