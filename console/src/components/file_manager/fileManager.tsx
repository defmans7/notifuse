import React from 'react'
import {
  Alert,
  Button,
  Form,
  Input,
  Modal,
  Popconfirm,
  Popover,
  Space,
  Table,
  Tooltip,
  message
} from 'antd'
import { FileManagerProps, StorageObject } from './interfaces'
import { ChangeEvent, useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Copy, Folder, Trash2, ExternalLink, Settings, RefreshCw, Plus } from 'lucide-react'
import dayjs from '../../lib/dayjs'
import { filesize } from 'filesize'
import ButtonFilesSettings from './buttonSettings'
import {
  S3Client,
  ListObjectsV2Command,
  ListObjectsV2CommandInput,
  PutObjectCommand,
  PutObjectCommandInput,
  DeleteObjectCommand,
  DeleteObjectCommandInput
} from '@aws-sdk/client-s3'
import GetContentType from './fileExtensions'
import { FileManagerSettings } from '../../services/api/types'
// Common styles
const styles = {
  folderRow: {
    fontWeight: 'bold' as const,
    cursor: 'pointer'
  },
  filesContainer: {
    position: 'relative' as const,
    overflow: 'auto' as const,
    paddingBottom: '40px'
  },
  marginBottomSmall: { marginBottom: 16 },
  marginBottomLarge: { marginBottom: 24 },
  padding: { paddingBottom: 16 },
  pullRight: { float: 'right' as const },
  paddingRightSmall: { paddingRight: 8 },
  textRight: { textAlign: 'right' as const },
  primary: { color: '#1890ff' } // Default antd primary color - replace with actual color if different
}

// Extend FileManagerProps to include settings and onUpdateSettings
export interface ExtendedFileManagerProps extends FileManagerProps {
  settings?: FileManagerSettings
  onUpdateSettings: (settings: FileManagerSettings) => Promise<void>
}

export const FileManager = (props: ExtendedFileManagerProps) => {
  const [currentPath, setCurrentPath] = useState(props.currentPath || '')
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([])
  const [items, setItems] = useState<StorageObject[] | undefined>(undefined)
  const [isLoading, setIsLoading] = useState(false)
  const [newFolderModalVisible, setNewFolderModalVisible] = useState(false)
  const [newFolderLoading, setNewFolderLoading] = useState(false)
  const s3ClientRef = useRef<S3Client | undefined>(undefined)
  const inputFileRef = useRef<HTMLInputElement>(null)
  const [isUploading, setIsUploading] = useState(false)
  const [form] = Form.useForm()

  const goToPath = (path: string) => {
    // reset selection on path change
    setSelectedRowKeys([])
    props.onSelect([])
    setCurrentPath(path)
  }

  const fetchObjects = useCallback(() => {
    if (!s3ClientRef.current) return

    setIsLoading(true)
    const input: ListObjectsV2CommandInput = {
      Bucket: props.settings?.bucket || ''
    }

    const command = new ListObjectsV2Command(input)
    s3ClientRef.current.send(command).then((response) => {
      // console.log('response', response)
      if (!response.Contents) {
        setItems([])
        setIsLoading(false)
        return
      }

      const newItems = response.Contents.map((x) => {
        const key = x.Key as string
        let endpoint = (props.settings?.endpoint || '') + '/' + (props.settings?.bucket || '')

        if (props.settings?.cdn_endpoint !== '') {
          endpoint = props.settings?.cdn_endpoint || ''
        }

        const isFolder = key.endsWith('/')
        let name =
          key
            .split('/')
            .filter((x) => x !== '')
            .pop() || ''

        if (!isFolder) {
          name = key.split('/').pop() || ''
        }

        // console.log('item', x)

        let itemPath = ''
        const pathParts = key.split('/')

        if (isFolder) {
          itemPath = pathParts.slice(0, pathParts.length - 2).join('/') + '/'
          // console.log('folder path', itemCurrentPath)
        } else {
          itemPath = pathParts.slice(0, pathParts.length - 1).join('/') + '/'
          // console.log('file path', itemCurrentPath)
        }

        if (itemPath === '/') itemPath = ''

        const item = {
          key: key,
          name: name,
          path: itemPath,
          is_folder: isFolder,
          last_modified: x.LastModified
        } as StorageObject

        if (!isFolder) {
          item.file_info = {
            size: x.Size as number,
            size_human: filesize(x.Size || 0, { round: 0 }),
            content_type: GetContentType(key),
            url: endpoint + '/' + key
          }
        }

        return item
      })

      // console.log('new items', newItems)
      setItems(newItems)
      setIsLoading(false)
    })
  }, [props.settings?.bucket, props.settings?.cdn_endpoint, props.settings?.endpoint])

  // init
  useEffect(() => {
    if (props.settings?.endpoint === '') {
      return
    }
    if (s3ClientRef.current) return

    s3ClientRef.current = new S3Client({
      endpoint: props.settings?.endpoint || '',
      credentials: {
        accessKeyId: props.settings?.access_key || '',
        secretAccessKey: props.settings?.secret_key || ''
      },
      region: props.settings?.region || 'us-east-1'
    })

    fetchObjects()
  }, [
    props.settings?.endpoint,
    props.settings?.access_key,
    props.settings?.secret_key,
    props.settings?.region,
    fetchObjects
  ])

  const deleteObject = (key: string, isFolder: boolean) => {
    if (!s3ClientRef.current) {
      message.error('S3 client is not initialized.')
      return
    }

    const s3Client = s3ClientRef.current

    const input: DeleteObjectCommandInput = {
      Bucket: props.settings?.bucket || '',
      Key: key
    }

    s3Client
      .send(new DeleteObjectCommand(input))
      .then(() => {
        if (isFolder) {
          fetchObjects()
          message.success('Folder deleted successfully.')
          // go to previous path
          setCurrentPath(key.split('/').slice(0, -2).join('/') + '/')
        } else {
          message.success('File deleted successfully.')
        }
        // refresh
        fetchObjects()
      })
      .catch((error) => {
        message.error('Failed to delete file: ' + error)
        props.onError(error)
      })
  }

  const selectItem = (items: StorageObject[]) => {
    console.log('selected items', items)
  }

  const toggleSelectionForItem = (item: StorageObject) => {
    // ignore items not accepted
    if (!props.acceptItem(item)) return

    if (props.multiple) {
      let newKeys = [...selectedRowKeys]
      // remove if exists
      if (newKeys.includes(item.key)) {
        newKeys = selectedRowKeys.filter((k) => k !== item.key)
      } else {
        newKeys.push(item.key)
      }
      setSelectedRowKeys(newKeys)
      props.onSelect(items ? items.filter((x) => newKeys.includes(x.key)) : [])
    } else {
      setSelectedRowKeys([item.key])
      props.onSelect([item])
    }
  }

  const toggleNewFolderModal = () => {
    setNewFolderModalVisible(!newFolderModalVisible)
  }

  const onSubmitNewFolder = () => {
    if (!s3ClientRef.current) {
      message.error('S3 client is not initialized.')
      return
    }

    if (newFolderLoading) return

    const s3Client = s3ClientRef.current

    form.validateFields().then((values) => {
      setNewFolderLoading(true)

      // create folder in S3
      const folderName = values.name
      const key = currentPath === '' ? folderName + '/' : currentPath + folderName + '/'

      const input: ListObjectsV2CommandInput = {
        Bucket: props.settings?.bucket || '',
        Prefix: key
      }

      s3Client
        .send(new ListObjectsV2Command(input))
        .then((response) => {
          // console.log('response', response)
          if (response.Contents && response.Contents.length > 0) {
            message.error('Folder already exists.')
            return
          }

          const input: PutObjectCommandInput = {
            Bucket: props.settings?.bucket || '',
            Key: key,
            Body: ''
          }

          s3Client
            .send(new PutObjectCommand(input))
            .then(() => {
              message.success('Folder created successfully.')
              setNewFolderLoading(false)
              fetchObjects()
            })
            .catch((error) => {
              message.error('Failed to create folder: ' + error)
              setNewFolderLoading(false)
              props.onError(error)
            })
        })
        .catch((error) => {
          message.error('Failed to create folder: ' + error)
          setNewFolderLoading(false)
          props.onError(error)
        })

      form.resetFields()
      toggleNewFolderModal()
    })
  }

  const itemsAtPath = useMemo(() => {
    if (!items) return []
    return items
      .filter((x) => x.path === currentPath)
      .sort((a, b) => {
        // by folders first, then by last_modified
        if (a.is_folder && !b.is_folder) return -1
        if (!a.is_folder && b.is_folder) return 1
        if (a.last_modified > b.last_modified) return -1
        if (a.last_modified < b.last_modified) return 1
        return 0
      })
  }, [items, currentPath])

  const onFileChange = (e: ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files) return
    if (isUploading) return
    if (!s3ClientRef.current) return

    // console.log(e.target.files)

    for (var i = 0; i < e.target.files.length; i++) {
      setIsUploading(true)
      const file = e.target.files.item(i) as File

      // Convert file to ArrayBuffer for browser compatibility with AWS SDK v3
      file
        .arrayBuffer()
        .then((arrayBuffer) => {
          const uint8Array = new Uint8Array(arrayBuffer)

          s3ClientRef
            .current!.send(
              new PutObjectCommand({
                Bucket: props.settings?.bucket || '',
                Key: currentPath + file.name,
                Body: uint8Array,
                ContentType: file.type
              })
            )
            .then(() => {
              message.success('File ' + file.name + ' uploaded successfully.')
              setIsUploading(false)
              fetchObjects()
            })
            .catch((error) => {
              message.error('Failed to upload file: ' + error)
              setIsUploading(false)
              props.onError(error)
            })
        })
        .catch((error) => {
          message.error('Failed to read file: ' + error)
          setIsUploading(false)
          props.onError(error)
        })
    }
  }

  const onBrowseFiles = () => {
    if (inputFileRef.current) {
      inputFileRef.current.click()
    }
  }

  if (!props.settings?.endpoint) {
    return (
      <Alert
        style={styles.marginBottomSmall}
        message={
          <>
            File storage is not configured.
            <ButtonFilesSettings
              settings={props.settings}
              onUpdateSettings={props.onUpdateSettings}
            >
              <Button type="link">Configure now</Button>
            </ButtonFilesSettings>
          </>
        }
        type="warning"
        showIcon
      />
    )
  }

  return (
    <div style={{ ...styles.filesContainer, height: props.height }}>
      {props.settings?.endpoint !== '' && (
        <>
          <div style={{ ...styles.padding, borderBottom: '1px solid rgba(0,0,0,0.1)' }}>
            <div style={styles.pullRight}>
              <Space>
                {currentPath !== '' && (
                  <Tooltip title="Delete folder" placement="bottom">
                    <Popconfirm
                      placement="topRight"
                      title={
                        <>
                          Do you want to delete the <b>{currentPath}</b> folder with all its
                          content?
                        </>
                      }
                      onConfirm={() => deleteObject(currentPath, true)}
                      okText="Delete folder"
                      cancelText="Cancel"
                      okButtonProps={{
                        danger: true
                      }}
                    >
                      <Button
                        size="small"
                        type="text"
                        onClick={() => fetchObjects()}
                        icon={<Trash2 size={16} />}
                      />
                    </Popconfirm>
                  </Tooltip>
                )}
                <Tooltip title="Refresh the list">
                  <Button
                    size="small"
                    type="text"
                    onClick={() => fetchObjects()}
                    icon={<RefreshCw size={16} />}
                  />
                </Tooltip>

                <ButtonFilesSettings
                  settings={props.settings}
                  onUpdateSettings={props.onUpdateSettings}
                >
                  <Tooltip title="Storage settings">
                    <Button type="text" size="small">
                      <Settings size={16} />
                    </Button>
                  </Tooltip>
                </ButtonFilesSettings>
                <span role="button" onClick={onBrowseFiles}>
                  <input
                    type="file"
                    ref={inputFileRef}
                    onChange={onFileChange}
                    hidden
                    accept={props.acceptFileType}
                    multiple={false}
                  />
                  <Button
                    type="primary"
                    // size="small"
                    style={styles.pullRight}
                    loading={isUploading}
                  >
                    <Plus size={16} />
                    Upload
                  </Button>
                </span>
              </Space>
            </div>

            <Space>
              <div>
                <Button type="text" onClick={() => goToPath('')}>
                  {props.settings?.bucket || ''}
                </Button>
                {currentPath
                  .split('/')
                  .filter((x) => x !== '')
                  .map((part, index, array) => {
                    const isLast = index === array.length - 1
                    const fullPath = array.slice(0, index + 1).join('/') + '/'
                    return (
                      <React.Fragment key={fullPath}>
                        /
                        <Button
                          disabled={isLast}
                          type="text"
                          // size="small"
                          onClick={() => goToPath(fullPath)}
                        >
                          {part}
                        </Button>
                      </React.Fragment>
                    )
                  })}
              </div>
              <Button type="primary" ghost onClick={toggleNewFolderModal}>
                New folder
              </Button>
            </Space>
          </div>
          <Table
            dataSource={itemsAtPath}
            loading={isLoading}
            pagination={false}
            size="middle"
            rowKey="key"
            locale={{ emptyText: 'Folder is empty' }}
            rowClassName={(record: StorageObject) => {
              return record.is_folder ? 'folder-row' : ''
            }}
            onRow={(record: StorageObject) => {
              return {
                onClick: () => {
                  if (record.is_folder) {
                    setCurrentPath(record.key)
                  }
                },
                style: record.is_folder ? styles.folderRow : undefined
              }
            }}
            rowSelection={
              props.withSelection
                ? {
                    type: props.multiple ? 'checkbox' : 'radio',
                    selectedRowKeys: selectedRowKeys,
                    onChange: (selectedRowKeys: React.Key[], selectedRows: any[]) => {
                      setSelectedRowKeys(selectedRowKeys)
                      selectItem(selectedRows)
                    },
                    getCheckboxProps: (record: any) => ({
                      disabled: !props.acceptItem(record as StorageObject)
                    })
                  }
                : undefined
            }
            columns={[
              {
                title: '',
                key: 'preview',
                width: 70,
                render: (item: StorageObject) => {
                  if (item.is_folder) {
                    return (
                      <div onClick={toggleSelectionForItem.bind(null, item)}>
                        <Folder size={16} style={styles.primary} />
                      </div>
                    )
                  }
                  console.log('item', item)
                  return (
                    <div onClick={toggleSelectionForItem.bind(null, item)}>
                      {item.file_info.content_type.includes('image') && (
                        <Popover
                          placement="right"
                          content={
                            <img src={item.file_info.url} alt="" style={{ maxHeight: '400px' }} />
                          }
                        >
                          <img src={item.file_info.url} alt="" height="30" />
                        </Popover>
                      )}
                    </div>
                  )
                }
              },
              {
                title: 'Name',
                key: 'name',
                render: (item: StorageObject) => {
                  return <div onClick={toggleSelectionForItem.bind(null, item)}>{item.name}</div>
                }
              },
              {
                title: 'Size',
                key: 'size',
                width: 100,
                render: (item: StorageObject) => {
                  return (
                    <div onClick={toggleSelectionForItem.bind(null, item)}>
                      {item.is_folder ? '-' : item.file_info.size_human}
                    </div>
                  )
                }
              },
              {
                title: 'Last modified',
                key: 'lastModified',
                width: 120,
                render: (item: StorageObject) => {
                  return (
                    <Tooltip title={dayjs(item.last_modified).format('llll')}>
                      <div onClick={toggleSelectionForItem.bind(null, item)}>
                        {dayjs(item.last_modified).format('ll')}
                      </div>
                    </Tooltip>
                  )
                }
              },
              {
                title: '',
                key: 'actions',
                width: 40,
                align: 'right',
                render: (item: StorageObject) => {
                  if (item.is_folder) return
                  return (
                    <Space>
                      <Tooltip title="Copy URL">
                        <Button
                          type="text"
                          size="small"
                          onClick={() => {
                            navigator.clipboard.writeText(item.file_info.url)
                            message.success('URL copied to clipboard.')
                          }}
                        >
                          <Copy size={16} />
                        </Button>
                      </Tooltip>
                      <Tooltip title="Open in a window">
                        <a href={item.file_info.url} target="_blank" rel="noreferrer">
                          <Button type="text" size="small">
                            <ExternalLink size={16} />
                          </Button>
                        </a>
                      </Tooltip>
                      <Popconfirm
                        title="Do you want to permanently delete this file from your storage?"
                        onConfirm={() => deleteObject(item.key, false)}
                        placement="topRight"
                        okText="Delete"
                        cancelText="Cancel"
                        okButtonProps={{
                          danger: true
                        }}
                      >
                        <Button type="text" size="small">
                          <Trash2 size={16} />
                        </Button>
                      </Popconfirm>
                    </Space>
                  )
                }
              }
            ]}
          />
        </>
      )}
      {newFolderModalVisible && (
        <Modal
          title="Create new folder"
          open={newFolderModalVisible}
          onOk={onSubmitNewFolder}
          onCancel={toggleNewFolderModal}
          confirmLoading={newFolderLoading}
        >
          <Form form={form}>
            <Form.Item
              label="Folder name"
              name="name"
              rules={[
                {
                  required: true,
                  type: 'string',
                  validator(_rule, value, callback) {
                    // alphanumeric, lowercase, underscore, dash
                    if (!/^[a-z0-9_-]+$/.test(value)) {
                      callback(
                        'Only lowercase alphanumeric characters, underscore, and dash are allowed.'
                      )
                      return
                    }
                    callback()
                  }
                }
              ]}
            >
              <Input
                addonBefore={currentPath !== '/' ? currentPath : '/'}
                onChange={(e) => {
                  // trim spaces
                  form.setFieldsValue({ folderName: e.target.value.trim() })
                }}
              />
            </Form.Item>
          </Form>
        </Modal>
      )}
    </div>
  )
}
