import React, { useState, useEffect } from 'react';
import { User, UserPlus } from 'lucide-react';
import type { Task, OrganizationMemberForAssignment } from '../types/backend';
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';

interface TaskDetailModalProps {
  task: Task | null;
  isOpen: boolean;
  onClose: () => void;
  organizationMembers: OrganizationMemberForAssignment[];
  onAssignUsers: (taskId: string, userIds: string[]) => Promise<void>;
}

const TaskDetailModal: React.FC<TaskDetailModalProps> = ({
  task,
  isOpen,
  onClose,
  organizationMembers,
  onAssignUsers,
}) => {
  const [selectedUserIds, setSelectedUserIds] = useState<string[]>([]);
  const [isSaving, setIsSaving] = useState(false);

  // Initialize selected users when task changes
  useEffect(() => {
    if (task) {
      setSelectedUserIds(task.assignments?.map(a => a.userId) || []);
    }
  }, [task]);

  // Reset form when modal closes
  useEffect(() => {
    if (!isOpen) {
      setSelectedUserIds([]);
      setIsSaving(false);
    }
  }, [isOpen]);

  if (!isOpen || !task) {
    return null;
  }

  const handleSave = async () => {
    setIsSaving(true);
    try {
      // Update assignments
      await onAssignUsers(task.id, selectedUserIds);
      onClose();
    } catch (error) {
      console.error('Error saving assignments:', error);
    } finally {
      setIsSaving(false);
    }
  };

  const handleUserToggle = (userId: string) => {
    setSelectedUserIds(prev => 
      prev.includes(userId) 
        ? prev.filter(id => id !== userId)
        : [...prev, userId]
    );
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <UserPlus className="h-5 w-5" />
            Assigned Members
          </DialogTitle>
        </DialogHeader>

        <ScrollArea className="max-h-[400px] -mx-6 px-6">
          <div className="space-y-2 py-2">
            {organizationMembers.length > 0 ? (
              organizationMembers.map(member => (
                <label
                  key={member.id}
                  className="flex items-center gap-3 p-3 hover:bg-muted rounded-lg cursor-pointer transition-colors"
                >
                  <input
                    type="checkbox"
                    checked={selectedUserIds.includes(member.id)}
                    onChange={() => handleUserToggle(member.id)}
                    className="w-4 h-4 rounded border-input"
                  />
                  <div className="flex items-center gap-3 flex-1">
                    {member.imageUrl ? (
                      <img
                        src={member.imageUrl}
                        alt={member.name}
                        className="w-10 h-10 rounded-full"
                      />
                    ) : (
                      <div className="w-10 h-10 bg-muted rounded-full flex items-center justify-center">
                        <User className="w-5 h-5" />
                      </div>
                    )}
                    <div className="flex-1 min-w-0">
                      <div className="font-medium truncate">
                        {member.name}
                      </div>
                      <div className="text-sm text-muted-foreground truncate">
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
              ))
            ) : (
              <div className="text-sm text-muted-foreground text-center py-8">
                No organization members available for assignment
              </div>
            )}
          </div>
        </ScrollArea>

        <DialogFooter>
          <Button
            type="button"
            onClick={handleSave}
            disabled={isSaving}
            className="w-full"
          >
            {isSaving ? 'Saving...' : 'Done'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default TaskDetailModal;