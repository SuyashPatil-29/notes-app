# Design Document

## Overview

The Task Management System integrates AI-powered task generation with a Kanban board interface, providing users with intelligent task organization capabilities. The system supports both note-associated tasks (generated from note content) and standalone task boards, enabling flexible project management within the existing note-taking application.

## Architecture

### High-Level Architecture

```
Frontend (React/TypeScript)
├── Task Management Components
│   ├── TaskButton (in NoteEditor)
│   ├── KanbanView Component
│   └── TaskBoard Component
├── AI Task Generation Service
├── Task API Integration
└── Navigation Integration

Backend (Go)
├── Task Models & Database
├── Task Controllers & Routes
├── AI Task Generation Service
├── Task Repository Layer
└── Organization Context Support
```

### Data Flow

1. **Note-to-Tasks Flow**: User clicks "Create Tasks" → AI analyzes note content → Tasks generated and stored → User navigates to Kanban view
2. **Standalone Tasks Flow**: User creates new Kanban → Empty task board created → User manually adds tasks
3. **Task Management Flow**: User interacts with Kanban → Real-time updates → Database persistence → Organization sync

## Components and Interfaces

### Frontend Components

#### 1. TaskButton Component
**Location**: `frontend/src/components/TaskButton.tsx`
**Purpose**: Dynamic button in NoteEditor that switches between "Create Tasks" and "View Tasks"

```typescript
interface TaskButtonProps {
  noteId: string
  noteContent: string
  hasAssociatedTasks: boolean
  onCreateTasks: () => Promise<void>
  onViewTasks: () => void
}
```

#### 2. KanbanView Component
**Location**: `frontend/src/components/KanbanView.tsx`
**Purpose**: Main task board interface using existing Kanban UI component

```typescript
interface KanbanViewProps {
  taskBoardId: string
  noteId?: string // Optional for note-associated boards
  isStandalone: boolean
  onTaskClick: (task: Task) => void
}
```

#### 3. TaskBoard Component
**Location**: `frontend/src/components/TaskBoard.tsx`
**Purpose**: Wrapper component that handles task board logic and state management

```typescript
interface TaskBoardProps {
  boardId: string
  initialTasks: TaskColumn[]
  onTaskUpdate: (tasks: TaskColumn[]) => void
  onTaskCreate: (task: Task) => void
  onTaskDelete: (taskId: string) => void
  onTaskClick: (task: Task) => void
}
```

#### 4. TaskDetailModal Component
**Location**: `frontend/src/components/TaskDetailModal.tsx`
**Purpose**: Modal for viewing and editing task details including assignments

```typescript
interface TaskDetailModalProps {
  task: Task | null
  isOpen: boolean
  onClose: () => void
  onSave: (task: Task) => void
  organizationMembers: OrganizationMember[]
}

interface OrganizationMember {
  id: string
  name: string
  email: string
  imageUrl?: string
}
```

### Backend Models

#### 1. Task Model
**Location**: `backend/internal/models/task.model.go`

```go
type Task struct {
    ID             string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
    Title          string    `json:"title"`
    Description    string    `json:"description" gorm:"type:text"`
    Status         string    `json:"status"` // "todo", "in_progress", "done"
    Priority       string    `json:"priority"` // "low", "medium", "high"
    TaskBoardID    string    `json:"taskBoardId" gorm:"type:varchar(255);index"`
    Position       int       `json:"position"`
    OrganizationID *string   `json:"organizationId,omitempty" gorm:"type:varchar(255);index"`
    CreatedAt      time.Time `json:"createdAt"`
    UpdatedAt      time.Time `json:"updatedAt"`
    Assignments    []TaskAssignment `json:"assignments" gorm:"foreignKey:TaskID"`
}

type TaskAssignment struct {
    ID        string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
    TaskID    string    `json:"taskId" gorm:"type:varchar(255);index"`
    UserID    string    `json:"userId" gorm:"type:varchar(255);index"`
    CreatedAt time.Time `json:"createdAt"`
}
```

#### 2. TaskBoard Model
**Location**: `backend/internal/models/task_board.model.go`

```go
type TaskBoard struct {
    ID             string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
    Name           string    `json:"name"`
    Description    string    `json:"description" gorm:"type:text"`
    NoteID         *string   `json:"noteId,omitempty" gorm:"type:varchar(255);index"`
    ClerkUserID    string    `json:"clerkUserId" gorm:"type:varchar(255);index"`
    OrganizationID *string   `json:"organizationId,omitempty" gorm:"type:varchar(255);index"`
    IsStandalone   bool      `json:"isStandalone" gorm:"default:false"`
    Tasks          []Task    `json:"tasks" gorm:"foreignKey:TaskBoardID"`
    CreatedAt      time.Time `json:"createdAt"`
    UpdatedAt      time.Time `json:"updatedAt"`
}
```

#### 3. Notes Model Extension
**Location**: `backend/internal/models/notes.model.go` (extend existing)

```go
// Add to existing Notes struct
type Notes struct {
    // ... existing fields
    TaskBoardID *string `json:"taskBoardId,omitempty" gorm:"type:varchar(255);index"`
    TaskBoard   *TaskBoard `json:"taskBoard,omitempty" gorm:"foreignKey:TaskBoardID"`
}
```

### API Endpoints

#### Task Management Endpoints
**Location**: `backend/internal/controllers/task.controller.go`

```
POST   /api/notes/:noteId/tasks/generate     - Generate tasks from note content
GET    /api/notes/:noteId/tasks              - Get tasks for a note
POST   /api/kanban                           - Create standalone task board
GET    /api/kanban/:boardId                  - Get task board with tasks
PUT    /api/kanban/:boardId                  - Update task board
DELETE /api/kanban/:boardId                  - Delete task board
POST   /api/kanban/:boardId/tasks            - Create task in board
PUT    /api/tasks/:taskId                    - Update task
DELETE /api/tasks/:taskId                    - Delete task
GET    /api/user/kanban                      - List user's task boards
POST   /api/tasks/:taskId/assign             - Assign users to task
DELETE /api/tasks/:taskId/assign/:userId     - Remove user assignment from task
GET    /api/organization/members             - Get organization members for assignment
```

### AI Task Generation Service

#### Enhanced AI Service
**Location**: `backend/internal/services/ai_task_service.go`

```go
type AITaskService struct {
    client *openai.Client
    apiKeyResolver *APIKeyResolver
}

type TaskGenerationRequest struct {
    NoteTitle   string `json:"note_title"`
    NoteContent string `json:"note_content"`
    UserID      string `json:"user_id"`
    OrgID       *string `json:"org_id,omitempty"`
}

type GeneratedTask struct {
    Title       string `json:"title"`
    Description string `json:"description"`
    Priority    string `json:"priority"`
    Status      string `json:"status"`
}

type TaskGenerationResponse struct {
    BoardName string          `json:"board_name"`
    Tasks     []GeneratedTask `json:"tasks"`
}
```

## Data Models

### Task Column Structure
Tasks are organized in three default columns:
- **To Do**: New tasks and backlog items
- **In Progress**: Currently active tasks
- **Done**: Completed tasks

### Task Priority Levels
- **High**: Urgent or critical tasks
- **Medium**: Standard priority tasks
- **Low**: Nice-to-have or future tasks

### Task Status Flow
```
todo → in_progress → done
```

## Error Handling

### Frontend Error Handling
- **Network Errors**: Retry mechanism with exponential backoff
- **AI Generation Failures**: Fallback to manual task creation
- **Validation Errors**: Real-time form validation with user feedback
- **Permission Errors**: Redirect to appropriate access level

### Backend Error Handling
- **AI Service Failures**: Graceful degradation with error logging
- **Database Errors**: Transaction rollback and error reporting
- **Authorization Errors**: Consistent 403/401 responses
- **Validation Errors**: Structured error responses with field details

## Testing Strategy

### Unit Tests
- **Task Model Validation**: Test CRUD operations and relationships
- **AI Service**: Mock AI responses and test parsing logic
- **Task Controller**: Test all endpoints with various scenarios
- **Frontend Components**: Test user interactions and state management

### Integration Tests
- **End-to-End Task Flow**: Note → AI Generation → Kanban View
- **Standalone Board Creation**: Full workflow testing
- **Organization Context**: Multi-tenant task isolation
- **Real-time Updates**: WebSocket or polling mechanisms

### Performance Tests
- **AI Generation Speed**: Measure response times for various content sizes
- **Kanban Rendering**: Test with large numbers of tasks
- **Database Queries**: Optimize task retrieval and updates

## Navigation Integration

### URL Structure
```
/notebook/:notebookId/chapter/:chapterId/note/:noteId/tasks  - Note-associated tasks
/kanban/:kanbanId                                            - Standalone task board
```

### Breadcrumb Navigation
- Note Tasks: Dashboard → Notebook → Chapter → Note → Tasks
- Standalone: Dashboard → Kanban Board

### Left Sidebar Integration
- Add "Task Boards" section below notebooks
- Show standalone boards with task count indicators
- Context menu: "+ New Kanban" option

## Security Considerations

### Authorization
- **Note-Associated Tasks**: Inherit note permissions
- **Standalone Boards**: User and organization-based access control
- **Task Operations**: Verify board ownership before modifications

### Data Privacy
- **AI Processing**: Ensure note content is processed securely
- **Organization Isolation**: Strict tenant separation
- **API Key Management**: Secure storage and rotation

## Performance Optimizations

### Frontend Optimizations
- **Lazy Loading**: Load task boards on demand
- **Virtual Scrolling**: Handle large task lists efficiently
- **Optimistic Updates**: Immediate UI feedback with rollback capability
- **Caching**: Cache task boards and reduce API calls

### Backend Optimizations
- **Database Indexing**: Optimize queries for task retrieval
- **AI Response Caching**: Cache similar content analysis
- **Batch Operations**: Bulk task updates and creation
- **Connection Pooling**: Efficient database connection management

## Deployment Considerations

### Database Migrations
- Create task and task_board tables
- Add foreign key relationships
- Create necessary indexes for performance

### Feature Flags
- **AI Task Generation**: Toggle AI features per organization
- **Standalone Boards**: Enable/disable standalone functionality
- **Advanced Features**: Gradual rollout of complex features

### Monitoring
- **AI Service Health**: Monitor API usage and response times
- **Task Board Performance**: Track user engagement and load times
- **Error Rates**: Monitor and alert on task operation failures