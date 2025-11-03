"use client";

import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { X as RemoveIcon } from "lucide-react";
import React from "react";

/**
 * used for identifying the split char and use will pasting
 */
// Split only on comma or newline (Enter)
const SPLITTER_REGEX = /[\n,]+/;

/**
 * used for formatting the pasted element for the correct value format to be added
 */

const FORMATTING_REGEX = /^[^a-zA-Z0-9]*|[^a-zA-Z0-9]*$/g;

interface TagsInputProps extends React.HTMLAttributes<HTMLDivElement> {
  value: string[];
  onValueChange: (value: string[]) => void;
  placeholder?: string;
  maxItems?: number;
  minItems?: number;
}

interface TagsInputContextProps {
  value: string[];
  onValueChange: (value: string[]) => void;
  inputValue: string;
  setInputValue: React.Dispatch<React.SetStateAction<string>>;
  activeIndex: number;
  setActiveIndex: React.Dispatch<React.SetStateAction<number>>;
}

const TagInputContext = React.createContext<TagsInputContextProps | null>(null);

export const TagsInput = React.forwardRef<HTMLDivElement, TagsInputProps>(
  (
    {
      children,
      value,
      onValueChange,
      placeholder,
      maxItems,
      minItems,
      className,
      dir,
      ...props
    },
    ref
  ) => {
    const [activeIndex, setActiveIndex] = React.useState(-1);
    const [inputValue, setInputValue] = React.useState("");
    const [disableInput, setDisableInput] = React.useState(false);
    const [disableButton, setDisableButton] = React.useState(false);
    const [isValueSelected, setIsValueSelected] = React.useState(false);
    const [selectedValue, setSelectedValue] = React.useState("");
    const [isTyping, setIsTyping] = React.useState(false);

    const parseMinItems = minItems ?? 0;
    const parseMaxItems = maxItems ?? Infinity;

    const onValueChangeHandler = React.useCallback(
      (val: string) => {
        if (
          !value.includes(val) &&
          value.length < parseMaxItems &&
          val.trim() !== ""
        ) {
          onValueChange([...value, val]);
        }
      },
      [value, onValueChange, parseMaxItems]
    );

    const RemoveValue = React.useCallback(
      (val: string) => {
        if (value.includes(val) && value.length > parseMinItems) {
          onValueChange(value.filter((item) => item !== val));
        }
      },
      [value, onValueChange, parseMinItems]
    );

    const handlePaste = React.useCallback(
      (e: React.ClipboardEvent<HTMLTextAreaElement>) => {
        e.preventDefault();
        const tags = e.clipboardData.getData("text").split(SPLITTER_REGEX);
        const newValue = [...value];
        tags.forEach((item) => {
          const parsedItem = item.replaceAll(FORMATTING_REGEX, "").trim();
          if (
            parsedItem.length > 0 &&
            !newValue.includes(parsedItem) &&
            newValue.length < parseMaxItems
          ) {
            newValue.push(parsedItem);
          }
        });
        onValueChange(newValue);
        setInputValue("");
      },
      [value, onValueChange, parseMaxItems]
    );

    const handleSelect = React.useCallback(
      (e: React.SyntheticEvent<HTMLTextAreaElement>) => {
        // Don't prevent default - let normal text selection work
        const target = e.currentTarget;
        const selection = target.value.substring(
          target.selectionStart ?? 0,
          target.selectionEnd ?? 0
        );

        setSelectedValue(selection);
        setIsValueSelected(selection === inputValue);
      },
      [inputValue]
    );

    React.useEffect(() => {
      const VerifyDisable = () => {
        if (value.length - 1 >= parseMinItems) {
          setDisableButton(false);
        } else {
          setDisableButton(true);
        }
        if (value.length + 1 <= parseMaxItems) {
          setDisableInput(false);
        } else {
          setDisableInput(true);
        }
      };
      VerifyDisable();
    }, [value, parseMinItems, parseMaxItems]);

    const handleKeyDown = React.useCallback(
      (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
        // Only handle specific keys that need special behavior
        if (e.key === "," || e.key === "Enter") {
          e.preventDefault();
          e.stopPropagation();
          if (inputValue.trim() !== "") {
            const newTag = inputValue.trim().replace(/,/g, "");
            onValueChangeHandler(newTag);
            setInputValue("");
          }
          return;
        }

        // Handle arrow keys for navigation
        if (e.key === "ArrowLeft" || e.key === "ArrowRight") {
          e.stopPropagation();
          const target = e.currentTarget;
          
          if (e.key === "ArrowLeft") {
            if (dir === "rtl") {
              if (value.length > 0 && activeIndex !== -1) {
                const nextIndex = activeIndex + 1 > value.length - 1 ? -1 : activeIndex + 1;
                setActiveIndex(nextIndex);
              }
            } else {
              if (value.length > 0 && target.selectionStart === 0) {
                const prevIndex = activeIndex - 1 < 0 ? value.length - 1 : activeIndex - 1;
                setActiveIndex(prevIndex);
              }
            }
          } else if (e.key === "ArrowRight") {
            if (dir === "rtl") {
              if (value.length > 0 && target.selectionStart === 0) {
                const prevIndex = activeIndex - 1 < 0 ? value.length - 1 : activeIndex - 1;
                setActiveIndex(prevIndex);
              }
            } else {
              if (value.length > 0 && activeIndex !== -1) {
                const nextIndex = activeIndex + 1 > value.length - 1 ? -1 : activeIndex + 1;
                setActiveIndex(nextIndex);
              }
            }
          }
          return;
        }

        // Handle backspace/delete for tag removal
        if (e.key === "Backspace" || e.key === "Delete") {
          e.stopPropagation();
          if (value.length > 0) {
            if (activeIndex !== -1 && activeIndex < value.length) {
              RemoveValue(value[activeIndex]);
              const newIndex = activeIndex - 1 <= 0 ? (value.length - 1 === 0 ? -1 : 0) : activeIndex - 1;
              setActiveIndex(newIndex);
            } else {
              const target = e.currentTarget;
              if (target.selectionStart === 0) {
                if (selectedValue === inputValue || isValueSelected) {
                  RemoveValue(value[value.length - 1]);
                }
              }
            }
          }
          return;
        }

        // Handle escape
        if (e.key === "Escape") {
          e.stopPropagation();
          const newIndex = activeIndex === -1 ? value.length - 1 : -1;
          setActiveIndex(newIndex);
          return;
        }

        // For all other keys, do nothing - let them pass through normally
      },
      [
        activeIndex,
        value,
        inputValue,
        RemoveValue,
        onValueChangeHandler,
        isValueSelected,
        selectedValue,
        dir,
      ]
    );

    const handleBlur = React.useCallback(() => {
      if (inputValue.trim() !== "") {
        onValueChangeHandler(inputValue.trim());
        setInputValue("");
      }
    }, [inputValue, onValueChangeHandler]);

    const mousePreventDefault = React.useCallback((e: React.MouseEvent) => {
      e.preventDefault();
      e.stopPropagation();
    }, []);

    const handleChange = React.useCallback(
      (e: React.ChangeEvent<HTMLTextAreaElement>) => {
        setInputValue(e.currentTarget.value);
      },
      []
    );

    return (
      <TagInputContext.Provider
        value={{
          value,
          onValueChange,
          inputValue,
          setInputValue,
          activeIndex,
          setActiveIndex,
        }}
      >
        <div
          {...props}
          ref={ref}
          dir={dir}
          className={cn(
            "flex items-start flex-wrap gap-1 p-1 rounded-lg bg-background overflow-hidden ring-1 ring-muted",
            {
              "focus-within:ring-ring": activeIndex === -1,
            },
            className
          )}
        >
          {value.map((item, index) => (
            <Badge
              tabIndex={activeIndex !== -1 ? 0 : activeIndex}
              key={item}
              aria-disabled={disableButton}
              data-active={activeIndex === index}
              className={cn(
                "relative px-1 rounded flex items-center gap-1 data-[active='true']:ring-2 data-[active='true']:ring-muted-foreground truncate aria-disabled:opacity-50 aria-disabled:cursor-not-allowed"
              )}
              variant={"secondary"}
            >
              <span className="text-xs">{item}</span>
              <button
                type="button"
                aria-label={`Remove ${item} option`}
                aria-roledescription="button to remove option"
                disabled={disableButton}
                onMouseDown={mousePreventDefault}
                onClick={() => RemoveValue(item)}
                className="disabled:cursor-not-allowed"
              >
                <span className="sr-only">Remove {item} option</span>
                <RemoveIcon className="h-4 w-4 hover:stroke-destructive" />
              </button>
            </Badge>
          ))}
          <textarea
            rows={3}
            tabIndex={0}
            aria-label="input tag"
            disabled={disableInput}
            onKeyDown={handleKeyDown}
            onBlur={handleBlur}
            onPaste={handlePaste}
            value={inputValue}
            onSelect={handleSelect}
            onChange={handleChange}
            placeholder={placeholder}
            onMouseDown={() => setActiveIndex(-1)}
            className={cn(
              "outline-0 border-none min-h-[84px] min-w-fit flex-1 focus-visible:outline-0 focus-visible:ring-0 focus-visible:ring-offset-0 focus-visible:border-0 placeholder:text-muted-foreground px-1",
              activeIndex !== -1 && "caret-transparent"
            )}
          />
        </div>
      </TagInputContext.Provider>
    );
  }
);

TagsInput.displayName = "TagsInput";

