package middleware

import (
	"backend/internal/models"
	"context"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// CheckNotebookAccess verifies user has access to notebook without preloading full relationships
func CheckNotebookAccess(ctx context.Context, db *gorm.DB, notebookID, clerkUserID string) (bool, error) {
	var result struct {
		ClerkUserID    string
		OrganizationID *string
	}

	err := db.Model(&models.Notebook{}).
		Select("clerk_user_id, organization_id").
		Where("id = ?", notebookID).
		First(&result).Error

	if err != nil {
		log.Error().Err(err).Str("notebook_id", notebookID).Str("user_id", clerkUserID).Msg("CheckNotebookAccess: Notebook not found")
		return false, err
	}

	log.Debug().
		Str("notebook_id", notebookID).
		Str("owner_id", result.ClerkUserID).
		Str("org_id", func() string {
			if result.OrganizationID != nil {
				return *result.OrganizationID
			}
			return "nil"
		}()).
		Str("user_id", clerkUserID).
		Msg("CheckNotebookAccess: Found notebook")

	// Personal notebook check
	if result.OrganizationID == nil || *result.OrganizationID == "" {
		hasAccess := result.ClerkUserID == clerkUserID
		if !hasAccess {
			log.Warn().
				Str("notebook_id", notebookID).
				Str("owner_id", result.ClerkUserID).
				Str("user_id", clerkUserID).
				Msg("CheckNotebookAccess: Personal notebook - user is not owner")
		} else {
			log.Debug().Str("notebook_id", notebookID).Str("user_id", clerkUserID).Msg("CheckNotebookAccess: Personal notebook - access granted")
		}
		return hasAccess, nil
	}

	// Organization notebook check - use cached version
	log.Debug().
		Str("notebook_id", notebookID).
		Str("org_id", *result.OrganizationID).
		Str("user_id", clerkUserID).
		Msg("CheckNotebookAccess: Organization notebook - checking membership")

	_, isMember, err := GetOrgMemberRoleCached(ctx, *result.OrganizationID, clerkUserID)
	if err != nil {
		log.Error().Err(err).Str("notebook_id", notebookID).Str("org_id", *result.OrganizationID).Str("user_id", clerkUserID).Msg("CheckNotebookAccess: Error checking org membership")
	} else if !isMember {
		log.Warn().Str("notebook_id", notebookID).Str("org_id", *result.OrganizationID).Str("user_id", clerkUserID).Msg("CheckNotebookAccess: User is not org member")
	} else {
		log.Debug().Str("notebook_id", notebookID).Str("org_id", *result.OrganizationID).Str("user_id", clerkUserID).Msg("CheckNotebookAccess: Org membership verified - access granted")
	}

	return isMember, err
}

// CheckNotebookAccessWithGin is a Gin-aware version that uses request-level caching
func CheckNotebookAccessWithGin(c *gin.Context, db *gorm.DB, notebookID, clerkUserID string) (bool, error) {
	var result struct {
		ClerkUserID    string
		OrganizationID *string
	}

	err := db.Model(&models.Notebook{}).
		Select("clerk_user_id, organization_id").
		Where("id = ?", notebookID).
		First(&result).Error

	if err != nil {
		return false, err
	}

	// Personal notebook check
	if result.OrganizationID == nil || *result.OrganizationID == "" {
		return result.ClerkUserID == clerkUserID, nil
	}

	// Organization notebook check - use request-level cache
	_, isMember, err := GetOrgMemberRoleWithRequestCache(c.Request.Context(), c, *result.OrganizationID, clerkUserID)
	return isMember, err
}

// CheckChapterAccess verifies user has access to chapter without preloading full relationships
func CheckChapterAccess(ctx context.Context, db *gorm.DB, chapterID, clerkUserID string) (bool, error) {
	var result struct {
		NotebookID     string
		OrganizationID *string
	}

	err := db.Model(&models.Chapter{}).
		Select("chapters.notebook_id, chapters.organization_id").
		Where("chapters.id = ?", chapterID).
		First(&result).Error

	if err != nil {
		log.Error().Err(err).Str("chapter_id", chapterID).Str("user_id", clerkUserID).Msg("CheckChapterAccess: Chapter not found")
		return false, err
	}

	log.Debug().
		Str("chapter_id", chapterID).
		Str("notebook_id", result.NotebookID).
		Str("org_id", func() string {
			if result.OrganizationID != nil {
				return *result.OrganizationID
			}
			return "nil"
		}()).
		Str("user_id", clerkUserID).
		Msg("CheckChapterAccess: Found chapter, checking notebook access")

	hasAccess, err := CheckNotebookAccess(ctx, db, result.NotebookID, clerkUserID)
	if err != nil {
		log.Error().Err(err).Str("chapter_id", chapterID).Str("notebook_id", result.NotebookID).Str("user_id", clerkUserID).Msg("CheckChapterAccess: Error checking notebook access")
	} else if !hasAccess {
		log.Warn().Str("chapter_id", chapterID).Str("notebook_id", result.NotebookID).Str("user_id", clerkUserID).Msg("CheckChapterAccess: User does not have notebook access")
	}

	return hasAccess, err
}

// CheckChapterAccessWithGin is a Gin-aware version that uses request-level caching
func CheckChapterAccessWithGin(c *gin.Context, db *gorm.DB, chapterID, clerkUserID string) (bool, error) {
	var result struct {
		NotebookID string
	}

	err := db.Model(&models.Chapter{}).
		Select("notebook_id").
		Where("id = ?", chapterID).
		First(&result).Error

	if err != nil {
		return false, err
	}

	return CheckNotebookAccessWithGin(c, db, result.NotebookID, clerkUserID)
}

// CheckNoteAccess verifies user has access to note without preloading full relationships
func CheckNoteAccess(ctx context.Context, db *gorm.DB, noteID, clerkUserID string) (bool, error) {
	var result struct {
		ChapterID      string
		OrganizationID *string
	}

	err := db.Model(&models.Notes{}).
		Select("notes.chapter_id, notes.organization_id").
		Where("notes.id = ?", noteID).
		First(&result).Error

	if err != nil {
		log.Error().Err(err).Str("note_id", noteID).Str("user_id", clerkUserID).Msg("CheckNoteAccess: Note not found")
		return false, err
	}

	log.Debug().
		Str("note_id", noteID).
		Str("chapter_id", result.ChapterID).
		Str("org_id", func() string {
			if result.OrganizationID != nil {
				return *result.OrganizationID
			}
			return "nil"
		}()).
		Str("user_id", clerkUserID).
		Msg("CheckNoteAccess: Found note, checking chapter access")

	hasAccess, err := CheckChapterAccess(ctx, db, result.ChapterID, clerkUserID)
	if err != nil {
		log.Error().Err(err).Str("note_id", noteID).Str("chapter_id", result.ChapterID).Str("user_id", clerkUserID).Msg("CheckNoteAccess: Error checking chapter access")
	} else if !hasAccess {
		log.Warn().Str("note_id", noteID).Str("chapter_id", result.ChapterID).Str("user_id", clerkUserID).Msg("CheckNoteAccess: User does not have chapter access")
	} else {
		log.Debug().Str("note_id", noteID).Str("user_id", clerkUserID).Msg("CheckNoteAccess: Access granted")
	}

	return hasAccess, err
}

// CheckNoteAccessWithGin is a Gin-aware version that uses request-level caching
func CheckNoteAccessWithGin(c *gin.Context, db *gorm.DB, noteID, clerkUserID string) (bool, error) {
	var result struct {
		ChapterID string
	}

	err := db.Model(&models.Notes{}).
		Select("chapter_id").
		Where("id = ?", noteID).
		First(&result).Error

	if err != nil {
		return false, err
	}

	return CheckChapterAccessWithGin(c, db, result.ChapterID, clerkUserID)
}
