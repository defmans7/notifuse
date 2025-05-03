import { useQuery } from '@tanstack/react-query'
import {
  Card,
  Row,
  Col,
  Statistic,
  Tag,
  Typography,
  Space,
  Tooltip,
  Descriptions,
  Button,
  Divider,
  Modal,
  Input,
  message
} from 'antd'
import { useParams } from '@tanstack/react-router'
import { listsApi } from '../services/api/list'
import { templatesApi } from '../services/api/template'
import type { List, TemplateReference } from '../services/api/types'
import { CreateListDrawer } from '../components/lists/ListDrawer'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import {
  faCircleCheck,
  faPenToSquare,
  faTrashCan,
  faFaceFrown
} from '@fortawesome/free-regular-svg-icons'
import { faUserPlus, faBan, faTriangleExclamation } from '@fortawesome/free-solid-svg-icons'
import { Check, X } from 'lucide-react'
import TemplatePreviewDrawer from '../components/templates/TemplatePreviewDrawer'
import { CreateTemplateDrawer } from '../components/templates/CreateTemplateDrawer'
import { useAuth } from '../contexts/AuthContext'
import { useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { ImportContactsToListButton } from '../components/lists/ImportContactsToListButton'

const { Title, Paragraph, Text } = Typography

// Component to fetch template data and render the preview popover
const TemplatePreviewButton = ({
  templateRef,
  workspaceId
}: {
  templateRef: TemplateReference
  workspaceId: string
}) => {
  const { workspaces } = useAuth()
  const workspace = workspaces.find((w) => w.id === workspaceId)

  const { data, isLoading } = useQuery({
    queryKey: ['template', workspaceId, templateRef.id, templateRef.version],
    queryFn: async () => {
      const response = await templatesApi.get({
        workspace_id: workspaceId,
        id: templateRef.id,
        version: templateRef.version
      })
      return response.template
    },
    enabled: !!templateRef && !!workspaceId,
    // No need to refetch often - template won't change
    staleTime: 1000 * 60 * 5 // 5 minutes
  })

  if (isLoading || !data) {
    return (
      <Button type="link" size="small" loading={isLoading}>
        preview
      </Button>
    )
  }

  return (
    <Space>
      <TemplatePreviewDrawer record={data} workspaceId={workspaceId}>
        <Button type="link" size="small">
          preview
        </Button>
      </TemplatePreviewDrawer>
      {workspace && (
        <CreateTemplateDrawer
          template={data}
          workspace={workspace}
          buttonContent="edit"
          buttonProps={{ type: 'link', size: 'small' }}
        />
      )}
    </Space>
  )
}

export function ListsPage() {
  const { workspaceId } = useParams({ from: '/workspace/$workspaceId/lists' })
  const [deleteModalVisible, setDeleteModalVisible] = useState(false)
  const [listToDelete, setListToDelete] = useState<List | null>(null)
  const [confirmationInput, setConfirmationInput] = useState('')
  const [isDeleting, setIsDeleting] = useState(false)
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['lists', workspaceId],
    queryFn: () => {
      return listsApi.list({ workspace_id: workspaceId })
    }
  })

  const handleDelete = async () => {
    if (!listToDelete) return

    setIsDeleting(true)
    try {
      await listsApi.delete({
        workspace_id: workspaceId,
        id: listToDelete.id
      })

      message.success(`List "${listToDelete.name}" deleted successfully`)
      queryClient.invalidateQueries({ queryKey: ['lists', workspaceId] })
      setDeleteModalVisible(false)
      setListToDelete(null)
      setConfirmationInput('')
    } catch (error) {
      message.error('Failed to delete list')
      console.error(error)
    } finally {
      setIsDeleting(false)
    }
  }

  const openDeleteModal = (list: List) => {
    setListToDelete(list)
    setDeleteModalVisible(true)
  }

  const closeDeleteModal = () => {
    setDeleteModalVisible(false)
    setListToDelete(null)
    setConfirmationInput('')
  }

  const hasLists = !isLoading && data?.lists && data.lists.length > 0

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <div className="text-2xl font-medium">Lists</div>
        {(isLoading || hasLists) && <CreateListDrawer workspaceId={workspaceId} />}
      </div>

      {isLoading ? (
        <Row gutter={[16, 16]}>
          {[1, 2, 3].map((key) => (
            <Col xs={24} sm={12} lg={8} key={key}>
              <Card loading variant="outlined" />
            </Col>
          ))}
        </Row>
      ) : hasLists ? (
        <Space direction="vertical" size="large">
          {data.lists.map((list: List) => (
            <Card
              title={
                <div className="flex items-center justify-between">
                  <Text strong>{list.name}</Text>
                </div>
              }
              extra={
                <Space>
                  <Button type="text" size="small" onClick={() => openDeleteModal(list)}>
                    <Tooltip title="Delete List">
                      <FontAwesomeIcon icon={faTrashCan} style={{ opacity: 0.7 }} />
                    </Tooltip>
                  </Button>
                  <CreateListDrawer
                    workspaceId={workspaceId}
                    list={list}
                    buttonProps={{
                      type: 'text',
                      size: 'small',
                      buttonContent: (
                        <Tooltip title="Edit List">
                          <FontAwesomeIcon icon={faPenToSquare} style={{ opacity: 0.7 }} />
                        </Tooltip>
                      )
                    }}
                  />
                  <ImportContactsToListButton
                    list={list}
                    workspaceId={workspaceId}
                    lists={data.lists}
                  />
                </Space>
              }
              key={list.id}
            >
              <Row gutter={[16, 16]} wrap={false}>
                <Col flex="1">
                  <Statistic
                    title={
                      <Space>
                        <FontAwesomeIcon
                          icon={faCircleCheck}
                          className="text-green-500"
                          style={{ opacity: 0.7 }}
                        />{' '}
                        Active
                      </Space>
                    }
                    value={list.total_active}
                    valueStyle={{ fontSize: '16px' }}
                  />
                </Col>
                <Col flex="1">
                  <Statistic
                    title={
                      <Space>
                        <FontAwesomeIcon
                          icon={faUserPlus}
                          className="text-blue-500"
                          style={{ opacity: 0.7 }}
                        />{' '}
                        Pending
                      </Space>
                    }
                    value={list.total_pending}
                    valueStyle={{ fontSize: '16px' }}
                  />
                </Col>
                <Col flex="1">
                  <Statistic
                    title={
                      <Space>
                        <FontAwesomeIcon
                          icon={faBan}
                          className="text-gray-500"
                          style={{ opacity: 0.7 }}
                        />{' '}
                        Unsub
                      </Space>
                    }
                    value={list.total_unsubscribed}
                    valueStyle={{ fontSize: '16px' }}
                  />
                </Col>
                <Col flex="1">
                  <Statistic
                    title={
                      <Space>
                        <FontAwesomeIcon
                          icon={faTriangleExclamation}
                          className="text-yellow-500"
                          style={{ opacity: 0.7 }}
                        />{' '}
                        Bounced
                      </Space>
                    }
                    value={list.total_bounced}
                    valueStyle={{ fontSize: '16px' }}
                  />
                </Col>
                <Col flex="1">
                  <Statistic
                    title={
                      <Space>
                        <FontAwesomeIcon
                          icon={faFaceFrown}
                          className="text-red-500"
                          style={{ opacity: 0.7 }}
                        />{' '}
                        Complaints
                      </Space>
                    }
                    value={list.total_complained}
                    valueStyle={{ fontSize: '16px' }}
                  />
                </Col>
              </Row>

              <Divider />

              <Descriptions size="small" column={2}>
                <Descriptions.Item label="ID">{list.id}</Descriptions.Item>

                <Descriptions.Item label="Description">{list.description}</Descriptions.Item>
                <Descriptions.Item label="Visibility">
                  {list.is_public ? (
                    <Tag color="green">Public</Tag>
                  ) : (
                    <Tag color="volcano">Private</Tag>
                  )}
                </Descriptions.Item>

                {/* Double Opt-in Template */}
                <Descriptions.Item label="Double Opt-in Template">
                  {list.double_optin_template ? (
                    <Space>
                      <Check size={16} className="text-green-500" />
                      <TemplatePreviewButton
                        templateRef={list.double_optin_template}
                        workspaceId={workspaceId}
                      />
                    </Space>
                  ) : (
                    <X size={16} className="text-red-500" />
                  )}
                </Descriptions.Item>

                {/* Welcome Template */}
                <Descriptions.Item label="Welcome Template">
                  {list.welcome_template ? (
                    <Space>
                      <Check size={16} className="text-green-500" />
                      <TemplatePreviewButton
                        templateRef={list.welcome_template}
                        workspaceId={workspaceId}
                      />
                    </Space>
                  ) : (
                    <X size={16} className="text-red-500" />
                  )}
                </Descriptions.Item>

                {/* Unsubscribe Template */}
                <Descriptions.Item label="Unsubscribe Template">
                  {list.unsubscribe_template ? (
                    <Space>
                      <Check size={16} className="text-green-500" />
                      <TemplatePreviewButton
                        templateRef={list.unsubscribe_template}
                        workspaceId={workspaceId}
                      />
                    </Space>
                  ) : (
                    <X size={16} className="text-red-500" />
                  )}
                </Descriptions.Item>
              </Descriptions>
            </Card>
          ))}
        </Space>
      ) : (
        <div className="text-center py-12">
          <Title level={4} type="secondary">
            No lists found
          </Title>
          <Paragraph type="secondary">Create your first list to get started</Paragraph>
          <div className="mt-4">
            <CreateListDrawer workspaceId={workspaceId} buttonProps={{ size: 'large' }} />
          </div>
        </div>
      )}

      <Modal
        title="Delete List"
        open={deleteModalVisible}
        onCancel={closeDeleteModal}
        footer={[
          <Button key="cancel" onClick={closeDeleteModal}>
            Cancel
          </Button>,
          <Button
            key="delete"
            type="primary"
            danger
            loading={isDeleting}
            disabled={confirmationInput !== (listToDelete?.id || '')}
            onClick={handleDelete}
          >
            Delete
          </Button>
        ]}
      >
        {listToDelete && (
          <>
            <p>Are you sure you want to delete the list "{listToDelete.name}"?</p>
            <p>
              This action cannot be undone. To confirm, please enter the list ID:{' '}
              <Text code>{listToDelete.id}</Text>
            </p>
            <Input
              placeholder="Enter list ID to confirm"
              value={confirmationInput}
              onChange={(e) => setConfirmationInput(e.target.value)}
              status={confirmationInput && confirmationInput !== listToDelete.id ? 'error' : ''}
            />
            {confirmationInput && confirmationInput !== listToDelete.id && (
              <p className="text-red-500 mt-2">ID doesn't match</p>
            )}
          </>
        )}
      </Modal>
    </div>
  )
}
