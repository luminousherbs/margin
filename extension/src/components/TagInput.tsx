import { useState, useRef, useEffect } from 'react';
import { X, Tag } from 'lucide-react';

interface TagInputProps {
  tags: string[];
  onChange: (tags: string[]) => void;
  suggestions?: string[];
  placeholder?: string;
}

export default function TagInput({
  tags,
  onChange,
  suggestions = [],
  placeholder = 'Add tag...',
}: TagInputProps) {
  const [input, setInput] = useState('');
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  const filtered = input.trim()
    ? suggestions.filter((s) => s.toLowerCase().includes(input.toLowerCase()) && !tags.includes(s))
    : [];

  useEffect(() => {
    setSelectedIndex(0);
  }, [input]);

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setShowSuggestions(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  function addTag(tag: string) {
    const normalized = tag
      .trim()
      .toLowerCase()
      .replace(/[^a-z0-9_-]/g, '');
    if (normalized && !tags.includes(normalized) && tags.length < 10) {
      onChange([...tags, normalized]);
    }
    setInput('');
    setShowSuggestions(false);
    inputRef.current?.focus();
  }

  function removeTag(tag: string) {
    onChange(tags.filter((t) => t !== tag));
    inputRef.current?.focus();
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault();
      if (filtered.length > 0 && showSuggestions) {
        addTag(filtered[selectedIndex] || filtered[0]);
      } else if (input.trim()) {
        addTag(input);
      }
    } else if (e.key === 'Backspace' && !input && tags.length > 0) {
      removeTag(tags[tags.length - 1]);
    } else if (e.key === 'ArrowDown' && showSuggestions) {
      e.preventDefault();
      setSelectedIndex((i) => Math.min(i + 1, filtered.length - 1));
    } else if (e.key === 'ArrowUp' && showSuggestions) {
      e.preventDefault();
      setSelectedIndex((i) => Math.max(i - 1, 0));
    } else if (e.key === 'Escape') {
      setShowSuggestions(false);
    }
  }

  return (
    <div ref={containerRef} className="relative">
      <div
        className="flex flex-wrap items-center gap-1.5 p-2 bg-[var(--bg-card)] border border-[var(--border)] rounded-lg text-xs cursor-text min-h-[34px] focus-within:border-[var(--accent)] ring-0 outline-none transition-all"
        onClick={() => inputRef.current?.focus()}
      >
        <Tag size={12} className="text-[var(--text-tertiary)] flex-shrink-0" />
        {tags.map((tag) => (
          <span
            key={tag}
            className="inline-flex items-center gap-1 px-2 py-0.5 bg-[var(--accent-subtle)] text-[var(--accent)] rounded-md font-medium text-[11px]"
          >
            {tag}
            <button
              type="button"
              onClick={(e) => {
                e.stopPropagation();
                removeTag(tag);
              }}
              className="hover:text-[var(--text-primary)] transition-colors"
            >
              <X size={10} />
            </button>
          </span>
        ))}
        <input
          ref={inputRef}
          type="text"
          value={input}
          onChange={(e) => {
            setInput(e.target.value);
            setShowSuggestions(true);
          }}
          onFocus={() => setShowSuggestions(true)}
          onKeyDown={handleKeyDown}
          placeholder={tags.length === 0 ? placeholder : ''}
          className="flex-1 min-w-[60px] bg-transparent border-none outline-none ring-0 text-[11px] text-[var(--text-primary)] placeholder:text-[var(--text-tertiary)] focus:outline-none focus:ring-0 focus:border-none focus:shadow-none"
          style={{ outline: 'none', boxShadow: 'none' }}
        />
      </div>

      {showSuggestions && filtered.length > 0 && (
        <div className="absolute z-50 mt-1 w-full bg-[var(--bg-card)] border border-[var(--border)] rounded-lg shadow-lg overflow-hidden max-h-[140px] overflow-y-auto">
          {filtered.slice(0, 8).map((suggestion, i) => (
            <button
              key={suggestion}
              type="button"
              onMouseDown={(e) => {
                e.preventDefault();
                addTag(suggestion);
              }}
              className={`w-full text-left px-3 py-1.5 text-[11px] transition-colors ${
                i === selectedIndex
                  ? 'bg-[var(--accent-subtle)] text-[var(--accent)]'
                  : 'text-[var(--text-secondary)] hover:bg-[var(--bg-hover)]'
              }`}
            >
              {suggestion}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
