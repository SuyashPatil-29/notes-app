import React, { useState, useEffect } from "react";
import { Plus, Trash, Edit, MoreHorizontal, Users, User } from "lucide-react";
import { motion } from "framer-motion";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { toast } from "sonner";
import type {
  Task,
  TaskBoard,
  OrganizationMemberForAssignment,
} from "@/types/backend";
import {
  updateTask,
  deleteTask,
  createTask,
  assignTaskToUsers,
  getOrganizationMembers,
} from "@/utils/tasks";
import { getTaskPriorityColor } from "@/utils/tasks";
import TaskDetailModal from "./TaskDetailModal";

type ColumnType = "backlog" | "todo" | "in_progress" | "done";

interface KanbanViewProps {
  taskBoard: TaskBoard;
  onTaskUpdate?: (updatedTask: Task) => void;
  onTaskDelete?: (taskId: string) => void;
  onTaskCreate?: (newTask: Task) => void;
  onTaskClick?: (task: Task) => void;
  className?: string;
}

interface ColumnProps {
  title: string;
  headingColor: string;
  column: ColumnType;
  tasks: Task[];
  taskBoard: TaskBoard;
  onTaskUpdate: (updatedTask: Task) => void;
  onTaskDelete: (taskId: string) => void;
  onTaskCreate: (newTask: Task) => void;
  onTaskClick: (task: Task) => void;
  organizationMembers: OrganizationMemberForAssignment[];
}

interface TaskCardProps {
  task: Task;
  handleDragStart: (e: React.DragEvent<HTMLDivElement>, task: Task) => void;
  onEdit: (task: Task) => void;
  onDelete: (taskId: string) => void;
  onClick: (task: Task) => void;
  organizationMembers: OrganizationMemberForAssignment[];
}

interface DropIndicatorProps {
  beforeId: string | null;
  column: ColumnType;
}

interface AddTaskProps {
  column: ColumnType;
  taskBoard: TaskBoard;
  onTaskCreate: (newTask: Task) => void;
  organizationMembers: OrganizationMemberForAssignment[];
}

export const KanbanView: React.FC<KanbanViewProps> = ({
  taskBoard,
  onTaskUpdate,
  onTaskDelete,
  onTaskCreate,
  onTaskClick,
  className = "",
}) => {
  const [tasks, setTasks] = useState<Task[]>(taskBoard.tasks || []);
  const [selectedTask, setSelectedTask] = useState<Task | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [organizationMembers, setOrganizationMembers] = useState<
    OrganizationMemberForAssignment[]
  >([]);
  const [, setIsLoadingMembers] = useState(false);

  // Update local tasks when taskBoard changes
  useEffect(() => {
    setTasks(taskBoard.tasks || []);
  }, [taskBoard.tasks]);

  // Load organization members when taskBoard has an organization
  useEffect(() => {
    const loadOrganizationMembers = async () => {
      if (taskBoard.organizationId) {
        setIsLoadingMembers(true);
        try {
          const response = await getOrganizationMembers(
            taskBoard.organizationId
          );
          setOrganizationMembers(response.members);
        } catch (error) {
          console.error("Failed to load organization members:", error);
          toast.error("Failed to load organization members");
        } finally {
          setIsLoadingMembers(false);
        }
      }
    };

    loadOrganizationMembers();
  }, [taskBoard.organizationId]);

  const handleTaskUpdate = async (updatedTask: Task) => {
    try {
      const result = await updateTask(updatedTask.id, {
        title: updatedTask.title,
        description: updatedTask.description,
        status: updatedTask.status,
        priority: updatedTask.priority,
        position: updatedTask.position,
      });

      // Update local state
      setTasks((prev) =>
        prev.map((task) => (task.id === updatedTask.id ? result : task))
      );

      // Notify parent component
      onTaskUpdate?.(result);
    } catch (error) {
      console.error("Failed to update task:", error);
      toast.error("Failed to update task");
    }
  };

  const handleTaskDelete = async (taskId: string) => {
    try {
      await deleteTask(taskId);

      // Update local state
      setTasks((prev) => prev.filter((task) => task.id !== taskId));

      // Notify parent component
      onTaskDelete?.(taskId);
      toast.success("Task deleted successfully");
    } catch (error) {
      console.error("Failed to delete task:", error);
      toast.error("Failed to delete task");
    }
  };

  const handleTaskCreate = async (newTask: Task) => {
    // Update local state
    setTasks((prev) => [...prev, newTask]);

    // Notify parent component
    onTaskCreate?.(newTask);
  };

  const handleTaskClick = (task: Task) => {
    setSelectedTask(task);
    setIsModalOpen(true);
    onTaskClick?.(task);
  };

  const handleModalClose = () => {
    setIsModalOpen(false);
    setSelectedTask(null);
  };

  const handleAssignUsers = async (taskId: string, userIds: string[]) => {
    try {
      await assignTaskToUsers(taskId, userIds);

      // Refresh the task to get updated assignments
      // In a real app, you might want to optimistically update the UI
      // For now, we'll just show a success message
      toast.success("Task assignments updated successfully");

      // Update the selected task with new assignments (optimistic update)
      if (selectedTask && selectedTask.id === taskId) {
        const updatedAssignments = userIds.map((userId) => ({
          id: `temp-${userId}`, // Temporary ID
          taskId,
          userId,
          createdAt: new Date().toISOString(),
        }));

        setSelectedTask({
          ...selectedTask,
          assignments: updatedAssignments,
        });

        // Also update the task in the tasks list
        setTasks((prev) =>
          prev.map((task) =>
            task.id === taskId
              ? { ...task, assignments: updatedAssignments }
              : task
          )
        );
      }
    } catch (error) {
      console.error("Failed to assign users to task:", error);
      toast.error("Failed to update task assignments");
      throw error; // Re-throw to let the modal handle the error
    }
  };

  return (
    <>
      <div className={`h-full w-full bg-background ${className}`}>
        <div className="flex h-full w-full gap-6 overflow-x-auto p-6">
          <Column
            title="To Do"
            column="todo"
            headingColor="text-blue-600"
            tasks={tasks}
            taskBoard={taskBoard}
            onTaskUpdate={handleTaskUpdate}
            onTaskDelete={handleTaskDelete}
            onTaskCreate={handleTaskCreate}
            onTaskClick={handleTaskClick}
            organizationMembers={organizationMembers}
          />
          <Column
            title="In Progress"
            column="in_progress"
            headingColor="text-orange-600"
            tasks={tasks}
            taskBoard={taskBoard}
            onTaskUpdate={handleTaskUpdate}
            onTaskDelete={handleTaskDelete}
            onTaskCreate={handleTaskCreate}
            onTaskClick={handleTaskClick}
            organizationMembers={organizationMembers}
          />
          <Column
            title="Done"
            column="done"
            headingColor="text-green-600"
            tasks={tasks}
            taskBoard={taskBoard}
            onTaskUpdate={handleTaskUpdate}
            onTaskDelete={handleTaskDelete}
            onTaskCreate={handleTaskCreate}
            onTaskClick={handleTaskClick}
            organizationMembers={organizationMembers}
          />
        </div>
      </div>

      {/* Task Detail Modal */}
      <TaskDetailModal
        task={selectedTask}
        isOpen={isModalOpen}
        onClose={handleModalClose}
        organizationMembers={organizationMembers}
        onAssignUsers={handleAssignUsers}
      />
    </>
  );
};

const Column: React.FC<ColumnProps> = ({
  title,
  headingColor,
  tasks,
  column,
  taskBoard,
  onTaskUpdate,
  onTaskDelete,
  onTaskCreate,
  onTaskClick,
  organizationMembers,
}) => {
  const [active, setActive] = useState<boolean>(false);

  const handleDragStart = (e: React.DragEvent<HTMLDivElement>, task: Task) => {
    e.dataTransfer.setData("taskId", task.id);
  };

  const handleDragEnd = async (e: React.DragEvent<HTMLDivElement>) => {
    const taskId = e.dataTransfer.getData("taskId");

    setActive(false);
    clearHighlights();

    const indicators = getIndicators();
    const { element } = getNearestIndicator(e, indicators);

    const before = element.dataset.before || "-1";

    if (before !== taskId) {
      const taskToMove = tasks.find((t) => t.id === taskId);
      if (!taskToMove) return;

      // Update task status and position
      const updatedTask = { ...taskToMove, status: column };

      // Calculate new position
      const columnTasks = tasks.filter((t) => t.status === column);
      let newPosition = 1;

      if (before === "-1") {
        // Move to end
        newPosition = columnTasks.length + 1;
      } else {
        // Move before specific task
        const beforeTask = tasks.find((t) => t.id === before);
        if (beforeTask) {
          newPosition = beforeTask.position;
        }
      }

      updatedTask.position = newPosition;
      await onTaskUpdate(updatedTask);
    }
  };

  const handleDragOver = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    highlightIndicator(e);
    setActive(true);
  };

  const clearHighlights = (els?: HTMLElement[]) => {
    const indicators = els || getIndicators();
    indicators.forEach((i) => {
      i.style.opacity = "0";
    });
  };

  const highlightIndicator = (e: React.DragEvent<HTMLDivElement>) => {
    const indicators = getIndicators();
    clearHighlights(indicators);
    const el = getNearestIndicator(e, indicators);
    el.element.style.opacity = "1";
  };

  const getNearestIndicator = (
    e: React.DragEvent<HTMLDivElement>,
    indicators: HTMLElement[]
  ) => {
    const DISTANCE_OFFSET = 50;

    const el = indicators.reduce(
      (closest, child) => {
        const box = child.getBoundingClientRect();
        const offset = e.clientY - (box.top + DISTANCE_OFFSET);

        if (offset < 0 && offset > closest.offset) {
          return { offset: offset, element: child };
        } else {
          return closest;
        }
      },
      {
        offset: Number.NEGATIVE_INFINITY,
        element: indicators[indicators.length - 1],
      }
    );

    return el;
  };

  const getIndicators = (): HTMLElement[] => {
    return Array.from(document.querySelectorAll(`[data-column="${column}"]`));
  };

  const handleDragLeave = () => {
    clearHighlights();
    setActive(false);
  };

  const filteredTasks = tasks.filter((t) => t.status === column);

  return (
    <div className="w-80 shrink-0">
      <div className="mb-4 flex items-center justify-between">
        <h3 className={`font-semibold text-lg ${headingColor}`}>{title}</h3>
        <Badge variant="secondary" className="text-xs">
          {filteredTasks.length}
        </Badge>
      </div>
      <div
        onDrop={handleDragEnd}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        className={`min-h-[500px] w-full rounded-lg border-2 border-dashed transition-colors p-2 ${
          active
            ? "border-primary bg-primary/5"
            : "border-muted-foreground/25 bg-muted/20"
        }`}
      >
        {filteredTasks.map((task) => {
          return (
            <TaskCard
              key={task.id}
              task={task}
              handleDragStart={handleDragStart}
              onEdit={onTaskUpdate}
              onDelete={onTaskDelete}
              onClick={onTaskClick}
              organizationMembers={organizationMembers}
            />
          );
        })}
        <DropIndicator beforeId={null} column={column} />
        <AddTask
          column={column}
          taskBoard={taskBoard}
          onTaskCreate={onTaskCreate}
          organizationMembers={organizationMembers}
        />
      </div>
    </div>
  );
};

const TaskCard: React.FC<TaskCardProps> = ({
  task,
  handleDragStart,
  onEdit,
  onDelete,
  onClick,
  organizationMembers,
}) => {
  const handleEdit = () => {
    // For now, we'll just allow editing the title inline
    // In a full implementation, this would open a modal or form
    const newTitle = prompt("Edit task title:", task.title);
    if (newTitle && newTitle !== task.title) {
      onEdit({ ...task, title: newTitle });
    }
  };

  const handleDelete = () => {
    if (confirm("Are you sure you want to delete this task?")) {
      onDelete(task.id);
    }
  };

  const handleAssign = () => {
    console.log("Assign clicked for task:", task.title);
    onClick(task);
  };

  // Get assigned members with their details
  const getAssignedMembers = () => {
    if (!task.assignments || task.assignments.length === 0) return [];
    
    return task.assignments
      .map(assignment => 
        organizationMembers.find(member => member.id === assignment.userId)
      )
      .filter(Boolean) as OrganizationMemberForAssignment[];
  };

  const assignedMembers = getAssignedMembers();

  return (
    <>
      <DropIndicator beforeId={task.id} column={task.status} />
      <motion.div
        layout
        layoutId={task.id}
        className="cursor-grab rounded-lg border bg-card mb-3 shadow-sm hover:shadow-md transition-shadow active:cursor-grabbing"
      >
        <div
          draggable="true"
          onDragStart={(e) => handleDragStart(e, task)}
          className="p-4 w-full"
        >
          <div className="space-y-3">
            <div className="flex items-start justify-between gap-2">
              <h4 className="font-medium text-sm leading-tight flex-1">
                {task.title}
              </h4>
              <DropdownMenu>
                <DropdownMenuTrigger asChild data-dropdown-trigger>
                  <Button variant="ghost" size="sm" className="h-6 w-6 p-0">
                    <MoreHorizontal className="h-3 w-3" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem
                    onClick={handleAssign}
                    className="font-medium"
                  >
                    <Users className="mr-2 h-4 w-4 text-blue-600" />
                    Assign Task
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={handleEdit}>
                    <Edit className="mr-2 h-4 w-4" />
                    Edit Task
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onClick={handleDelete}
                    className="text-destructive"
                  >
                    <Trash className="mr-2 h-4 w-4" />
                    Delete
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>

            {task.description && (
              <p className="text-xs text-muted-foreground line-clamp-2">
                {task.description}
              </p>
            )}

            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Badge
                  variant="outline"
                  className={`text-xs ${getTaskPriorityColor(task.priority)}`}
                >
                  {task.priority}
                </Badge>

                {/* Assignment Avatars */}
                {assignedMembers.length > 0 && (
                  <TooltipProvider>
                    <div className="flex -space-x-1">
                      {assignedMembers.slice(0, 3).map((member) => (
                        <Tooltip key={member.id} delayDuration={200}>
                          <TooltipTrigger asChild>
                            <div className="relative cursor-pointer">
                              {member.imageUrl ? (
                                <img
                                  src={member.imageUrl}
                                  alt={member.name}
                                  className="w-6 h-6 rounded-full border-2 border-card bg-card"
                                />
                              ) : (
                                <div className="w-6 h-6 rounded-full border-2 border-card bg-muted flex items-center justify-center">
                                  <User className="w-3 h-3 text-muted-foreground" />
                                </div>
                              )}
                            </div>
                          </TooltipTrigger>
                          <TooltipContent>
                            <p>{member.name}</p>
                          </TooltipContent>
                        </Tooltip>
                      ))}
                      {assignedMembers.length > 3 && (
                        <Tooltip delayDuration={200}>
                          <TooltipTrigger asChild>
                            <div className="w-6 h-6 rounded-full border-2 border-card bg-muted flex items-center justify-center cursor-pointer">
                              <span className="text-[10px] font-medium text-muted-foreground">
                                +{assignedMembers.length - 3}
                              </span>
                            </div>
                          </TooltipTrigger>
                          <TooltipContent>
                            <div className="space-y-1">
                              {assignedMembers.slice(3).map((member) => (
                                <p key={member.id} className="text-sm">{member.name}</p>
                              ))}
                            </div>
                          </TooltipContent>
                        </Tooltip>
                      )}
                    </div>
                  </TooltipProvider>
                )}
              </div>

              <span className="text-xs text-muted-foreground">
                #{task.position}
              </span>
            </div>
          </div>
        </div>
      </motion.div>
    </>
  );
};

const DropIndicator: React.FC<DropIndicatorProps> = ({ beforeId, column }) => {
  return (
    <div
      data-before={beforeId || "-1"}
      data-column={column}
      className="my-0.5 h-0.5 w-full bg-primary opacity-0 rounded"
    />
  );
};

const AddTask: React.FC<AddTaskProps> = ({ column, taskBoard, onTaskCreate, organizationMembers }) => {
  const [title, setTitle] = useState<string>("");
  const [description, setDescription] = useState<string>("");
  const [adding, setAdding] = useState<boolean>(false);
  const [showAssignDialog, setShowAssignDialog] = useState(false);
  const [pendingTask, setPendingTask] = useState<Task | null>(null);
  const [selectedUserIds, setSelectedUserIds] = useState<string[]>([]);
  const [isCreating, setIsCreating] = useState<boolean>(false);

  const handleUserToggle = (userId: string) => {
    setSelectedUserIds(prev => 
      prev.includes(userId) 
        ? prev.filter(id => id !== userId)
        : [...prev, userId]
    );
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!title.trim().length) return;

    setIsCreating(true);

    try {
      const newTask = await createTask(taskBoard.id, {
        title: title.trim(),
        description: description.trim(),
        status: column,
        priority: "medium",
      });

      // Store pending task and show assign dialog if members are available
      if (organizationMembers.length > 0) {
        setPendingTask(newTask);
        setShowAssignDialog(true);
      } else {
        // No members to assign, just add the task
        onTaskCreate(newTask);
        toast.success("Task created successfully");
      }
      
      setTitle("");
      setDescription("");
      setAdding(false);
    } catch (error) {
      console.error("Failed to create task:", error);
      toast.error("Failed to create task");
    } finally {
      setIsCreating(false);
    }
  };

  const handleAssignAndClose = async () => {
    if (!pendingTask) return;

    // Add assignments if any selected
    if (selectedUserIds.length > 0) {
      try {
        await assignTaskToUsers(pendingTask.id, selectedUserIds);
        // Add assignments to the task object for optimistic update
        pendingTask.assignments = selectedUserIds.map(userId => ({
          id: `temp-${userId}`,
          taskId: pendingTask.id,
          userId,
          createdAt: new Date().toISOString(),
        }));
        toast.success("Task created and assigned successfully");
      } catch (assignError) {
        console.error("Failed to assign users:", assignError);
        toast.error("Task created but failed to assign users");
      }
    } else {
      toast.success("Task created successfully");
    }

    // Add task to UI
    onTaskCreate(pendingTask);
    
    // Reset state
    setPendingTask(null);
    setSelectedUserIds([]);
    setShowAssignDialog(false);
  };

  const handleSkipAssignment = () => {
    if (pendingTask) {
      onTaskCreate(pendingTask);
      toast.success("Task created successfully");
    }
    setPendingTask(null);
    setSelectedUserIds([]);
    setShowAssignDialog(false);
  };

  return (
    <>
      {adding ? (
        <motion.form layout onSubmit={handleSubmit} className="space-y-3">
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            autoFocus
            placeholder="Task title..."
            className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
          />
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Description (optional)..."
            rows={2}
            className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring resize-none"
          />
          <div className="flex items-center justify-end gap-2">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => setAdding(false)}
              disabled={isCreating}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              size="sm"
              disabled={isCreating || !title.trim()}
            >
              {isCreating ? "Creating..." : "Add Task"}
            </Button>
          </div>
        </motion.form>
      ) : (
        <motion.button
          layout
          onClick={() => setAdding(true)}
          className="flex w-full items-center justify-center gap-2 rounded-md border-2 border-dashed border-muted-foreground/25 p-4 text-sm text-muted-foreground transition-colors hover:border-muted-foreground/50 hover:text-foreground"
        >
          <Plus className="h-4 w-4" />
          <span>Add task</span>
        </motion.button>
      )}

      {/* Assignment Dialog */}
      <Dialog open={showAssignDialog} onOpenChange={setShowAssignDialog}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Users className="h-5 w-5" />
              Assign Task
            </DialogTitle>
            <DialogDescription>
              Assign "{pendingTask?.title}" to team members
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-2 py-2">
            <div className="border rounded-lg max-h-[300px] overflow-y-auto">
              <div className="p-2 space-y-1">
                {organizationMembers.map(member => (
                  <label
                    key={member.id}
                    className="flex items-center gap-3 p-2 hover:bg-muted rounded-lg cursor-pointer transition-colors"
                  >
                    <input
                      type="checkbox"
                      checked={selectedUserIds.includes(member.id)}
                      onChange={() => handleUserToggle(member.id)}
                      className="w-4 h-4 rounded border-input"
                    />
                    <div className="flex items-center gap-2 flex-1">
                      {member.imageUrl ? (
                        <img
                          src={member.imageUrl}
                          alt={member.name}
                          className="w-8 h-8 rounded-full"
                        />
                      ) : (
                        <div className="w-8 h-8 bg-muted rounded-full flex items-center justify-center">
                          <User className="w-4 h-4" />
                        </div>
                      )}
                      <div className="flex-1 min-w-0">
                        <div className="text-sm font-medium truncate">
                          {member.name}
                        </div>
                        <div className="text-xs text-muted-foreground truncate">
                          {member.email}
                        </div>
                      </div>
                      {member.role === 'admin' && (
                        <Badge variant="outline" className="text-xs">
                          Admin
                        </Badge>
                      )}
                    </div>
                  </label>
                ))}
              </div>
            </div>
            {selectedUserIds.length > 0 && (
              <p className="text-xs text-muted-foreground">
                {selectedUserIds.length} member{selectedUserIds.length !== 1 ? 's' : ''} selected
              </p>
            )}
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={handleSkipAssignment}
            >
              Skip
            </Button>
            <Button
              type="button"
              onClick={handleAssignAndClose}
            >
              {selectedUserIds.length > 0 ? 'Assign' : 'Done'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
};
