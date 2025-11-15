import { useCallback, useState, type JSX } from 'react';
import type { CategoryNode } from '../hooks/useCategories';

export interface CategoryTreeProps {
  tree: CategoryNode[];
  onAdd: (parentId: string | null) => void;
  onEdit: (categoryId: string) => void;
  onDelete: (categoryId: string) => void;
  loading?: boolean;
  error?: string | null;
}

export default function CategoryTree(props: Readonly<CategoryTreeProps>): JSX.Element {
  const { tree, onAdd, onEdit, onDelete, loading = false, error = null } = props;
  const [expanded, setExpanded] = useState<Set<string>>(new Set());

  const toggleExpanded = useCallback((categoryId: string): void => {
    setExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(categoryId)) {
        next.delete(categoryId);
      } else {
        next.add(categoryId);
      }
      return next;
    });
  }, []);

  const handleAdd = useCallback(
    (parentId: string | null): void => {
      onAdd(parentId);
    },
    [onAdd],
  );

  const handleEdit = useCallback(
    (categoryId: string): void => {
      onEdit(categoryId);
    },
    [onEdit],
  );

  const handleDelete = useCallback(
    (categoryId: string): void => {
      onDelete(categoryId);
    },
    [onDelete],
  );

  if (loading) {
    return <p>Loading categories…</p>;
  }

  if (error) {
    return (
      <div role="alert">
        <p>Error: {error}</p>
      </div>
    );
  }

  if (tree.length === 0) {
    return <p>No categories yet.</p>;
  }

  return (
    <ul style={{ listStyle: 'none', paddingLeft: 0 }}>
      {tree.map((node) => (
        <CategoryTreeNode
          key={node.id}
          node={node}
          expandedSet={expanded}
          onToggleExpanded={toggleExpanded}
          onAdd={handleAdd}
          onEdit={handleEdit}
          onDelete={handleDelete}
          level={0}
        />
      ))}
    </ul>
  );
}

interface CategoryTreeNodeProps {
  node: CategoryNode;
  expandedSet: Set<string>;
  onToggleExpanded: (categoryId: string) => void;
  onAdd: (parentId: string | null) => void;
  onEdit: (categoryId: string) => void;
  onDelete: (categoryId: string) => void;
  level: number;
}

function CategoryTreeNode(props: Readonly<CategoryTreeNodeProps>): JSX.Element {
  const { node, expandedSet, onToggleExpanded, onAdd, onEdit, onDelete, level } = props;
  const hasChildren = node.children.length > 0;
  const isExpanded = expandedSet.has(node.id);
  const indent = level * 1.5;

  const handleToggle = useCallback((): void => {
    onToggleExpanded(node.id);
  }, [node.id, onToggleExpanded]);

  const handleAdd = useCallback((): void => {
    onAdd(node.id);
  }, [node.id, onAdd]);

  const handleEdit = useCallback((): void => {
    onEdit(node.id);
  }, [node.id, onEdit]);

  const handleDelete = useCallback((): void => {
    onDelete(node.id);
  }, [node.id, onDelete]);

  return (
    <li>
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: '0.5rem',
          paddingLeft: `${indent}rem`,
          marginBottom: '0.25rem',
        }}
      >
        {hasChildren ? (
          <button
            type="button"
            onClick={handleToggle}
            aria-label={isExpanded ? 'Collapse' : 'Expand'}
            style={{
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              padding: '0.25rem',
              fontSize: '0.875rem',
            }}
          >
            {isExpanded ? '▼' : '▶'}
          </button>
        ) : (
          <span style={{ width: '1.5rem', display: 'inline-block' }} />
        )}
        <span style={{ flex: 1 }}>{node.name}</span>
        {node.description ? (
          <span style={{ color: 'var(--color-muted)', fontSize: '0.875rem' }}>
            {node.description}
          </span>
        ) : null}
        <div style={{ display: 'flex', gap: '0.25rem' }}>
          <button type="button" onClick={handleAdd} aria-label={`Add subcategory to ${node.name}`}>
            Add Child
          </button>
          <button type="button" onClick={handleEdit} aria-label={`Edit ${node.name}`}>
            Edit
          </button>
          <button type="button" onClick={handleDelete} aria-label={`Delete ${node.name}`}>
            Delete
          </button>
        </div>
      </div>
      {hasChildren && isExpanded ? (
        <ul style={{ listStyle: 'none', paddingLeft: 0 }}>
          {node.children.map((child) => (
            <CategoryTreeNode
              key={child.id}
              node={child}
              expandedSet={expandedSet}
              onToggleExpanded={onToggleExpanded}
              onAdd={onAdd}
              onEdit={onEdit}
              onDelete={onDelete}
              level={level + 1}
            />
          ))}
        </ul>
      ) : null}
    </li>
  );
}

