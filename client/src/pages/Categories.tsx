import { useCallback, useMemo, useState, type JSX } from 'react';
import { useParams, Link } from 'react-router-dom';
import {
  useCategories,
  buildCategoryTree,
  type CategoryNode,
  type UseCategoriesError,
} from '../hooks/useCategories';
import CategoryForm, { type CategoryFormValues } from '../components/CategoryForm';
import CategoryTree from '../components/CategoryTree';
import type { Category } from '../types';

export default function Categories(): JSX.Element {
  const { projectId } = useParams<{ projectId: string }>();
  if (!projectId) {
    return (
      <section>
        <div role="alert">
          <p>Error: Project ID is required</p>
        </div>
      </section>
    );
  }
  return <CategoryList projectId={projectId} />;
}

export function CategoryList({ projectId }: Readonly<{ projectId: string }>): JSX.Element {
  const { status, categories, error, refresh, create, update, remove } = useCategories(projectId);
  const [editingCategoryId, setEditingCategoryId] = useState<string | null>(null);
  const [addingChildToId, setAddingChildToId] = useState<string | null>(null);
  const [creatingTopLevel, setCreatingTopLevel] = useState<boolean>(false);

  const tree = useMemo<CategoryNode[]>(() => buildCategoryTree(categories), [categories]);
  const isLoading: boolean = status === 'loading';

  // Build parent options for forms (exclude the category being edited and its descendants to prevent cycles)
  const buildParentOptions = useCallback(
    (excludeCategoryId: string | null): Array<{ value: string | null; label: string }> => {
      const options: Array<{ value: string | null; label: string }> = [
        { value: null, label: 'None (Top-level)' },
      ];

      if (!excludeCategoryId) {
        // If not editing, include all categories
        for (const category of categories) {
          options.push({ value: category.id, label: category.name });
        }
        return options;
      }

      // If editing, exclude the category being edited and all its descendants
      const excludeSet = new Set<string>([excludeCategoryId]);
      const findDescendants = (node: CategoryNode): void => {
        for (const child of node.children) {
          excludeSet.add(child.id);
          findDescendants(child);
        }
      };

      for (const rootNode of tree) {
        const editingNode = findNodeById(rootNode, excludeCategoryId);
        if (editingNode) {
          findDescendants(editingNode);
          break;
        }
      }

      for (const category of categories) {
        if (!excludeSet.has(category.id)) {
          options.push({ value: category.id, label: category.name });
        }
      }

      return options;
    },
    [categories, tree],
  );

  const handleCancelEdit = useCallback((): void => {
    setEditingCategoryId(null);
    setAddingChildToId(null);
    setCreatingTopLevel(false);
  }, []);

  const handleCreateTopLevel = useCallback(
    async (values: CategoryFormValues): Promise<void> => {
      await create(values);
      setCreatingTopLevel(false);
    },
    [create],
  );

  const handleAddChild = useCallback((parentId: string | null): void => {
    setAddingChildToId(parentId);
    setEditingCategoryId(null);
    setCreatingTopLevel(false);
  }, []);

  const handleAddChildSubmit = useCallback(
    async (values: CategoryFormValues): Promise<void> => {
      await create(values);
      setAddingChildToId(null);
    },
    [create],
  );

  const handleEdit = useCallback((categoryId: string): void => {
    setEditingCategoryId(categoryId);
    setAddingChildToId(null);
    setCreatingTopLevel(false);
  }, []);

  const handleEditSubmit = useMemo<
    Record<string, (values: CategoryFormValues) => Promise<void>>
  >(() => {
    const entries: Array<[string, (values: CategoryFormValues) => Promise<void>]> = categories.map(
      (c: Category) => [
        c.id,
        async (values: CategoryFormValues): Promise<void> => {
          await update(c.id, values);
          setEditingCategoryId(null);
        },
      ],
    );
    return Object.fromEntries(entries);
  }, [categories, update]);

  const handleDelete = useCallback(
    async (categoryId: string): Promise<void> => {
      await remove(categoryId);
      if (editingCategoryId === categoryId) {
        setEditingCategoryId(null);
      }
      if (addingChildToId === categoryId) {
        setAddingChildToId(null);
      }
    },
    [remove, editingCategoryId, addingChildToId],
  );

  // Render tree with edit/add/delete handlers
  const renderTree = useCallback((): JSX.Element => {
    if (editingCategoryId || addingChildToId !== null || creatingTopLevel) {
      // Don't show tree actions when editing/adding
      return (
        <CategoryTree
          tree={tree}
          onAdd={() => {}}
          onEdit={() => {}}
          onDelete={() => {}}
          loading={isLoading}
          error={null}
        />
      );
    }

    return (
      <CategoryTree
        tree={tree}
        onAdd={handleAddChild}
        onEdit={handleEdit}
        onDelete={handleDelete}
        loading={isLoading}
        error={null}
      />
    );
  }, [
    tree,
    editingCategoryId,
    addingChildToId,
    creatingTopLevel,
    isLoading,
    handleAddChild,
    handleEdit,
    handleDelete,
  ]);

  const editingCategory = useMemo<Category | undefined>(() => {
    if (!editingCategoryId) return undefined;
    return categories.find((c: Category) => c.id === editingCategoryId);
  }, [categories, editingCategoryId]);

  return (
    <section>
      <header>
        <h1>Categories</h1>
        <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
          <Link to="/projects">← Back to Projects</Link>
          <button type="button" onClick={refresh} disabled={isLoading}>
            {isLoading ? 'Refreshing…' : 'Refresh'}
          </button>
        </div>
      </header>

      {status === 'loading' && categories.length === 0 ? <p>Loading categories…</p> : null}
      {status === 'error' ? renderError(error) : null}

      {!creatingTopLevel && !editingCategoryId && addingChildToId === null ? (
        <div style={{ marginTop: '1rem', marginBottom: '1rem' }}>
          <h2>Create Top-Level Category</h2>
          <button type="button" onClick={() => setCreatingTopLevel(true)}>
            Create Category
          </button>
        </div>
      ) : null}

      {creatingTopLevel ? (
        <div style={{ marginTop: '1rem', marginBottom: '1rem' }}>
          <h2>Create Top-Level Category</h2>
          <CategoryForm
            onSubmit={handleCreateTopLevel}
            submitLabel="Create"
            parentOptions={buildParentOptions(null)}
          />
          <div style={{ marginTop: '0.5rem' }}>
            <button type="button" onClick={handleCancelEdit}>
              Cancel
            </button>
          </div>
        </div>
      ) : null}

      {editingCategoryId && editingCategory ? (
        <div style={{ marginTop: '1rem', marginBottom: '1rem' }}>
          <h2>Edit Category: {editingCategory.name}</h2>
          <CategoryForm
            initial={{
              name: editingCategory.name,
              description: editingCategory.description ?? null,
              parentCategoryId: editingCategory.parentCategoryId ?? null,
            }}
            onSubmit={handleEditSubmit[editingCategoryId]!}
            submitLabel="Save"
            parentOptions={buildParentOptions(editingCategoryId)}
          />
          <div style={{ marginTop: '0.5rem' }}>
            <button type="button" onClick={handleCancelEdit}>
              Cancel
            </button>
          </div>
        </div>
      ) : null}

      {addingChildToId === null ? null : (
        <div style={{ marginTop: '1rem', marginBottom: '1rem' }}>
          <h2>Add Child Category</h2>
          <CategoryForm
            initial={{ parentCategoryId: addingChildToId }}
            onSubmit={handleAddChildSubmit}
            submitLabel="Create"
            parentOptions={buildParentOptions(null)}
          />
          <div style={{ marginTop: '0.5rem' }}>
            <button type="button" onClick={handleCancelEdit}>
              Cancel
            </button>
          </div>
        </div>
      )}

      <div style={{ marginTop: '1rem' }}>
        <h2>Category Tree</h2>
        {renderTree()}
      </div>
    </section>
  );
}

function findNodeById(node: CategoryNode, id: string): CategoryNode | null {
  if (node.id === id) return node;
  for (const child of node.children) {
    const found = findNodeById(child, id);
    if (found) return found;
  }
  return null;
}

function renderError(err: UseCategoriesError | null): JSX.Element | null {
  if (!err) return null;

  const errorMessage = getErrorMessage(err);

  return (
    <div role="alert">
      <p>{errorMessage}</p>
      {err.code ? <p>Code: {err.code}</p> : null}
      {err.requestId ? <p>Request ID: {err.requestId}</p> : null}
    </div>
  );
}

function getErrorMessage(err: UseCategoriesError): string {
  switch (err.code) {
    case 'invalid_parent':
      return 'Error: Invalid parent category. The parent category does not exist or does not belong to this project.';
    case 'cross_project_parent':
      return 'Error: Cannot use a parent category from a different project.';
    case 'category_cycle':
      return 'Error: Cannot create a cycle. A category cannot be its own parent or ancestor.';
    default:
      return `Error: ${err.message}`;
  }
}
