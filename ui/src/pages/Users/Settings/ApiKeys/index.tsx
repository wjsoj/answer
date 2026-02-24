/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import React, { useState, useEffect } from 'react';
import { Button, Table, Alert, Modal, Form, Card } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import { useToast } from '@/hooks';
import { getUserApiKeys, createUserApiKey, deleteUserApiKey } from '@/services';

interface ApiKey {
  id: number;
  name: string;
  access_key: string;
  created_at: number;
  last_used_at: number;
  expires_at?: number;
  usage_count: number;
}

const ApiKeys: React.FC = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'settings.api_keys',
  });
  const toast = useToast();
  const [keys, setKeys] = useState<ApiKey[]>([]);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newKey, setNewKey] = useState<string | null>(null);
  const [formData, setFormData] = useState({ name: '', description: '' });
  const [loading, setLoading] = useState(false);

  const loadKeys = async () => {
    try {
      const data = await getUserApiKeys();
      setKeys(data || []);
    } catch (error: any) {
      toast.onShow({
        msg: error?.msg || 'Failed to load API keys',
        variant: 'danger',
      });
    }
  };

  useEffect(() => {
    loadKeys();
  }, []);

  const handleCreate = async () => {
    if (!formData.name.trim()) {
      toast.onShow({
        msg: t('name_required'),
        variant: 'warning',
      });
      return;
    }

    setLoading(true);
    try {
      const result = await createUserApiKey(formData);
      setNewKey(result.access_key);
      setShowCreateModal(false);
      setFormData({ name: '', description: '' });
      loadKeys();
    } catch (error: any) {
      toast.onShow({
        msg: error?.msg || 'Failed to create API key',
        variant: 'danger',
      });
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: number, name: string) => {
    if (!window.confirm(t('confirm_delete', { name }))) {
      return;
    }

    try {
      await deleteUserApiKey(id);
      toast.onShow({
        msg: t('delete_success'),
        variant: 'success',
      });
      loadKeys();
    } catch (error: any) {
      toast.onShow({
        msg: error?.msg || 'Failed to delete API key',
        variant: 'danger',
      });
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text).then(() => {
      toast.onShow({ msg: t('copied'), variant: 'success' });
    });
  };

  const maskKey = (key: string) => {
    if (key.length <= 8) return key;
    return `${key.substring(0, 4)}...${key.substring(key.length - 4)}`;
  };

  const formatDate = (timestamp: number) => {
    if (!timestamp) return t('never');
    return new Date(timestamp * 1000).toLocaleString();
  };

  return (
    <div>
      <h3 className="mb-4">{t('title')}</h3>

      <div className="mb-3">
        <p className="text-muted">{t('description')}</p>
      </div>

      <Card className="mb-4 border-secondary">
        <Card.Body>
          <Card.Title className="fs-6 fw-semibold">
            {t('mcp_info_title')}
          </Card.Title>
          <p className="text-muted small mb-2">{t('mcp_info_desc')}</p>
          <div className="d-flex align-items-center gap-2 mb-1">
            <code className="text-body bg-body-secondary px-2 py-1 rounded flex-grow-1">
              {window.location.origin}/answer/api/v1/mcp
            </code>
            <Button
              size="sm"
              variant="outline-secondary"
              onClick={() =>
                copyToClipboard(`${window.location.origin}/answer/api/v1/mcp`)
              }>
              {t('copy')}
            </Button>
          </div>
          <p className="text-muted small mb-0">{t('mcp_auth_desc')}</p>
        </Card.Body>
      </Card>

      {newKey && (
        <Alert variant="success" dismissible onClose={() => setNewKey(null)}>
          <Alert.Heading>{t('created_success')}</Alert.Heading>
          <p>{t('save_key_warning')}</p>
          <div className="d-flex align-items-center gap-2 bg-body-secondary p-2 rounded">
            <code className="text-body flex-grow-1">{newKey}</code>
            <Button
              size="sm"
              variant="outline-secondary"
              onClick={() => copyToClipboard(newKey)}>
              {t('copy')}
            </Button>
          </div>
        </Alert>
      )}

      <div className="d-flex justify-content-between mb-3">
        <div />
        <Button onClick={() => setShowCreateModal(true)}>{t('create')}</Button>
      </div>

      {keys.length === 0 ? (
        <Alert variant="info">{t('no_keys')}</Alert>
      ) : (
        <Table striped bordered hover responsive>
          <thead>
            <tr>
              <th>{t('name')}</th>
              <th>{t('key')}</th>
              <th>{t('created_at')}</th>
              <th>{t('last_used')}</th>
              <th>Usage Count</th>
              <th>Expires</th>
              <th>{t('actions')}</th>
            </tr>
          </thead>
          <tbody>
            {keys.map((key) => (
              <tr key={key.id}>
                <td>{key.name}</td>
                <td>
                  <code>{maskKey(key.access_key)}</code>
                </td>
                <td>{formatDate(key.created_at)}</td>
                <td>{formatDate(key.last_used_at)}</td>
                <td>{key.usage_count || 0}</td>
                <td>
                  {key.expires_at ? (
                    <span
                      className={
                        key.expires_at * 1000 < Date.now() ? 'text-danger' : ''
                      }>
                      {formatDate(key.expires_at)}
                    </span>
                  ) : (
                    <span className="text-muted">{t('never')}</span>
                  )}
                </td>
                <td>
                  <Button
                    variant="outline-danger"
                    size="sm"
                    onClick={() => handleDelete(key.id, key.name)}>
                    {t('delete')}
                  </Button>
                </td>
              </tr>
            ))}
          </tbody>
        </Table>
      )}

      <Modal show={showCreateModal} onHide={() => setShowCreateModal(false)}>
        <Modal.Header closeButton>
          <Modal.Title>{t('create_modal_title')}</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <Form>
            <Form.Group className="mb-3">
              <Form.Label>{t('name')}</Form.Label>
              <Form.Control
                type="text"
                value={formData.name}
                onChange={(e) =>
                  setFormData({ ...formData, name: e.target.value })
                }
                placeholder={t('name_placeholder')}
              />
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Label>{t('desc')}</Form.Label>
              <Form.Control
                as="textarea"
                rows={3}
                value={formData.description}
                onChange={(e) =>
                  setFormData({ ...formData, description: e.target.value })
                }
                placeholder={t('description_placeholder')}
              />
            </Form.Group>
          </Form>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => setShowCreateModal(false)}>
            {t('cancel')}
          </Button>
          <Button variant="primary" onClick={handleCreate} disabled={loading}>
            {loading ? t('creating') : t('create')}
          </Button>
        </Modal.Footer>
      </Modal>
    </div>
  );
};

export default ApiKeys;
