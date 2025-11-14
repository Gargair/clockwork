import { useCallback, useMemo, useState, type ChangeEvent, type FormEvent, type JSX } from 'react';

export interface ProjectFormValues {
	name: string;
	description?: string | null;
}

export interface ProjectFormProps {
	initial?: { name: string; description?: string | null };
	onSubmit: (values: ProjectFormValues) => void | Promise<void>;
	submitLabel?: string;
	disabled?: boolean;
}

function validateName(value: string): string | null {
	const trimmed = value.trim();
	if (trimmed.length === 0) return 'Name is required';
	return null;
}

export default function ProjectForm(props: Readonly<ProjectFormProps>): JSX.Element {
	const { initial, onSubmit, submitLabel = 'Save', disabled = false } = props;

	const [name, setName] = useState<string>(initial?.name ?? '');
	const [description, setDescription] = useState<string>(initial?.description ?? '');
	const [nameError, setNameError] = useState<string | null>(null);
	const [submitting, setSubmitting] = useState<boolean>(false);

	const isDisabled = useMemo<boolean>(() => disabled || submitting, [disabled, submitting]);

	const handleNameChange = useCallback((e: ChangeEvent<HTMLInputElement>): void => {
		setName(e.target.value);
	}, []);

	const handleDescriptionChange = useCallback((e: ChangeEvent<HTMLTextAreaElement>): void => {
		setDescription(e.target.value);
	}, []);

	const handleNameBlur = useCallback((): void => {
		setNameError(validateName(name));
	}, [name]);

	const handleSubmit = useCallback(async (e: FormEvent<HTMLFormElement>): Promise<void> => {
		e.preventDefault();
		const trimmedName = name.trim();
		const error = validateName(trimmedName);
		if (error) {
			setNameError(error);
			return;
		}
		setNameError(null);

		const descTrimmed = description?.trim?.() ?? '';
		const payload: ProjectFormValues = {
			name: trimmedName,
			description: descTrimmed.length > 0 ? descTrimmed : null,
		};

		try {
			setSubmitting(true);
			await onSubmit(payload);
		} finally {
			setSubmitting(false);
		}
	}, [name, description, onSubmit]);

	return (
		<form onSubmit={handleSubmit}>
			<div style={{ marginBottom: '0.75rem' }}>
				<label htmlFor="project-name">Name</label>
				<input
					id="project-name"
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
				<label htmlFor="project-description">Description (optional)</label>
				<textarea
					id="project-description"
					name="description"
					value={description}
					onChange={handleDescriptionChange}
					disabled={isDisabled}
					rows={3}
				/>
			</div>

			<div>
				<button type="submit" disabled={isDisabled}>
					{submitting ? 'Submitting…' : submitLabel}
				</button>
			</div>
		</form>
	);
}


