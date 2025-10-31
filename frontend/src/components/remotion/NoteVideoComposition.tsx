import React from 'react';
import { AbsoluteFill, interpolate, spring, useCurrentFrame, useVideoConfig } from 'remotion';

export const NoteVideoComposition: React.FC<Record<string, any>> = ({
  title,
  content,
  durationInFrames,
  fps,
  theme,
  themeColors,
  isDark,
}) => {
  const frame = useCurrentFrame();
  const { width, height } = useVideoConfig();

  // Animation progress (0 to 1)
  const progress = frame / durationInFrames;

  // Title animation - bounce in with scale and rotation
  const titleSpring = spring({
    frame,
    fps,
    config: {
      damping: 15,
      stiffness: 200,
      mass: 1,
    },
  });

  const titleScale = interpolate(titleSpring, [0, 1], [0.5, 1]);
  const titleRotate = interpolate(titleSpring, [0, 1], [10, 0]);
  const titleOpacity = interpolate(frame, [0, 15], [0, 1], { extrapolateRight: 'clamp' });

  // Content animation - slide in from left with fade
  const contentStartFrame = Math.floor(durationInFrames * 0.25);
  const contentSpring = spring({
    frame: Math.max(0, frame - contentStartFrame),
    fps,
    config: {
      damping: 20,
      stiffness: 100,
    },
  });

  const contentX = interpolate(contentSpring, [0, 1], [-50, 0]);
  const contentOpacity = interpolate(contentSpring, [0, 1], [0, 1]);

  // Use dynamic theme colors or fallback to defaults
  const bgColor = isDark ? '#0a0a0a' : '#ffffff';
  const textColor = isDark ? '#f1f5f9' : '#1e293b';
  const accentColor = themeColors?.primary || '#3b82f6';
  const secondaryAccent = themeColors?.secondary || '#8b5cf6';

  // Animated gradient background
  const gradientRotation = interpolate(frame, [0, durationInFrames], [0, 360]);

  // Debug logging (will show in browser console when playing video)
  if (frame === 0) {
    console.log('Video composition rendering:', { title, content, contentLength: content?.length });
  }

  // Split content into lines for better display
  // First try splitting by newlines, if that doesn't work, split by sentences
  let contentLines = content.split('\n').filter((line: string) => line.trim());
  
  // If no newlines or very few lines, split into chunks by sentences or words
  if (contentLines.length <= 1 && content.length > 0) {
    // Split by sentences (., !, ?)
    const sentences = content.match(/[^.!?]+[.!?]+/g) || [];
    if (sentences.length > 1) {
      contentLines = sentences.map((s: string) => s.trim()).filter((s: string) => s);
    } else {
      // If still no good split, chunk by words (max ~60 chars per line)
      const words = content.split(' ');
      contentLines = [];
      let currentLine = '';
      for (const word of words) {
        if ((currentLine + ' ' + word).length > 60) {
          if (currentLine) contentLines.push(currentLine.trim());
          currentLine = word;
        } else {
          currentLine = currentLine ? currentLine + ' ' + word : word;
        }
      }
      if (currentLine) contentLines.push(currentLine.trim());
    }
  }
  
  // Limit to 8 lines for better visibility
  contentLines = contentLines.slice(0, 8).filter((line: string) => line.trim());
  
  if (frame === 0) {
    console.log('Content lines after processing:', contentLines);
    console.log('Number of lines:', contentLines.length);
  }

  return (
    <AbsoluteFill>
      {/* Animated gradient background with theme colors */}
      <div
        style={{
          position: 'absolute',
          width: '100%',
          height: '100%',
          background: `linear-gradient(${gradientRotation}deg, ${accentColor}${isDark ? '15' : '08'} 0%, ${secondaryAccent}${isDark ? '15' : '08'} 100%)`,
          backgroundColor: bgColor,
        }}
      />
      
      <div
        style={{
          position: 'relative',
          display: 'flex',
          flexDirection: 'column',
          justifyContent: 'flex-start',
          alignItems: 'center',
          width: '100%',
          height: '100%',
          padding: '8% 5%',
          boxSizing: 'border-box',
        }}
      >
        {/* Title with bounce and scale animation */}
        <div
          style={{
            transform: `scale(${titleScale}) rotate(${titleRotate}deg)`,
            opacity: titleOpacity,
            marginBottom: '2rem',
            textAlign: 'center',
          }}
        >
          <h1
            style={{
              fontSize: Math.min(width * 0.065, 60),
              fontWeight: '800',
              color: accentColor,
              margin: 0,
              fontFamily: 'system-ui, -apple-system, sans-serif',
              textShadow: theme === 'dark' 
                ? '3px 3px 6px rgba(0,0,0,0.7)' 
                : '2px 2px 4px rgba(0,0,0,0.1)',
              letterSpacing: '-0.02em',
            }}
          >
            {title}
          </h1>
          {/* Animated underline */}
          <div
            style={{
              width: `${titleSpring * 100}%`,
              height: '4px',
              background: `linear-gradient(90deg, ${accentColor}, ${secondaryAccent})`,
              marginTop: '1rem',
              borderRadius: '2px',
              boxShadow: `0 2px 8px ${accentColor}40`,
            }}
          />
        </div>

        {/* Content with slide-in animation */}
        <div
          style={{
            transform: `translateX(${contentX}px)`,
            opacity: contentOpacity,
            textAlign: 'left',
            maxWidth: '90%',
            flex: 1,
            overflow: 'hidden',
            display: 'flex',
            flexDirection: 'column',
            gap: '0.6rem',
            justifyContent: 'center',
            paddingBottom: '4rem',
          }}
        >
          {contentLines.length === 0 && (
            <p
              style={{
                fontSize: Math.min(width * 0.025, 24),
                color: textColor,
                fontFamily: 'system-ui, -apple-system, sans-serif',
                opacity: 0.7,
                fontStyle: 'italic',
              }}
            >
              No content available
            </p>
          )}
          {contentLines.map((line: string, index: number) => {
            // Stagger the animation for each line
            const lineStartFrame = contentStartFrame + (index * 8);
            const lineSpring = spring({
              frame: Math.max(0, frame - lineStartFrame),
              fps,
              config: {
                damping: 20,
                stiffness: 150,
              },
            });
            
            const lineScale = interpolate(lineSpring, [0, 1], [0.95, 1]);
            const lineX = interpolate(lineSpring, [0, 1], [20, 0]);

            return (
              <div
                key={index}
                style={{
                  transform: `translateX(${lineX}px) scale(${lineScale})`,
                  opacity: lineSpring,
                  padding: '0.7rem 1.5rem',
                  backgroundColor: theme === 'dark' 
                    ? 'rgba(59, 130, 246, 0.08)' 
                    : 'rgba(59, 130, 246, 0.06)',
                  borderRadius: '8px',
                  borderLeft: `4px solid ${index % 2 === 0 ? accentColor : secondaryAccent}`,
                  boxShadow: theme === 'dark'
                    ? '0 4px 12px rgba(0,0,0,0.4)'
                    : '0 4px 12px rgba(0,0,0,0.08)',
                  backdropFilter: 'blur(10px)',
                }}
              >
                <p
                  style={{
                    fontSize: Math.min(width * 0.022, 26),
                    color: textColor,
                    margin: 0,
                    fontFamily: 'system-ui, -apple-system, sans-serif',
                    lineHeight: 1.5,
                    fontWeight: '600',
                    textShadow: theme === 'dark' 
                      ? '0 2px 4px rgba(0,0,0,0.6)' 
                      : '0 1px 2px rgba(0,0,0,0.1)',
                  }}
                >
                  {line}
                </p>
              </div>
            );
          })}
        </div>

        {/* Animated progress bar at bottom */}
        <div
          style={{
            position: 'absolute',
            bottom: '5%',
            left: '10%',
            right: '10%',
            height: '6px',
            backgroundColor: theme === 'dark' ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.1)',
            borderRadius: '3px',
            overflow: 'hidden',
          }}
        >
          <div
            style={{
              width: `${progress * 100}%`,
              height: '100%',
              background: `linear-gradient(90deg, ${accentColor}, ${secondaryAccent})`,
              borderRadius: '3px',
              boxShadow: `0 0 10px ${accentColor}60`,
            }}
          />
        </div>

        {/* Floating particles for extra flair */}
        {[...Array(5)].map((_, i) => {
          const particleSpring = spring({
            frame: frame + (i * 10),
            fps,
            config: {
              damping: 100,
            },
          });
          const particleY = interpolate(
            particleSpring,
            [0, 1],
            [height * 0.8, height * 0.2]
          );
          const particleOpacity = interpolate(
            particleSpring,
            [0, 0.5, 1],
            [0, 0.6, 0]
          );
          
          return (
            <div
              key={i}
              style={{
                position: 'absolute',
                left: `${15 + i * 17}%`,
                top: particleY,
                width: '8px',
                height: '8px',
                borderRadius: '50%',
                backgroundColor: i % 2 === 0 ? accentColor : secondaryAccent,
                opacity: particleOpacity * 0.4,
                boxShadow: `0 0 10px ${i % 2 === 0 ? accentColor : secondaryAccent}`,
              }}
            />
          );
        })}
      </div>
    </AbsoluteFill>
  );
};
