import { useCallback, useMemo, useState, type ChangeEvent, type FormEvent, type JSX } from 'react';

export interface CategoryFormValues {
  name: string;
  description?: string | null;
  parentCategoryId?: string | null;
}

export interface CategoryFormProps {
  initial?: { name: string; description?: string | null; parentCategoryId?: string | null };
  onSubmit: (values: CategoryFormValues) => void | Promise<void>;
  submitLabel?: string;
  disabled?: boolean;
  parentOptions: Array<{ value: string | null; label: string }>;
}

function validateName(value: string): string | null {
  const trimmed = value.trim();
  if (trimmed.length === 0) return 'Name is required';
  return null;
}

export default function CategoryForm(props: Readonly<CategoryFormProps>): JSX.Element {
  const { initial, onSubmit, submitLabel = 'Save', disabled = false, parentOptions } = props;

  const [name, setName] = useState<string>(initial?.name ?? '');
  const [description, setDescription] = useState<string>(initial?.description ?? '');
  const [parentCategoryId, setParentCategoryId] = useState<string | null>(
    initial?.parentCategoryId ?? null,
  );
  const [nameError, setNameError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState<boolean>(false);

  const isDisabled = useMemo<boolean>(() => disabled || submitting, [disabled, submitting]);

  const handleNameChange = useCallback((e: ChangeEvent<HTMLInputElement>): void => {
    setName(e.target.value);
  }, []);

  const handleDescriptionChange = useCallback((e: ChangeEvent<HTMLTextAreaElement>): void => {
    setDescription(e.target.value);
  }, []);

  const handleParentChange = useCallback((e: ChangeEvent<HTMLSelectElement>): void => {
    const value = e.target.value;
    setParentCategoryId(value === '' ? null : value);
  }, []);

  const handleNameBlur = useCallback((): void => {
    setNameError(validateName(name));
  }, [name]);

  const handleSubmit = useCallback(
    async (e: FormEvent<HTMLFormElement>): Promise<void> => {
      e.preventDefault();
      const trimmedName = name.trim();
      const error = validateName(trimmedName);
      if (error) {
        setNameError(error);
        return;
      }
      setNameError(null);

      const descTrimmed = description?.trim?.() ?? '';
      const payload: CategoryFormValues = {
        name: trimmedName,
        description: descTrimmed.length > 0 ? descTrimmed : null,
        parentCategoryId,
      };

      try {
        setSubmitting(true);
        await onSubmit(payload);
      } finally {
        setSubmitting(false);
      }
    },
    [name, description, parentCategoryId, onSubmit],
  );

  return (
    <form onSubmit={handleSubmit}>
      <div style={{ marginBottom: '0.75rem' }}>
        <label htmlFor="category-name">Name</label>
        <input
          id="category-name"
          type="text"
          name="name"
          required
          value={name}
          onChange={handleNameChange}
          onBlur={handleNameBlur}
          disabled={isDisabled}
        />
        {nameError ? (
          <p role="alert" style={{ color: 'var(--color-danger, red)', marginTop: '0.25rem' }}>
            {nameError}
          </p>
        ) : null}
      </div>

      <div style={{ marginBottom: '0.75rem' }}>
        <label htmlFor="category-description">Description (optional)</label>
        <textarea
          id="category-description"
          name="description"
          value={description}
          onChange={handleDescriptionChange}
          disabled={isDisabled}
          rows={3}
        />
      </div>

      <div style={{ marginBottom: '0.75rem' }}>
        <label htmlFor="category-parent">Parent Category</label>
        <select
          id="category-parent"
          name="parentCategoryId"
          value={parentCategoryId ?? ''}
          onChange={handleParentChange}
          disabled={isDisabled}
        >
          {parentOptions.map((option) => (
            <option key={option.value ?? 'none'} value={option.value ?? ''}>
              {option.label}
            </option>
          ))}
        </select>
      </div>

      <div>
        <button type="submit" disabled={isDisabled}>
          {submitting ? 'Submitting…' : submitLabel}
        </button>
      </div>
    </form>
  );
}

