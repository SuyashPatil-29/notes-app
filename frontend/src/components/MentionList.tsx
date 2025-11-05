import {
  forwardRef,
  useEffect,
  useImperativeHandle,
  useRef,
  useState,
} from 'react';
import { Loader2 } from 'lucide-react';

interface MentionListProps {
  items: Array<{ id: string; label: string }>;
  command: (item: { id: string; label: string }) => void;
  loading?: boolean;
}

export interface MentionListRef {
  onKeyDown: (props: { event: KeyboardEvent }) => boolean;
}

const MentionList = forwardRef<MentionListRef, MentionListProps>(
  (props, ref) => {
    const [selectedIndex, setSelectedIndex] = useState(0);
    const selectedRef = useRef<HTMLButtonElement>(null);

    const selectItem = (index: number) => {
      const item = props.items[index];
      if (item) {
        props.command(item);
      }
    };

    const upHandler = () => {
      setSelectedIndex(
        (selectedIndex + props.items.length - 1) % props.items.length
      );
    };

    const downHandler = () => {
      setSelectedIndex((selectedIndex + 1) % props.items.length);
    };

    const enterHandler = () => {
      selectItem(selectedIndex);
    };

    useEffect(() => setSelectedIndex(0), [props.items]);

    // Scroll selected item into view
    useEffect(() => {
      if (selectedRef.current) {
        selectedRef.current.scrollIntoView({
          block: 'nearest',
          behavior: 'smooth',
        });
      }
    }, [selectedIndex]);

    useImperativeHandle(ref, () => ({
      onKeyDown: ({ event }: { event: KeyboardEvent }) => {
        if (event.key === 'ArrowUp') {
          upHandler();
          return true;
        }

        if (event.key === 'ArrowDown') {
          downHandler();
          return true;
        }

        if (event.key === 'Enter') {
          enterHandler();
          return true;
        }

        return false;
      },
    }));

    if (props.loading) {
      return (
        <div className="bg-popover text-popover-foreground rounded-lg shadow-lg border p-3 text-sm flex items-center gap-2">
          <Loader2 className="h-4 w-4 animate-spin" />
          <span>Loading notes...</span>
        </div>
      );
    }

    if (props.items.length === 0) {
      return (
        <div className="bg-popover text-popover-foreground rounded-lg shadow-lg border p-2 text-sm">
          No notes found
        </div>
      );
    }

    return (
      <div className="bg-popover text-popover-foreground rounded-lg shadow-lg border p-1 max-h-60 overflow-y-auto">
        {props.items.map((item, index) => (
          <button
            ref={index === selectedIndex ? selectedRef : null}
            className={`block w-full text-left px-3 py-2 rounded text-sm ${
              index === selectedIndex
                ? 'bg-accent text-accent-foreground'
                : 'hover:bg-accent/50'
            }`}
            key={item.id}
            onClick={() => selectItem(index)}
          >
            {item.label}
          </button>
        ))}
      </div>
    );
  }
);

MentionList.displayName = 'MentionList';

export default MentionList;

