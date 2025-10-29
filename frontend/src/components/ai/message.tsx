import { cn } from '@/lib/utils';
import type { UIMessage } from 'ai';
import type { ComponentProps, HTMLAttributes } from 'react';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Bot, User } from 'lucide-react';

export type MessageProps = HTMLAttributes<HTMLDivElement> & {
  from: UIMessage['role'];
};

export const Message = ({ className, from, ...props }: MessageProps) => (
  <div
    className={cn(
      'group flex items-start gap-3 py-2.5',
      from === 'user' ? 'is-user flex-row-reverse justify-start' : 'is-assistant justify-start',
      className
    )}
    {...props}
  />
);

export type MessageContentProps = HTMLAttributes<HTMLDivElement>;

export const MessageContent = ({
  children,
  className,
  ...props
}: MessageContentProps) => (
  <div
    className={cn(
      'flex flex-col gap-2 overflow-hidden rounded-2xl px-4 py-2.5 text-foreground text-sm max-w-[85%]',
      'group-[.is-user]:bg-primary group-[.is-user]:text-primary-foreground',
      'group-[.is-assistant]:bg-muted/50 group-[.is-assistant]:text-foreground',
      className
    )}
    {...props}
  >
    <div>{children}</div>
  </div>
);

export type MessageAvatarProps = ComponentProps<typeof Avatar> & {
  role: UIMessage['role'];
  image?: string | null;
};

export const MessageAvatar = ({
  role,
  className,
  image,
  ...props
}: MessageAvatarProps) => (
  <Avatar
    className={cn('size-9 shrink-0 mt-1', className)}
    {...props}
  >
    {role === 'user' && image && (
      <AvatarImage src={image} alt="User avatar" />
    )}
    <AvatarFallback className={cn(
      'font-medium',
      role === 'user'
        ? 'bg-primary text-primary-foreground'
        : 'bg-muted text-muted-foreground'
    )}>
      {role === 'user' ? (
        <User className="size-4" />
      ) : (
        <Bot className="size-4" />
      )}
    </AvatarFallback>
  </Avatar>
);
