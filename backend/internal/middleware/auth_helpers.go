package middleware

import (
	"backend/internal/models"
	"context"

	"github.com/gin-gonic/gin"
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
		return false, err
	}

	// Personal notebook check
	if result.OrganizationID == nil || *result.OrganizationID == "" {
		return result.ClerkUserID == clerkUserID, nil
	}

	// Organization notebook check - use cached version
	_, isMember, err := GetOrgMemberRoleCached(ctx, *result.OrganizationID, clerkUserID)
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
		NotebookID string
	}

	err := db.Model(&models.Chapter{}).
		Select("notebook_id").
		Where("id = ?", chapterID).
		First(&result).Error

	if err != nil {
		return false, err
	}

	return CheckNotebookAccess(ctx, db, result.NotebookID, clerkUserID)
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
		ChapterID string
	}

	err := db.Model(&models.Notes{}).
		Select("chapter_id").
		Where("id = ?", noteID).
		First(&result).Error

	if err != nil {
		return false, err
	}

	return CheckChapterAccess(ctx, db, result.ChapterID, clerkUserID)
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
