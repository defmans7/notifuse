import { Alert, Button, Form, Input, Modal, message } from 'antd'
import { useState } from 'react'
import { useForm } from 'antd/lib/form/Form'
import { FileManagerSettings } from '../../services/api/types'
import { ListObjectsV2Command, ListObjectsV2CommandInput, S3Client } from '@aws-sdk/client-s3'

interface ButtonFilesSettingsProps {
  children: React.ReactNode
  settings?: FileManagerSettings
  onUpdateSettings: (settings: FileManagerSettings) => Promise<void>
}

const ButtonFilesSettings = (props: ButtonFilesSettingsProps) => {
  const [loading, setLoading] = useState(false)
  const [form] = useForm()
  const [settingsVisible, setSettingsVisible] = useState(false)

  const toggleSettings = () => {
    setSettingsVisible(!settingsVisible)
  }

  const onFinish = () => {
    form
      .validateFields()
      .then((values: any) => {
        if (loading) return

        setLoading(true)

        // check if the bucket can be reached
        const input: ListObjectsV2CommandInput = {
          Bucket: values.bucket || ''
        }
        const command = new ListObjectsV2Command(input)

        const s3Client = new S3Client({
          endpoint: values.endpoint || '',
          credentials: {
            accessKeyId: values.access_key || '',
            secretAccessKey: values.secret_key || ''
          },
          region: values.region || 'us-east-1'
        })

        if (values.region === '') {
          delete values.region
        }
        if (values.cdn_endpoint === '') {
          delete values.cdn_endpoint
        }

        s3Client
          .send(command)
          .then(() => {
            // console.log('response', response)

            props
              .onUpdateSettings(values)
              .then(() => {
                message.success('The workspace settings have been updated!')
                setLoading(false)
                toggleSettings()
              })
              .catch((error) => {
                console.error(error)
                message.error(error.toString())
                setLoading(false)
              })
          })
          .catch((e: any) => {
            console.error(e)
            message.error(e.toString())
            setLoading(false)
          })
      })
      .catch((e: any) => {
        console.error(e)
        message.error(e.toString())
        setLoading(false)
      })
  }

  return (
    <span>
      <span onClick={toggleSettings}>{props.children}</span>
      <Modal
        title="File storage settings"
        open={settingsVisible}
        onCancel={toggleSettings}
        footer={[
          <Button key="cancel" loading={loading} onClick={toggleSettings}>
            Cancel
          </Button>,
          <Button key="submit" loading={loading} type="primary" onClick={onFinish}>
            Save
          </Button>
        ]}
      >
        <Form
          form={form}
          layout="horizontal"
          initialValues={props.settings}
          labelCol={{ span: 6 }}
          wrapperCol={{ span: 18 }}
          style={{ marginTop: 24, marginBottom: 40 }}
          onFinish={onFinish}
        >
          <Alert
            message="Your files can be uploaded to any S3 compatible storage."
            type="info"
            showIcon
            style={{ marginBottom: 16 }}
          />

          <Form.Item label="S3 Endpoint" name="endpoint" rules={[{ type: 'url', required: true }]}>
            <Input placeholder="https://storage.googleapis.com" />
          </Form.Item>
          <Form.Item
            label="S3 access key"
            name="access_key"
            rules={[{ type: 'string', required: true }]}
          >
            <Input />
          </Form.Item>
          <Form.Item
            label="S3 secret key"
            name="secret_key"
            rules={[{ type: 'string', required: true }]}
          >
            <Input type="password" />
          </Form.Item>
          <Form.Item label="S3 bucket" name="bucket" rules={[{ type: 'string', required: false }]}>
            <Input />
          </Form.Item>
          <Form.Item label="S3 region" name="region" rules={[{ type: 'string', required: false }]}>
            <Input placeholder="us-east-1" />
          </Form.Item>
          <Form.Item
            label="CDN endpoint"
            name="cdn_endpoint"
            help="URL of the CDN that caches your files"
            rules={[{ type: 'url', required: false }]}
          >
            <Input placeholder="https://cdn.yourbusiness.com" />
          </Form.Item>
        </Form>
      </Modal>
    </span>
  )
}

export default ButtonFilesSettings
