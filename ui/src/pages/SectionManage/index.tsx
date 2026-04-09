import React, { FC, useState, useEffect, useCallback, useRef } from 'react';
import {
  Table,
  Button,
  Form,
  Badge,
  Spinner,
  ListGroup,
} from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import { useForumSections } from '@/services/common';
import { userSearchByName } from '@/services/client/user';
import {
  updateForumSectionVisibility,
  inviteForumSectionUsers,
  getForumSectionPermissions,
  removeForumSectionUsers,
} from '@/services/admin';
import { loggedUserInfoStore } from '@/stores';
import type * as Type from '@/common/interface';

interface SectionPermissions {
  members: string[];
  moderators: string[];
}

const UserSearchInput: FC<{
  selectedUsers: string[];
  onSelect: (username: string) => void;
  onRemove: (username: string) => void;
  placeholder: string;
}> = ({ selectedUsers, onSelect, onRemove, placeholder }) => {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<Type.UserInfoBase[]>([]);
  const [showDropdown, setShowDropdown] = useState(false);
  const [searching, setSearching] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout>>();
  const wrapperRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (
        wrapperRef.current &&
        !wrapperRef.current.contains(e.target as Node)
      ) {
        setShowDropdown(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const doSearch = useCallback(
    (q: string) => {
      if (q.trim().length < 1) {
        setResults([]);
        setShowDropdown(false);
        return;
      }
      setSearching(true);
      userSearchByName(q)
        .then((res) => {
          const list = (res as Type.UserInfoBase[]) || [];
          setResults(list.filter((u) => !selectedUsers.includes(u.username)));
          setShowDropdown(true);
        })
        .catch(() => {
          setResults([]);
        })
        .finally(() => setSearching(false));
    },
    [selectedUsers],
  );

  const handleInputChange = (val: string) => {
    setQuery(val);
    if (timerRef.current) clearTimeout(timerRef.current);
    timerRef.current = setTimeout(() => doSearch(val), 300);
  };

  const handleSelect = (username: string) => {
    onSelect(username);
    setQuery('');
    setResults([]);
    setShowDropdown(false);
  };

  return (
    <div ref={wrapperRef} className="position-relative">
      <div className="d-flex flex-wrap gap-1 mb-2">
        {selectedUsers.map((name) => (
          <Badge
            key={name}
            bg="primary"
            className="d-inline-flex align-items-center">
            {name}
            <Button
              variant="link"
              size="sm"
              className="p-0 ms-1 text-white"
              onClick={() => onRemove(name)}>
              &times;
            </Button>
          </Badge>
        ))}
      </div>
      <Form.Control
        type="text"
        size="sm"
        placeholder={placeholder}
        value={query}
        onChange={(e) => handleInputChange(e.target.value)}
        onFocus={() => {
          if (results.length > 0) setShowDropdown(true);
        }}
      />
      {searching && (
        <div className="position-absolute end-0 top-50 translate-middle-y me-2">
          <Spinner animation="border" size="sm" />
        </div>
      )}
      {showDropdown && results.length > 0 && (
        <ListGroup
          className="position-absolute w-100 shadow-sm"
          style={{ zIndex: 1050, maxHeight: '200px', overflowY: 'auto' }}>
          {results.map((user) => (
            <ListGroup.Item
              key={user.id}
              action
              onClick={() => handleSelect(user.username)}
              className="d-flex align-items-center py-1">
              {user.avatar && (
                <img
                  src={`${user.avatar}?s=24`}
                  alt=""
                  width={24}
                  height={24}
                  className="rounded-circle me-2"
                />
              )}
              <span>
                <strong>{user.display_name}</strong>
                <small className="text-muted ms-1">@{user.username}</small>
              </span>
            </ListGroup.Item>
          ))}
        </ListGroup>
      )}
    </div>
  );
};

const SectionManage: FC = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'section_manage',
  });
  const { user: userInfo } = loggedUserInfoStore();
  const isAdmin = userInfo?.role_id === 2;
  const { data: sectionsData, mutate: mutateSections } = useForumSections();
  const allSections = sectionsData?.list || [];
  const sections = allSections.filter((s) => s.can_manage);

  const [expandedSection, setExpandedSection] = useState<string | null>(null);
  const [permissions, setPermissions] = useState<SectionPermissions | null>(
    null,
  );
  const [loadingPermissions, setLoadingPermissions] = useState(false);
  const [inviteRole, setInviteRole] = useState<'member' | 'moderator'>(
    'member',
  );
  const [selectedUsers, setSelectedUsers] = useState<string[]>([]);
  const [saving, setSaving] = useState(false);
  const [skippedUsers, setSkippedUsers] = useState<string[]>([]);

  const loadPermissions = useCallback(async (slug: string) => {
    setLoadingPermissions(true);
    try {
      const resp = await getForumSectionPermissions(slug);
      setPermissions(resp as SectionPermissions);
    } catch {
      setPermissions({ members: [], moderators: [] });
    }
    setLoadingPermissions(false);
  }, []);

  useEffect(() => {
    if (expandedSection) {
      loadPermissions(expandedSection);
      setSelectedUsers([]);
      setSkippedUsers([]);
      setInviteRole('member');
    }
  }, [expandedSection, loadPermissions]);

  const handleVisibilityChange = async (
    section: Type.ForumSectionItem,
    newVisibility: string,
  ) => {
    setSaving(true);
    try {
      await updateForumSectionVisibility({
        section: section.slug_name,
        visibility: newVisibility,
      });
      mutateSections();
    } catch {
      // error handled by request interceptor
    }
    setSaving(false);
  };

  const handleInvite = async () => {
    if (!expandedSection || selectedUsers.length === 0) return;
    setSaving(true);
    try {
      const resp = await inviteForumSectionUsers({
        section: expandedSection,
        users: selectedUsers,
        role: inviteRole,
      });
      setSelectedUsers([]);
      loadPermissions(expandedSection);
      const skipped = (resp as { skipped_users?: string[] })?.skipped_users;
      setSkippedUsers(skipped || []);
    } catch {
      // error handled by request interceptor
    }
    setSaving(false);
  };

  const handleRemove = async (username: string, role: string) => {
    if (!expandedSection) return;
    setSaving(true);
    try {
      await removeForumSectionUsers({
        section: expandedSection,
        users: [username],
        role,
      });
      loadPermissions(expandedSection);
    } catch {
      // error handled by request interceptor
    }
    setSaving(false);
  };

  if (sections.length === 0) {
    return (
      <div className="py-5 text-center text-muted">
        <p>{t('no_manageable_sections')}</p>
      </div>
    );
  }

  return (
    <>
      <h3 className="mb-4">{t('title')}</h3>
      <Table responsive>
        <thead>
          <tr>
            <th>{t('section')}</th>
            <th style={{ width: '150px' }}>{t('visibility')}</th>
            <th style={{ width: '120px' }} />
          </tr>
        </thead>
        <tbody>
          {sections.map((section) => (
            <React.Fragment key={section.tag_id}>
              <tr>
                <td>
                  <strong>{section.display_name}</strong>
                  <small className="text-muted ms-2">
                    ({section.slug_name})
                  </small>
                </td>
                <td>
                  <Form.Select
                    size="sm"
                    value={section.visibility}
                    onChange={(e) =>
                      handleVisibilityChange(section, e.target.value)
                    }>
                    <option value="public">{t('public')}</option>
                    <option value="private">{t('private')}</option>
                  </Form.Select>
                </td>
                <td>
                  <Button
                    variant="outline-primary"
                    size="sm"
                    onClick={() =>
                      setExpandedSection(
                        expandedSection === section.slug_name
                          ? null
                          : section.slug_name,
                      )
                    }>
                    {expandedSection === section.slug_name
                      ? '\u25B2'
                      : '\u25BC'}{' '}
                    {t('members')}
                  </Button>
                </td>
              </tr>
              {expandedSection === section.slug_name && (
                <tr>
                  <td colSpan={3} className="bg-light">
                    {loadingPermissions ? (
                      <div className="text-center py-3">
                        <Spinner animation="border" size="sm" />
                      </div>
                    ) : (
                      <div className="p-2">
                        {/* Moderators - visible to all, but only admin can remove */}
                        <h6>{t('moderators')}</h6>
                        <div className="mb-3">
                          {permissions?.moderators &&
                          permissions.moderators.length > 0 ? (
                            permissions.moderators.map((name) => (
                              <Badge
                                key={name}
                                bg="primary"
                                className="me-1 mb-1 d-inline-flex align-items-center">
                                {name}
                                {isAdmin && (
                                  <Button
                                    variant="link"
                                    size="sm"
                                    className="p-0 ms-1 text-white"
                                    onClick={() =>
                                      handleRemove(name, 'moderator')
                                    }>
                                    &times;
                                  </Button>
                                )}
                              </Badge>
                            ))
                          ) : (
                            <span className="text-muted">
                              {t('no_moderators')}
                            </span>
                          )}
                        </div>

                        {/* Members */}
                        <h6>{t('members')}</h6>
                        <div className="mb-3">
                          {permissions?.members &&
                          permissions.members.length > 0 ? (
                            permissions.members.map((name) => (
                              <Badge
                                key={name}
                                bg="secondary"
                                className="me-1 mb-1 d-inline-flex align-items-center">
                                {name}
                                <Button
                                  variant="link"
                                  size="sm"
                                  className="p-0 ms-1 text-white"
                                  onClick={() => handleRemove(name, 'member')}>
                                  &times;
                                </Button>
                              </Badge>
                            ))
                          ) : (
                            <span className="text-muted">
                              {t('no_members')}
                            </span>
                          )}
                        </div>

                        {/* Invite form */}
                        <hr />
                        <h6>{t('invite_users')}</h6>
                        <UserSearchInput
                          selectedUsers={selectedUsers}
                          onSelect={(name) =>
                            setSelectedUsers((prev) => [...prev, name])
                          }
                          onRemove={(name) =>
                            setSelectedUsers((prev) =>
                              prev.filter((n) => n !== name),
                            )
                          }
                          placeholder={t('search_user_placeholder')}
                        />
                        <div className="d-flex align-items-center gap-2 mt-2">
                          <Form.Check
                            inline
                            type="radio"
                            name="inviteRole"
                            label={t('invite_as_member')}
                            checked={inviteRole === 'member'}
                            onChange={() => setInviteRole('member')}
                          />
                          {isAdmin && (
                            <Form.Check
                              inline
                              type="radio"
                              name="inviteRole"
                              label={t('invite_as_moderator')}
                              checked={inviteRole === 'moderator'}
                              onChange={() => setInviteRole('moderator')}
                            />
                          )}
                          <Button
                            variant="primary"
                            size="sm"
                            disabled={saving || selectedUsers.length === 0}
                            onClick={handleInvite}>
                            {saving ? (
                              <Spinner animation="border" size="sm" />
                            ) : (
                              t('invite')
                            )}
                          </Button>
                        </div>
                        {skippedUsers.length > 0 && (
                          <div className="text-warning mt-2 small">
                            {t('skipped_users')}: {skippedUsers.join(', ')}
                          </div>
                        )}
                      </div>
                    )}
                  </td>
                </tr>
              )}
            </React.Fragment>
          ))}
        </tbody>
      </Table>
    </>
  );
};

export default SectionManage;
