import React from 'react';
import { AbsoluteFill, interpolate, spring, useCurrentFrame, useVideoConfig } from 'remotion';

interface VideoSlide {
  type: string;
  title: string;
  content: string;
  items?: string[];
  duration: number;
}

interface VideoCompositionProps extends Record<string, unknown> {
  title: string;
  content?: string;
  slides?: VideoSlide[];
  durationInFrames: number;
  fps: number;
  theme: 'light' | 'dark';
  themeColors: any;
  isDark: boolean;
  backgroundStyle?: string;
  transitionStyle?: string;
}

export const NoteVideoComposition: React.FC<VideoCompositionProps> = ({
  title,
  content,
  slides,
  durationInFrames,
  fps,
  theme,
  themeColors,
  isDark,
  backgroundStyle = 'gradient',
  transitionStyle = 'slide',
}) => {
  const frame = useCurrentFrame();
  const { width, height } = useVideoConfig();
  
  // If slides exist, use multi-slide rendering, otherwise use legacy single-content rendering
  const hasSlides = slides && slides.length > 0;

  // Use dynamic theme colors or fallback to defaults
  const bgColor = isDark ? '#0a0a0a' : '#ffffff';
  const textColor = isDark ? '#f1f5f9' : '#1e293b';
  const accentColor = themeColors?.primary || '#3b82f6';
  const secondaryAccent = themeColors?.secondary || '#8b5cf6';

  // Animated gradient background
  const gradientRotation = interpolate(frame, [0, durationInFrames], [0, 360]);

  // Animation progress (0 to 1)
  const progress = frame / durationInFrames;

  // Get current slide information for multi-slide videos
  const getCurrentSlide = () => {
    if (!hasSlides) return null;
    
    let accumulatedFrames = 0;
    for (let i = 0; i < slides!.length; i++) {
      const slide = slides![i];
      const slideStartFrame = accumulatedFrames;
      const slideEndFrame = accumulatedFrames + slide.duration;
      
      if (frame >= slideStartFrame && frame < slideEndFrame) {
        return {
          slide,
          index: i,
          slideStartFrame,
          slideEndFrame,
          slideLocalFrame: frame - slideStartFrame,
          isFirstSlide: i === 0,
          isLastSlide: i === slides!.length - 1,
        };
      }
      
      accumulatedFrames += slide.duration;
    }
    
    // If we're past all slides, return the last slide
    const lastSlide = slides![slides!.length - 1];
    let accumulatedTotal = 0;
    for (let i = 0; i < slides!.length; i++) {
      accumulatedTotal += slides![i].duration;
    }
    return {
      slide: lastSlide,
      index: slides!.length - 1,
      slideStartFrame: accumulatedTotal - lastSlide.duration,
      slideEndFrame: accumulatedTotal,
      slideLocalFrame: lastSlide.duration - 1,
      isFirstSlide: false,
      isLastSlide: true,
    };
  };

  const currentSlideInfo = hasSlides ? getCurrentSlide() : null;

  // Render slide content based on type
  const renderSlideContent = (slide: VideoSlide, localFrame: number) => {
    // Slide animations
    const slideSpring = spring({
      frame: localFrame,
      fps,
      config: {
        damping: 20,
        stiffness: 150,
      },
    });

    const slideOpacity = interpolate(localFrame, [0, 15], [0, 1], { extrapolateRight: 'clamp' });
    const slideScale = interpolate(slideSpring, [0, 1], [0.95, 1]);
    
    // Transition animation
    let slideX = 0;
    if (transitionStyle === 'slide') {
      slideX = interpolate(slideSpring, [0, 1], [50, 0]);
    }

    const containerStyle = {
      transform: `translateX(${slideX}px) scale(${slideScale})`,
      opacity: slideOpacity,
      width: '90%',
      maxWidth: '1600px',
      margin: '0 auto',
      display: 'flex',
      flexDirection: 'column' as const,
      justifyContent: 'center' as const,
      minHeight: '80%',
      padding: '5% 0',
    };

    switch (slide.type) {
      case 'title':
        return (
          <div style={containerStyle}>
            <div style={{ textAlign: 'center' }}>
              <h1
                style={{
                  fontSize: Math.min(width * 0.07, 80),
                  fontWeight: '900',
                  color: accentColor,
                  margin: 0,
                  marginBottom: '2rem',
                  fontFamily: 'system-ui, -apple-system, sans-serif',
                  textShadow: isDark 
                    ? '4px 4px 8px rgba(0,0,0,0.8)' 
                    : '3px 3px 6px rgba(0,0,0,0.15)',
                  letterSpacing: '-0.03em',
                }}
              >
                {slide.title}
              </h1>
              {slide.content && (
                <p
                  style={{
                    fontSize: Math.min(width * 0.03, 36),
                    color: textColor,
                    opacity: 0.85,
                    fontFamily: 'system-ui, -apple-system, sans-serif',
                    fontWeight: '500',
                    lineHeight: 1.6,
                  }}
                >
                  {slide.content}
                </p>
              )}
            </div>
          </div>
        );

      case 'list':
        return (
          <div style={containerStyle}>
            <h2
              style={{
                fontSize: Math.min(width * 0.05, 60),
                fontWeight: '800',
                color: accentColor,
                marginBottom: '3rem',
                fontFamily: 'system-ui, -apple-system, sans-serif',
                textShadow: isDark 
                  ? '3px 3px 6px rgba(0,0,0,0.7)' 
                  : '2px 2px 4px rgba(0,0,0,0.1)',
              }}
            >
              {slide.title}
            </h2>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
              {slide.items?.map((item, index) => {
                const itemStartFrame = 15 + (index * 10);
                const itemSpring = spring({
                  frame: Math.max(0, localFrame - itemStartFrame),
                  fps,
                  config: {
                    damping: 20,
                    stiffness: 150,
                  },
                });
                
                const itemX = interpolate(itemSpring, [0, 1], [30, 0]);
                const itemOpacity = interpolate(itemSpring, [0, 1], [0, 1]);

                return (
                  <div
                    key={index}
                    style={{
                      transform: `translateX(${itemX}px)`,
                      opacity: itemOpacity,
                      display: 'flex',
                      alignItems: 'center',
                      gap: '1.5rem',
                      padding: '1.5rem 2rem',
                      backgroundColor: isDark 
                        ? 'rgba(59, 130, 246, 0.08)' 
                        : 'rgba(59, 130, 246, 0.06)',
                      borderRadius: '12px',
                      borderLeft: `5px solid ${index % 2 === 0 ? accentColor : secondaryAccent}`,
                      boxShadow: isDark
                        ? '0 4px 12px rgba(0,0,0,0.4)'
                        : '0 4px 12px rgba(0,0,0,0.08)',
                    }}
                  >
                    <div
                      style={{
                        width: '40px',
                        height: '40px',
                        borderRadius: '50%',
                        backgroundColor: index % 2 === 0 ? accentColor : secondaryAccent,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        fontSize: '20px',
                        fontWeight: '700',
                        color: 'white',
                        flexShrink: 0,
                      }}
                    >
                      {index + 1}
                    </div>
                    <p
                      style={{
                        fontSize: Math.min(width * 0.025, 28),
                        color: textColor,
                        margin: 0,
                        fontFamily: 'system-ui, -apple-system, sans-serif',
                        fontWeight: '600',
                        lineHeight: 1.5,
                      }}
                    >
                      {item}
                    </p>
                  </div>
                );
              })}
            </div>
          </div>
        );

      case 'quote':
        return (
          <div style={containerStyle}>
            <div
              style={{
                padding: '4rem',
                backgroundColor: isDark 
                  ? 'rgba(139, 92, 246, 0.1)' 
                  : 'rgba(139, 92, 246, 0.08)',
                borderRadius: '20px',
                borderLeft: `10px solid ${secondaryAccent}`,
                boxShadow: isDark
                  ? '0 8px 24px rgba(0,0,0,0.5)'
                  : '0 8px 24px rgba(0,0,0,0.1)',
              }}
            >
              <div
                style={{
                  fontSize: Math.min(width * 0.06, 60),
                  color: secondaryAccent,
                  marginBottom: '1rem',
                  fontFamily: 'Georgia, serif',
                  opacity: 0.5,
                }}
              >
                "
              </div>
              <p
                style={{
                  fontSize: Math.min(width * 0.03, 36),
                  color: textColor,
                  fontFamily: 'Georgia, serif',
                  fontStyle: 'italic',
                  lineHeight: 1.8,
                  margin: 0,
                }}
              >
                {slide.content}
              </p>
              {slide.title && (
                <p
                  style={{
                    fontSize: Math.min(width * 0.025, 28),
                    color: accentColor,
                    fontFamily: 'system-ui, -apple-system, sans-serif',
                    fontWeight: '600',
                    marginTop: '2rem',
                    marginBottom: 0,
                  }}
                >
                  — {slide.title}
                </p>
              )}
            </div>
          </div>
        );

      case 'content':
      default:
        return (
          <div style={containerStyle}>
            <h2
              style={{
                fontSize: Math.min(width * 0.05, 60),
                fontWeight: '800',
                color: accentColor,
                marginBottom: '2rem',
                fontFamily: 'system-ui, -apple-system, sans-serif',
                textShadow: isDark 
                  ? '3px 3px 6px rgba(0,0,0,0.7)' 
                  : '2px 2px 4px rgba(0,0,0,0.1)',
              }}
            >
              {slide.title}
            </h2>
            <div
              style={{
                padding: '2rem 2.5rem',
                backgroundColor: isDark 
                  ? 'rgba(59, 130, 246, 0.08)' 
                  : 'rgba(59, 130, 246, 0.06)',
                borderRadius: '12px',
                borderLeft: `5px solid ${accentColor}`,
                boxShadow: isDark
                  ? '0 6px 16px rgba(0,0,0,0.4)'
                  : '0 6px 16px rgba(0,0,0,0.08)',
              }}
            >
              <p
                style={{
                  fontSize: Math.min(width * 0.028, 32),
                  color: textColor,
                  margin: 0,
                  fontFamily: 'system-ui, -apple-system, sans-serif',
                  lineHeight: 1.7,
                  fontWeight: '500',
                }}
              >
                {slide.content}
              </p>
            </div>
          </div>
        );
    }
  };

  // Render legacy single-content format
  const renderLegacyContent = () => {
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

    // Legacy content parsing
    if (!content) return null;
    
    // Split content into lines for better display
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

    return (
      <>
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
      </>
    );
  };

  // Main render
  return (
    <AbsoluteFill>
      {/* Animated gradient background */}
      <div
        style={{
          position: 'absolute',
          width: '100%',
          height: '100%',
          background: backgroundStyle === 'gradient' 
            ? `linear-gradient(${gradientRotation}deg, ${accentColor}${isDark ? '15' : '08'} 0%, ${secondaryAccent}${isDark ? '15' : '08'} 100%)`
            : 'transparent',
          backgroundColor: bgColor,
        }}
      />
      
      <div
        style={{
          position: 'relative',
          display: 'flex',
          flexDirection: 'column',
          justifyContent: hasSlides ? 'center' : 'flex-start',
          alignItems: 'center',
          width: '100%',
          height: '100%',
          padding: hasSlides ? '5%' : '8% 5%',
          boxSizing: 'border-box',
        }}
      >
        {/* Render AI-generated slides or legacy content */}
        {hasSlides && currentSlideInfo ? (
          renderSlideContent(currentSlideInfo.slide, currentSlideInfo.slideLocalFrame)
        ) : (
          renderLegacyContent()
        )}
      </div>

      {/* Animated progress bar at bottom */}
      <div
        style={{
          position: 'absolute',
          bottom: '5%',
          left: '10%',
          right: '10%',
          height: '6px',
          backgroundColor: isDark ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.1)',
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
    </AbsoluteFill>
  );
};
