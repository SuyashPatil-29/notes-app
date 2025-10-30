import React from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import remarkMath from 'remark-math';
import rehypeKatex from 'rehype-katex';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { oneDark } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { getMarkdownContent } from '@/utils/markdown';
import 'katex/dist/katex.min.css';

interface MarkdownRendererProps {
  content: string;
  className?: string;
}

export const MarkdownRenderer: React.FC<MarkdownRendererProps> = ({
  content,
  className = ''
}) => {
  const markdownContent = getMarkdownContent(content);

  return (
    <div className={`prose prose-slate dark:prose-invert max-w-none ${className}`}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm, remarkMath]}
        rehypePlugins={[rehypeKatex]}
        components={{
          code({ className, children, ...props }) {
            const match = /language-(\w+)/.exec(className || '');
            const language = match ? match[1] : '';
            const isInline = !className;

            return !isInline ? (
              <SyntaxHighlighter
                style={oneDark as any}
                language={language}
                PreTag="div"
                className="rounded-md"
              >
                {String(children).replace(/\n$/, '')}
              </SyntaxHighlighter>
            ) : (
              <code className="bg-slate-100 dark:bg-slate-800 px-1 py-0.5 rounded text-sm" {...props}>
                {children}
              </code>
            );
          },
          // Custom blockquote styling
          blockquote({ children }) {
            return (
              <blockquote className="border-l-4 border-slate-300 dark:border-slate-600 pl-4 italic text-slate-600 dark:text-slate-400">
                {children}
              </blockquote>
            );
          },
          // Custom heading styling
          h1({ children }) {
            return <h1 className="text-3xl font-bold mb-4 mt-8 first:mt-0">{children}</h1>;
          },
          h2({ children }) {
            return <h2 className="text-2xl font-bold mb-3 mt-6">{children}</h2>;
          },
          h3({ children }) {
            return <h3 className="text-xl font-semibold mb-2 mt-5">{children}</h3>;
          },
          h4({ children }) {
            return <h4 className="text-lg font-semibold mb-2 mt-4">{children}</h4>;
          },
          h5({ children }) {
            return <h5 className="text-base font-semibold mb-1 mt-3">{children}</h5>;
          },
          h6({ children }) {
            return <h6 className="text-sm font-semibold mb-1 mt-3">{children}</h6>;
          },
          // Custom paragraph styling
          p({ children }) {
            return <p className="mb-4 leading-relaxed">{children}</p>;
          },
          // Custom list styling
          ul({ children }) {
            return <ul className="list-disc list-inside mb-4 space-y-1">{children}</ul>;
          },
          ol({ children }) {
            return <ol className="list-decimal list-inside mb-4 space-y-1">{children}</ol>;
          },
          li({ children }) {
            return <li className="leading-relaxed">{children}</li>;
          },
          // Custom link styling
          a({ children, href }) {
            return (
              <a
                href={href}
                className="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 underline"
                target="_blank"
                rel="noopener noreferrer"
              >
                {children}
              </a>
            );
          },
          // Custom table styling
          table({ children }) {
            return (
              <div className="overflow-x-auto mb-4">
                <table className="min-w-full border-collapse border border-slate-300 dark:border-slate-600">
                  {children}
                </table>
              </div>
            );
          },
          thead({ children }) {
            return <thead className="bg-slate-100 dark:bg-slate-800">{children}</thead>;
          },
          tbody({ children }) {
            return <tbody>{children}</tbody>;
          },
          tr({ children }) {
            return <tr className="border-b border-slate-200 dark:border-slate-700">{children}</tr>;
          },
          th({ children }) {
            return (
              <th className="border border-slate-300 dark:border-slate-600 px-4 py-2 text-left font-semibold">
                {children}
              </th>
            );
          },
          td({ children }) {
            return (
              <td className="border border-slate-300 dark:border-slate-600 px-4 py-2">
                {children}
              </td>
            );
          },
          // Custom horizontal rule
          hr() {
            return <hr className="border-slate-300 dark:border-slate-600 my-8" />;
          },
        }}
      >
        {markdownContent}
      </ReactMarkdown>
    </div>
  );
};
