package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	// Ensure this import path is correct for your project
	"github.com/mickali02/mood-notes-app/internal/validator"
)

// MoodNote struct represents a single mood note entry in the database.
type MoodNote struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Version   int       `json:"version"`
}

// ValidateMoodNote checks the mood note fields against validation rules.
func ValidateMoodNote(v *validator.Validator, note *MoodNote) {
	v.Check(validator.NotBlank(note.Title), "title", "must be provided")
	v.Check(validator.NotBlank(note.Content), "content", "must be provided")
	v.Check(validator.MaxLength(note.Title, 150), "title", "must not be more than 150 characters long")
	v.Check(validator.MaxLength(note.Content, 5000), "content", "must not be more than 5000 characters long")
}

// MoodNoteModel struct provides methods for interacting with mood note data.
type MoodNoteModel struct {
	DB *sql.DB
}

// Insert adds a new MoodNote record into the 'mood_notes' table.
func (m *MoodNoteModel) Insert(note *MoodNote) error {
	query := `
		INSERT INTO mood_notes (title, content)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at, version`

	args := []any{note.Title, note.Content}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&note.ID, &note.CreatedAt, &note.UpdatedAt, &note.Version)
}

// Get retrieves a specific MoodNote record by ID.
func (m *MoodNoteModel) Get(id int64) (*MoodNote, error) {
	if id < 1 {
		return nil, errors.New("invalid mood note ID provided")
	}

	query := `
		SELECT id, created_at, updated_at, title, content, version
		FROM mood_notes
		WHERE id = $1`

	var note MoodNote
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&note.ID,
		&note.CreatedAt,
		&note.UpdatedAt,
		&note.Title,
		&note.Content,
		&note.Version,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("mood note record not found")
		}
		return nil, err
	}

	return &note, nil
}

// GetAll retrieves all mood note entries from the database.
func (m *MoodNoteModel) GetAll() ([]*MoodNote, error) {
	query := `
		SELECT id, created_at, updated_at, title, content, version
		FROM mood_notes
		ORDER BY created_at DESC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []*MoodNote
	for rows.Next() {
		n := &MoodNote{}
		err := rows.Scan(
			&n.ID,
			&n.CreatedAt,
			&n.UpdatedAt,
			&n.Title,
			&n.Content,
			&n.Version,
		)
		if err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}

// Update modifies an existing mood note record.
func (m *MoodNoteModel) Update(note *MoodNote) error {
	if note.ID < 1 {
		return errors.New("invalid mood note ID for update")
	}

	query := `
		UPDATE mood_notes
		SET title = $1, content = $2, version = version + 1
		WHERE id = $3 AND version = $4
		RETURNING updated_at, version`

	args := []any{note.Title, note.Content, note.ID, note.Version}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&note.UpdatedAt, &note.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("mood note record not found or version mismatch")
		}
		return err
	}

	return nil
}

// Delete removes a specific mood note entry from the database.
func (m *MoodNoteModel) Delete(id int64) error {
	if id < 1 {
		return errors.New("invalid mood note ID provided")
	}

	query := `
		DELETE FROM mood_notes
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("mood note record not found or already deleted")
	}

	return nil
}