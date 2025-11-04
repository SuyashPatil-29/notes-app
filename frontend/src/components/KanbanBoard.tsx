import React, { useState, useEffect } from "react";
import { Plus, Trash, Flame, Edit, MoreHorizontal, Users, User } from "lucide-react";
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
  DropdownMenuSeparator,
} from "@/components/ui/dropdown-menu";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { toast } from "sonner";
import type { Task, TaskBoard, OrganizationMemberForAssignment } from "@/types/backend";
import { updateTask, deleteTask, createTask, assignTaskToUsers, getOrganizationMembers } from "@/utils/tasks";
import TaskDetailModal from "./TaskDetailModal";

type ColumnType = "backlog" | "todo" | "in_progress" | "done";

interface KanbanBoardProps {
  taskBoard: TaskBoard;
  onTaskUpdate?: (updatedTask: Task) => void;
  onTaskDelete?: (taskId: string) => void;
  onTaskCreate?: (newTask: Task) => void;
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
  onTaskAssign: (task: Task) => void;
  organizationMembers: OrganizationMemberForAssignment[];
  isFirst?: boolean;
  isLast?: boolean;
}

interface TaskCardProps {
  task: Task;
  handleDragStart: (e: React.DragEvent<HTMLDivElement>, task: Task) => void;
  onEdit: (task: Task) => void;
  onDelete: (taskId: string) => void;
  onAssign: (task: Task) => void;
  organizationMembers: OrganizationMemberForAssignment[];
}

interface DropIndicatorProps {
  beforeId: string | null;
  column: ColumnType;
}

interface BurnBarrelProps {
  onDelete: (taskId: string) => void;
}

interface AddTaskProps {
  column: ColumnType;
  taskBoard: TaskBoard;
  onTaskCreate: (newTask: Task) => void;
  organizationMembers: OrganizationMemberForAssignment[];
}

interface EditTaskDialogProps {
  task: Task | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: (task: Task) => void;
}

export const KanbanBoard: React.FC<KanbanBoardProps> = ({
  taskBoard,
  onTaskUpdate,
  onTaskDelete,
  onTaskCreate,
  className = "",
}) => {
  const [tasks, setTasks] = useState<Task[]>(taskBoard.tasks || []);
  const taskBoardIdRef = React.useRef(taskBoard.id);
  const [selectedTaskForAssignment, setSelectedTaskForAssignment] = useState<Task | null>(null);
  const [isAssignmentModalOpen, setIsAssignmentModalOpen] = useState(false);
  const [organizationMembers, setOrganizationMembers] = useState<OrganizationMemberForAssignment[]>([]);

  // Only sync when the board ID changes (switching to a different board)
  // Don't sync on regular re-renders to preserve optimistic updates
  useEffect(() => {
    if (taskBoardIdRef.current !== taskBoard.id) {
      taskBoardIdRef.current = taskBoard.id;
      setTasks(taskBoard.tasks || []);
    }
  }, [taskBoard.id, taskBoard.tasks]);

  // Load organization members when taskBoard has an organization
  useEffect(() => {
    const loadOrganizationMembers = async () => {
      if (taskBoard.organizationId) {
        try {
          const response = await getOrganizationMembers(taskBoard.organizationId);
          setOrganizationMembers(response.members);
        } catch (error) {
          console.error("Failed to load organization members:", error);
          toast.error("Failed to load organization members");
        }
      }
    };

    loadOrganizationMembers();
  }, [taskBoard.organizationId]);

  const handleTaskUpdate = async (updatedTask: Task) => {
    // Optimistically update local state immediately
    const previousTasks = tasks;
    setTasks((prev) => prev.map((task) => (task.id === updatedTask.id ? updatedTask : task)));

    // Notify parent component optimistically
    onTaskUpdate?.(updatedTask);

    try {
      // Update backend
      const result = await updateTask(updatedTask.id, {
        title: updatedTask.title,
        description: updatedTask.description,
        status: updatedTask.status,
        priority: updatedTask.priority,
        position: updatedTask.position,
      });

      // Update with server response
      setTasks((prev) => prev.map((task) => (task.id === updatedTask.id ? result : task)));
      onTaskUpdate?.(result);
    } catch (error) {
      console.error("Failed to update task:", error);
      toast.error("Failed to update task");
      
      // Rollback on error
      setTasks(previousTasks);
      onTaskUpdate?.(previousTasks.find(t => t.id === updatedTask.id)!);
    }
  };

  const handleTaskDelete = async (taskId: string) => {
    // Optimistically remove task immediately
    const previousTasks = tasks;
    const deletedTask = tasks.find(t => t.id === taskId);
    setTasks((prev) => prev.filter((task) => task.id !== taskId));

    // Notify parent component optimistically
    onTaskDelete?.(taskId);

    try {
      // Delete from backend
      await deleteTask(taskId);
      toast.success("Task deleted successfully");
    } catch (error) {
      console.error("Failed to delete task:", error);
      toast.error("Failed to delete task");
      
      // Rollback on error
      if (deletedTask) {
        setTasks(previousTasks);
        onTaskCreate?.(deletedTask);
      }
    }
  };

  const handleTaskCreate = async (newTask: Task) => {
    // If this is a real task (not temporary), replace any temporary task
    if (!newTask.id.startsWith('temp-')) {
      setTasks((prev) => {
        // Remove temporary task if exists and add real task
        const withoutTemp = prev.filter(t => !t.id.startsWith('temp-'));
        return [...withoutTemp, newTask];
      });
    } else {
      // Optimistically add temporary task immediately
      setTasks((prev) => [...prev, newTask]);
    }

    // Notify parent component
    onTaskCreate?.(newTask);
  };

  const handleOpenAssignmentModal = (task: Task) => {
    setSelectedTaskForAssignment(task);
    setIsAssignmentModalOpen(true);
  };

  const handleCloseAssignmentModal = () => {
    setIsAssignmentModalOpen(false);
    setSelectedTaskForAssignment(null);
  };

  const handleAssignUsers = async (taskId: string, userIds: string[]) => {
    try {
      await assignTaskToUsers(taskId, userIds);

      // Update the task with new assignments (optimistic update)
      const updatedAssignments = userIds.map((userId) => ({
        id: `temp-${userId}`,
        taskId,
        userId,
        createdAt: new Date().toISOString(),
      }));

      setTasks((prev) =>
        prev.map((task) =>
          task.id === taskId
            ? { ...task, assignments: updatedAssignments }
            : task
        )
      );

      // Update selected task for modal
      if (selectedTaskForAssignment && selectedTaskForAssignment.id === taskId) {
        setSelectedTaskForAssignment({
          ...selectedTaskForAssignment,
          assignments: updatedAssignments,
        });
      }

      toast.success("Task assignments updated successfully");
    } catch (error) {
      console.error("Failed to assign users to task:", error);
      toast.error("Failed to update task assignments");
      throw error;
    }
  };

  return (
    <>
      <div className={`h-full w-full bg-background ${className}`}>
        <div className="flex h-full w-full gap-0 overflow-x-auto p-6 md:p-12">
        <Column
          title="Backlog"
          column="backlog"
          headingColor="text-slate-600 dark:text-slate-400"
          tasks={tasks}
          taskBoard={taskBoard}
          onTaskUpdate={handleTaskUpdate}
          onTaskDelete={handleTaskDelete}
          onTaskCreate={handleTaskCreate}
          onTaskAssign={handleOpenAssignmentModal}
          organizationMembers={organizationMembers}
          isFirst
        />
        <Column
          title="TODO"
          column="todo"
          headingColor="text-amber-600 dark:text-amber-400"
          tasks={tasks}
          taskBoard={taskBoard}
          onTaskUpdate={handleTaskUpdate}
          onTaskDelete={handleTaskDelete}
          onTaskCreate={handleTaskCreate}
          onTaskAssign={handleOpenAssignmentModal}
          organizationMembers={organizationMembers}
        />
        <Column
          title="In progress"
          column="in_progress"
          headingColor="text-blue-600 dark:text-blue-400"
          tasks={tasks}
          taskBoard={taskBoard}
          onTaskUpdate={handleTaskUpdate}
          onTaskDelete={handleTaskDelete}
          onTaskCreate={handleTaskCreate}
          onTaskAssign={handleOpenAssignmentModal}
          organizationMembers={organizationMembers}
        />
        <Column
          title="Complete"
          column="done"
          headingColor="text-green-600 dark:text-green-400"
          tasks={tasks}
          taskBoard={taskBoard}
          onTaskUpdate={handleTaskUpdate}
          onTaskDelete={handleTaskDelete}
          onTaskCreate={handleTaskCreate}
          onTaskAssign={handleOpenAssignmentModal}
          organizationMembers={organizationMembers}
          isLast
        />
          <BurnBarrel onDelete={handleTaskDelete} />
        </div>
      </div>

      {/* Task Assignment Modal */}
      <TaskDetailModal
        task={selectedTaskForAssignment}
        isOpen={isAssignmentModalOpen}
        onClose={handleCloseAssignmentModal}
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
  onTaskAssign,
  organizationMembers,
  isFirst,
  isLast,
}) => {
  const [active, setActive] = useState<boolean>(false);
  const [editingTask, setEditingTask] = useState<Task | null>(null);

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

      // Update task status
      const updatedTask = { ...taskToMove, status: column };

      // Calculate new position
      const columnTasks = tasks.filter((t) => t.status === column && t.id !== taskId);
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
    <div className={`w-72 shrink-0 ${!isLast ? 'border-r border-border/40 pr-4' : ''} ${!isFirst ? 'pl-4' : ''}`}>
      <div className="mb-3 flex items-center justify-between">
        <h3 className={`font-medium ${headingColor}`}>{title}</h3>
        <span className="rounded text-sm text-muted-foreground">{filteredTasks.length}</span>
      </div>
      <div
        onDrop={handleDragEnd}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        className={`h-full w-full transition-colors ${active ? "bg-muted/50" : "bg-muted/0"
          }`}
      >
        {filteredTasks.map((task) => {
          return (
            <TaskCard
              key={task.id}
              task={task}
              handleDragStart={handleDragStart}
              onEdit={(task) => setEditingTask(task)}
              onDelete={onTaskDelete}
              onAssign={onTaskAssign}
              organizationMembers={organizationMembers}
            />
          );
        })}
        <DropIndicator beforeId={null} column={column} />
        <AddTask column={column} taskBoard={taskBoard} onTaskCreate={onTaskCreate} organizationMembers={organizationMembers} />
      </div>

      {/* Edit Task Dialog */}
      <EditTaskDialog
        task={editingTask}
        open={!!editingTask}
        onOpenChange={(open) => !open && setEditingTask(null)}
        onSave={onTaskUpdate}
      />
    </div>
  );
};

const TaskCard: React.FC<TaskCardProps> = ({ task, handleDragStart, onEdit, onDelete, onAssign, organizationMembers }) => {
  const handleDelete = () => {
    if (confirm("Are you sure you want to delete this task?")) {
      onDelete(task.id);
    }
  };

  const handleAssign = () => {
    onAssign(task);
  };

  const getPriorityBadge = () => {
    const colors = {
      high: "bg-red-500/15 text-red-700 dark:text-red-400 border-red-500/30",
      medium: "bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-500/30",
      low: "bg-blue-500/15 text-blue-700 dark:text-blue-400 border-blue-500/30",
    };
    return colors[task.priority] || colors.medium;
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
        className="cursor-grab rounded border border-border bg-card p-3 active:cursor-grabbing mb-3 hover:border-primary/50 transition-colors group"
      >
        <div
          draggable="true"
          onDragStart={(e) => handleDragStart(e, task)}
          className="w-full"
        >
          <div className="space-y-2">
            <div className="flex items-start justify-between gap-2">
              <p className="text-sm text-foreground flex-1 leading-tight">{task.title}</p>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-6 w-6 p-0 opacity-0 group-hover:opacity-100 transition-opacity"
                  >
                    <MoreHorizontal className="h-3 w-3" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem onClick={handleAssign}>
                    <Users className="mr-2 h-4 w-4" />
                    Assign
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => onEdit(task)}>
                    <Edit className="mr-2 h-4 w-4" />
                    Edit
                  </DropdownMenuItem>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onClick={handleDelete} className="text-destructive">
                    <Trash className="mr-2 h-4 w-4" />
                    Delete
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>

            {task.description && (
              <p className="text-xs text-muted-foreground line-clamp-2">{task.description}</p>
            )}

            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Badge variant="outline" className={`text-xs ${getPriorityBadge()}`}>
                  {task.priority}
                </Badge>
                
                {/* Assigned Members Avatars */}
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
              <span className="text-xs text-muted-foreground">#{task.position}</span>
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
      className="my-0.5 h-0.5 w-full bg-primary opacity-0"
    />
  );
};

const BurnBarrel: React.FC<BurnBarrelProps> = ({ onDelete }) => {
  const [active, setActive] = useState<boolean>(false);

  const handleDragOver = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setActive(true);
  };

  const handleDragLeave = () => {
    setActive(false);
  };

  const handleDragEnd = (e: React.DragEvent<HTMLDivElement>) => {
    const taskId = e.dataTransfer.getData("taskId");

    if (confirm("Are you sure you want to delete this task?")) {
      onDelete(taskId);
    }

    setActive(false);
  };

  return (
    <div
      onDrop={handleDragEnd}
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      className={`mt-10 grid h-56 w-72 shrink-0 place-content-center rounded border text-3xl transition-colors ${active
          ? "border-destructive bg-destructive/20 text-destructive"
          : "border-border bg-muted/20 text-muted-foreground"
        }`}
    >
      {active ? <Flame className="animate-bounce" /> : <Trash />}
    </div>
  );
};

const AddTask: React.FC<AddTaskProps> = ({ column, taskBoard, onTaskCreate, organizationMembers }) => {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [adding, setAdding] = useState(false);
  const [showAssignDialog, setShowAssignDialog] = useState(false);
  const [pendingTask, setPendingTask] = useState<Task | null>(null);
  const [selectedUserIds, setSelectedUserIds] = useState<string[]>([]);
  const [isCreating, setIsCreating] = useState(false);

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
    const titleValue = title.trim();
    const descriptionValue = description.trim();
    
    try {
      // Create task in backend
      const newTask = await createTask(taskBoard.id, {
        title: titleValue,
        description: descriptionValue,
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
      
      // Reset fields
      setTitle("");
      setDescription("");
      setAdding(false);
    } catch (error) {
      console.error("Failed to create task:", error);
      toast.error("Failed to create task - please try again");
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
        <motion.form layout onSubmit={handleSubmit} className="space-y-2">
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            autoFocus
            placeholder="Task title..."
            className="w-full rounded border border-primary bg-primary/20 p-3 text-sm text-foreground placeholder:text-muted-foreground focus:outline-0"
          />
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Description (optional)..."
            rows={2}
            className="w-full rounded border border-primary bg-primary/20 p-3 text-sm text-foreground placeholder:text-muted-foreground focus:outline-0 resize-none"
          />
          <div className="flex items-center justify-end gap-1.5">
            <button
              type="button"
              onClick={() => setAdding(false)}
              className="px-3 py-1.5 text-xs text-muted-foreground transition-colors hover:text-foreground"
              disabled={isCreating}
            >
              Close
            </button>
            <button
              type="submit"
              className="flex items-center gap-1.5 rounded bg-primary px-3 py-1.5 text-xs text-primary-foreground transition-colors hover:bg-primary/90"
              disabled={!title.trim() || isCreating}
            >
              <span>{isCreating ? 'Creating...' : 'Add'}</span>
              <Plus className="h-3 w-3" />
            </button>
          </div>
        </motion.form>
      ) : (
        <motion.button
          layout
          onClick={() => setAdding(true)}
          className="flex w-full items-center gap-1.5 px-3 py-1.5 text-xs text-muted-foreground transition-colors hover:text-foreground"
        >
          <span>Add card</span>
          <Plus className="h-4 w-4" />
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

const EditTaskDialog: React.FC<EditTaskDialogProps> = ({ task, open, onOpenChange, onSave }) => {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [priority, setPriority] = useState<Task["priority"]>("medium");

  useEffect(() => {
    if (task) {
      setTitle(task.title);
      setDescription(task.description || "");
      setPriority(task.priority);
    }
  }, [task]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!task || !title.trim()) return;

    const updatedTask = {
      ...task,
      title: title.trim(),
      description: description.trim(),
      priority,
    };

    // Optimistically update and close dialog
    onSave(updatedTask);
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit Task</DialogTitle>
          <DialogDescription>Update task details and priority.</DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <Label htmlFor="title">Title</Label>
            <Input
              id="title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Task title"
              className="mt-1"
            />
          </div>

          <div>
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Task description (optional)"
              rows={3}
              className="mt-1"
            />
          </div>

          <div>
            <Label htmlFor="priority">Priority</Label>
            <select
              id="priority"
              value={priority}
              onChange={(e) => setPriority(e.target.value as Task["priority"])}
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring mt-1"
            >
              <option value="low">Low Priority</option>
              <option value="medium">Medium Priority</option>
              <option value="high">High Priority</option>
            </select>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={!title.trim()}>
              Update Task
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
};

