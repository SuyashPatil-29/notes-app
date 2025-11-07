package commands

import (
	"backend/internal/models"
	"backend/internal/whatsapp"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

// ListCommand handles the /list command for listing notes, notebooks, and chapters
type ListCommand struct{}

// NewListCommand creates a new list command
func NewListCommand() *ListCommand {
	return &ListCommand{}
}

// Name returns the command name
func (c *ListCommand) Name() string {
	return "list"
}

// Description returns the command description
func (c *ListCommand) Description() string {
	return "List your notes, notebooks, or chapters"
}

// Usage returns usage instructions
func (c *ListCommand) Usage() string {
	return "/list [notes|notebooks|chapters] - List recent notes, all notebooks, or chapters"
}

// RequiresAuth returns whether authentication is required
func (c *ListCommand) RequiresAuth() bool {
	return true
}

// Execute runs the list command
func (c *ListCommand) Execute(ctx *whatsapp.CommandContext) error {
	// Determine what to list
	listType := "notes" // Default to notes
	if len(ctx.Args) > 0 {
		listType = strings.ToLower(ctx.Args[0])
	}

	switch listType {
	case "notes", "note":
		return c.listNotes(ctx)
	case "notebooks", "notebook":
		return c.listNotebooks(ctx)
	case "chapters", "chapter":
		return c.listChapters(ctx)
	default:
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ Invalid list type.\n\n*Usage:* /list [notes|notebooks|chapters]")
	}
}

// listNotes lists recent notes
func (c *ListCommand) listNotes(ctx *whatsapp.CommandContext) error {
	var notes []models.Notes

	query := ctx.DB.Preload("Chapter").Preload("Chapter.Notebook").
		Joins("JOIN chapters ON chapters.id = notes.chapter_id").
		Joins("JOIN notebooks ON notebooks.id = chapters.notebook_id").
		Where("notebooks.clerk_user_id = ?", ctx.User.ClerkUserID)

	if ctx.OrganizationID != nil {
		query = query.Where("notes.organization_id = ?", *ctx.OrganizationID)
	} else {
		query = query.Where("notes.organization_id IS NULL")
	}

	if err := query.Order("notes.updated_at DESC").Limit(10).Find(&notes).Error; err != nil {
		log.Error().Err(err).Msg("Failed to list notes")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred while listing notes. Please try again.")
	}

	if len(notes) == 0 {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"📝 You don't have any notes yet.\n\nUse /add to create your first note!")
	}

	var message strings.Builder
	message.WriteString(fmt.Sprintf("📝 *Your Recent Notes* (%d)\n\n", len(notes)))

	for i, note := range notes {
		notebookName := "Unknown"
		chapterName := "Unknown"

		if note.Chapter.Notebook.Name != "" {
			notebookName = note.Chapter.Notebook.Name
		}
		if note.Chapter.Name != "" {
			chapterName = note.Chapter.Name
		}

		message.WriteString(fmt.Sprintf("%d. *%s*\n", i+1, note.Name))
		message.WriteString(fmt.Sprintf("   📓 %s › 📑 %s\n", notebookName, chapterName))
		message.WriteString(fmt.Sprintf("   🕒 %s\n\n", note.UpdatedAt.Format("Jan 2, 2006")))
	}

	message.WriteString("_Use /retrieve [note name] to view a note's content_")

	return ctx.Client.SendTextMessage(ctx.PhoneNumber, message.String())
}

// listNotebooks lists all notebooks
func (c *ListCommand) listNotebooks(ctx *whatsapp.CommandContext) error {
	var notebooks []models.Notebook

	query := ctx.DB.Preload("Chapters").
		Where("clerk_user_id = ?", ctx.User.ClerkUserID)

	if ctx.OrganizationID != nil {
		query = query.Where("organization_id = ?", *ctx.OrganizationID)
	} else {
		query = query.Where("organization_id IS NULL")
	}

	if err := query.Order("created_at DESC").Find(&notebooks).Error; err != nil {
		log.Error().Err(err).Msg("Failed to list notebooks")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred while listing notebooks. Please try again.")
	}

	if len(notebooks) == 0 {
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"📓 You don't have any notebooks yet.\n\nUse /create notebook [name] to create your first notebook!")
	}

	var message strings.Builder
	message.WriteString(fmt.Sprintf("📓 *Your Notebooks* (%d)\n\n", len(notebooks)))

	for i, notebook := range notebooks {
		chapterCount := len(notebook.Chapters)
		message.WriteString(fmt.Sprintf("%d. *%s*\n", i+1, notebook.Name))
		message.WriteString(fmt.Sprintf("   📑 %d chapter(s)\n", chapterCount))
		message.WriteString(fmt.Sprintf("   🕒 Created %s\n\n", notebook.CreatedAt.Format("Jan 2, 2006")))
	}

	message.WriteString("_Use /list chapters to see chapters in your notebooks_")

	return ctx.Client.SendTextMessage(ctx.PhoneNumber, message.String())
}

// listChapters lists chapters, optionally filtered by notebook
func (c *ListCommand) listChapters(ctx *whatsapp.CommandContext) error {
	var chapters []models.Chapter

	query := ctx.DB.Preload("Notebook").Preload("Files").
		Joins("JOIN notebooks ON notebooks.id = chapters.notebook_id").
		Where("notebooks.clerk_user_id = ?", ctx.User.ClerkUserID)

	if ctx.OrganizationID != nil {
		query = query.Where("chapters.organization_id = ?", *ctx.OrganizationID)
	} else {
		query = query.Where("chapters.organization_id IS NULL")
	}

	// Check if user specified a notebook filter
	if len(ctx.Args) > 1 {
		notebookFilter := strings.Join(ctx.Args[1:], " ")
		query = query.Where("LOWER(notebooks.name) LIKE ?", "%"+strings.ToLower(notebookFilter)+"%")
	}

	if err := query.Order("chapters.created_at DESC").Find(&chapters).Error; err != nil {
		log.Error().Err(err).Msg("Failed to list chapters")
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"❌ An error occurred while listing chapters. Please try again.")
	}

	if len(chapters) == 0 {
		if len(ctx.Args) > 1 {
			notebookFilter := strings.Join(ctx.Args[1:], " ")
			return ctx.Client.SendTextMessage(ctx.PhoneNumber,
				fmt.Sprintf("📑 No chapters found in notebooks matching: *%s*", notebookFilter))
		}
		return ctx.Client.SendTextMessage(ctx.PhoneNumber,
			"📑 You don't have any chapters yet.\n\nUse /create chapter [name] to create your first chapter!")
	}

	var message strings.Builder
	if len(ctx.Args) > 1 {
		notebookFilter := strings.Join(ctx.Args[1:], " ")
		message.WriteString(fmt.Sprintf("📑 *Chapters in '%s'* (%d)\n\n", notebookFilter, len(chapters)))
	} else {
		message.WriteString(fmt.Sprintf("📑 *Your Chapters* (%d)\n\n", len(chapters)))
	}

	for i, chapter := range chapters {
		notebookName := "Unknown"
		if chapter.Notebook.Name != "" {
			notebookName = chapter.Notebook.Name
		}

		noteCount := len(chapter.Files)
		message.WriteString(fmt.Sprintf("%d. *%s*\n", i+1, chapter.Name))
		message.WriteString(fmt.Sprintf("   📓 %s\n", notebookName))
		message.WriteString(fmt.Sprintf("   📝 %d note(s)\n", noteCount))
		message.WriteString(fmt.Sprintf("   🕒 Created %s\n\n", chapter.CreatedAt.Format("Jan 2, 2006")))
	}

	message.WriteString("_Use /list notes to see your notes_")

	return ctx.Client.SendTextMessage(ctx.PhoneNumber, message.String())
}
