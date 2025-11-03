package main

import (
	"backend/db"
	"backend/internal/models"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/joho/godotenv"
)

var (
	userID = "user_34xT6iCOgagjmRT9wmO8q0hLT6P"
	orgID  = "org_34xTFKmaPVwo7CWuPjjeWpvBgvk"
)

var (
	noteTitles = []string{
		"Getting Started", "API Design", "Database Schema", "Authentication Flow",
		"Error Handling", "Testing Strategy", "Deployment Guide", "Performance Tips",
		"Security Best Practices", "Code Review Notes", "Bug Fixes", "Feature Ideas",
		"Meeting Notes", "Architecture Decisions", "Refactoring Tasks", "Documentation",
		"User Feedback", "Sprint Planning", "Retrospective", "Daily Standup",
		"Technical Debt", "Code Optimization", "API Endpoints", "Data Models",
		"Middleware Setup", "Caching Strategy", "Query Optimization", "Index Design",
		"Migration Scripts", "Backup Strategy", "Monitoring Setup", "Logging",
		"CI/CD Pipeline", "Docker Setup", "Kubernetes Config", "Load Balancing",
		"Rate Limiting", "CORS Configuration", "JWT Implementation", "OAuth Flow",
		"WebSocket Setup", "Real-time Features", "Notification System", "Email Templates",
		"SMS Integration", "Payment Gateway", "Third-party APIs", "Webhooks",
		"GraphQL Schema", "REST vs GraphQL", "API Versioning", "Deprecation Strategy",
		"Database Indexing", "Connection Pooling", "Error Logging", "Input Validation",
		"Session Management", "Token Storage", "Request Throttling", "API Throttling",
		"Data Encryption", "Access Control", "Audit Logging", "Compliance Checks",
		"Data Migration", "Schema Updates", "Data Seeding", "Data Archiving",
		"Data Backup", "Data Recovery", "Data Retention", "Data Privacy",
		"Data Security", "Data Integrity", "Data Consistency", "Data Replication",
		"Data Sharding", "Data Partitioning", "Data Caching", "Data Synchronization",
		"Data Transformation", "Data Aggregation", "Data Analysis", "Data Visualization",
		"Data Reporting", "Data Export", "Data Import", "Data Cleaning",
		"Data Validation", "Data Normalization", "Data Denormalization", "Data Modeling",
		"Data Warehousing", "Data Mining", "Data Science", "Machine Learning",
		"Artificial Intelligence", "Natural Language Processing", "Computer Vision", "Robotics",
		"Blockchain", "Cryptography", "Cybersecurity", "Network Security",
		"Application Security", "Cloud Security", "DevSecOps", "Security Audits",
		"Penetration Testing", "Vulnerability Assessment", "Security Patching", "Security Monitoring",
		"Security Incident Response", "Security Compliance", "Security Policies", "Security Training",
		"Security Awareness", "Security Best Practices", "Security Frameworks", "Security Standards",
		"Security Certifications", "Security Tools", "Security Technologies", "Security Research",
	}

	noteContents = []string{
		"This is a quick note about implementation details.",
		"Remember to check edge cases and error handling.",
		"TODO: Review this section before deployment.",
		"Key insight: Always validate user input.",
		"Performance consideration: Use indexes for frequent queries.",
		"Security note: Never expose sensitive data in logs.",
		"Best practice: Keep functions small and focused.",
		"Refactoring idea: Extract this logic into a separate service.",
		"Bug found: Need to handle null values properly.",
		"Feature request: Add pagination to this endpoint.",
		"Code review feedback: Consider using a more descriptive variable name.",
		"Testing note: Add integration tests for this flow.",
		"Documentation needed: Explain the algorithm used here.",
		"Optimization: Cache this expensive operation.",
		"Architecture decision: Use event-driven approach for this feature.",
		"Meeting outcome: Agreed to implement this in next sprint.",
		"User feedback: Interface is confusing, needs improvement.",
		"Technical debt: Refactor this legacy code when time permits.",
		"Deployment note: Remember to run migrations first.",
		"Monitoring: Set up alerts for this critical path.",
		"Consider using a connection pool for database connections.",
		"Always sanitize user input to prevent SQL injection.",
		"Use prepared statements for database queries.",
		"Implement proper error handling for all external calls.",
		"Use environment variables for configuration.",
		"Document all public APIs and their usage.",
		"Use meaningful commit messages for better traceability.",
		"Implement proper logging for debugging and auditing.",
		"Use a linter to enforce code style and quality.",
		"Write unit tests for all critical functions.",
		"Use a task runner for automation.",
		"Implement proper access control for sensitive operations.",
		"Use HTTPS for all external communications.",
		"Implement rate limiting to prevent abuse.",
		"Use a CDN for static assets.",
		"Implement proper caching for performance.",
		"Use a message broker for asynchronous processing.",
		"Implement proper retry logic for transient failures.",
		"Use a configuration management tool for deployment.",
		"Implement proper health checks for monitoring.",
		"Use a container orchestration tool for scaling.",
		"Implement proper secrets management.",
		"Use a CI/CD pipeline for automation.",
		"Implement proper backup and restore procedures.",
		"Use a monitoring tool for observability.",
		"Implement proper alerting for critical issues.",
		"Use a logging framework for structured logging.",
		"Implement proper metrics for performance monitoring.",
		"Use a tracing tool for distributed tracing.",
		"Implement proper documentation for onboarding.",
		"Use a collaboration tool for team communication.",
		"Implement proper code review processes.",
		"Use a project management tool for tracking.",
		"Implement proper sprint planning and retrospectives.",
		"Use a version control system for collaboration.",
		"Implement proper branching and merging strategies.",
		"Use a code coverage tool for testing.",
		"Implement proper security scanning for vulnerabilities.",
		"Use a dependency management tool for libraries.",
		"Implement proper license compliance for dependencies.",
		"Use a static analysis tool for code quality.",
		"Implement proper performance profiling for optimization.",
		"Use a load testing tool for scalability.",
		"Implement proper disaster recovery planning.",
		"Use a secrets management tool for credentials.",
		"Implement proper compliance checks for regulations.",
		"Use a policy as code tool for governance.",
		"Implement proper audit logging for accountability.",
		"Use a security information and event management tool for monitoring.",
		"Implement proper incident response planning for security.",
		"Use a threat modeling tool for risk assessment.",
		"Implement proper security training for awareness.",
		"Use a security best practices guide for reference.",
		"Implement proper security frameworks for compliance.",
		"Use a security standards reference for guidance.",
		"Implement proper security certifications for validation.",
		"Use a security tools suite for protection.",
		"Implement proper security technologies for defense.",
		"Use a security research repository for knowledge.",
	}

	notebookNames = []string{
		"Backend Development",
		"Frontend Development",
		"DevOps & Infrastructure",
		"Database Design",
		"API Documentation",
		"Security & Authentication",
		"Testing & QA",
		"Project Management",
		"Research & Learning",
		"Bug Tracking",
	}

	chapterNames = []string{
		"Getting Started", "Core Features", "Advanced Topics", "Best Practices",
		"Common Issues", "Performance", "Security", "Testing", "Deployment", "Maintenance",
		"Architecture", "Data Modeling", "Integration", "Monitoring", "Optimization",
	}
)

func generateCUID() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 24)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func generateNoteContent() string {
	title := noteTitles[rand.Intn(len(noteTitles))]
	content := noteContents[rand.Intn(len(noteContents))]
	return fmt.Sprintf(`{"type":"doc","content":[{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"%s"}]},{"type":"paragraph","content":[{"type":"text","text":"%s"}]}]}`, title, content)
}

func main() {
	// Load environment variables
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	// Initialize database
	db.InitDB()
	log.Println("Starting seed data generation...")

	// Create notebooks
	notebooks := make([]models.Notebook, 0, 10)
	for i := 0; i < 10; i++ {
		notebook := models.Notebook{
			ID:             generateCUID(),
			Name:           notebookNames[i],
			ClerkUserID:    userID,
			OrganizationID: &orgID,
			IsPublic:       false,
			CreatedAt:      time.Now().Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour),
			UpdatedAt:      time.Now(),
		}
		if err := db.DB.Create(&notebook).Error; err != nil {
			log.Printf("Error creating notebook %s: %v", notebook.Name, err)
			continue
		}
		notebooks = append(notebooks, notebook)
		log.Printf("Created notebook: %s (ID: %s)", notebook.Name, notebook.ID)
	}

	// Create chapters (15 total, distributed across notebooks)
	chapters := make([]models.Chapter, 0, 15)
	for i := 0; i < 15; i++ {
		notebook := notebooks[i%len(notebooks)]
		chapter := models.Chapter{
			ID:             generateCUID(),
			Name:           chapterNames[i%len(chapterNames)],
			NotebookID:     notebook.ID,
			OrganizationID: &orgID,
			IsPublic:       false,
			CreatedAt:      time.Now().Add(-time.Duration(rand.Intn(20)) * 24 * time.Hour),
			UpdatedAt:      time.Now(),
		}
		if err := db.DB.Create(&chapter).Error; err != nil {
			log.Printf("Error creating chapter %s: %v", chapter.Name, err)
			continue
		}
		chapters = append(chapters, chapter)
		log.Printf("Created chapter: %s in notebook %s (ID: %s)", chapter.Name, notebook.Name, chapter.ID)
	}

	// Create 150 notes distributed across chapters
	notesPerChapter := 150 / len(chapters)
	remainder := 150 % len(chapters)
	noteCount := 0

	for i, chapter := range chapters {
		numNotes := notesPerChapter
		if i < remainder {
			numNotes++
		}
		for j := 0; j < numNotes; j++ {
			note := models.Notes{
				ID:             generateCUID(),
				Name:           noteTitles[rand.Intn(len(noteTitles))],
				Content:        generateNoteContent(),
				ChapterID:      chapter.ID,
				OrganizationID: &orgID,
				IsPublic:       false,
				HasVideo:       rand.Intn(2) == 1,
				CreatedAt:      time.Now().Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour),
				UpdatedAt:      time.Now().Add(-time.Duration(rand.Intn(7)) * 24 * time.Hour),
			}
			if err := db.DB.Create(&note).Error; err != nil {
				log.Printf("Error creating note %s: %v", note.Name, err)
				continue
			}
			noteCount++
			if noteCount%10 == 0 {
				log.Printf("Created %d notes...", noteCount)
			}
		}
	}

	log.Printf("\n✅ Seed data generation complete!")
	log.Printf("📊 Summary:")
	log.Printf("   - Notebooks: %d", len(notebooks))
	log.Printf("   - Chapters: %d", len(chapters))
	log.Printf("   - Notes: %d", noteCount)
	log.Printf("\n🔑 User ID: %s", userID)
	log.Printf("🏢 Organization ID: %s", orgID)
}
