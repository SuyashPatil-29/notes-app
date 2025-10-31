import React, { useState, useEffect } from 'react';
import { Player } from '@remotion/player';
import { NoteVideoComposition } from './remotion/NoteVideoComposition';
import { Loader2 } from 'lucide-react';
import { useTheme } from './theme-provider';
import { getThemeColors } from '@/utils/themeColors';

interface VideoData {
  title: string;
  content: string;
  durationInFrames: number;
  fps: number;
  theme: 'light' | 'dark';
}

interface NoteVideoPlayerProps {
  videoData: string;
  className?: string;
}

export const NoteVideoPlayer: React.FC<NoteVideoPlayerProps> = ({
  videoData,
  className = '',
}) => {
  const { theme, colorScheme } = useTheme();
  const [parsedVideoData, setParsedVideoData] = useState<VideoData | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Get theme colors dynamically
  const themeColors = getThemeColors(theme);
  const isDark = colorScheme === 'dark' || (colorScheme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches);

  useEffect(() => {
    try {
      const data = JSON.parse(videoData) as VideoData;
      console.log('Parsed video data:', data);
      console.log('Content:', data.content);
      console.log('Content length:', data.content?.length);
      setParsedVideoData(data);
      setIsLoading(false);
    } catch (err) {
      console.error('Failed to parse video data:', err);
      setError('Failed to load video data');
      setIsLoading(false);
    }
  }, [videoData]);

  if (isLoading) {
    return (
      <div className={`flex items-center justify-center p-8 border rounded-lg bg-muted/50 ${className}`}>
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        <span className="ml-2 text-sm text-muted-foreground">Loading video...</span>
      </div>
    );
  }

  if (error || !parsedVideoData) {
    return (
      <div className={`flex items-center justify-center p-8 border rounded-lg bg-destructive/10 ${className}`}>
        <span className="text-sm text-destructive">{error || 'Video data not available'}</span>
      </div>
    );
  }

  return (
    <div className={`space-y-4 ${className}`}>
      {/* Video Title */}
      <div className="text-center">
        <h3 className="text-lg font-semibold">{parsedVideoData.title}</h3>
        <p className="text-sm text-muted-foreground">
          {parsedVideoData.durationInFrames / parsedVideoData.fps}s video • {parsedVideoData.fps}fps
        </p>
      </div>

      {/* Video Player */}
      <div className="relative border rounded-lg overflow-hidden bg-black">
        <Player
          component={NoteVideoComposition as any}
          inputProps={{
            ...parsedVideoData,
            themeColors,
            isDark,
          } as any}
          durationInFrames={parsedVideoData.durationInFrames}
          compositionWidth={1920}
          compositionHeight={1080}
          fps={parsedVideoData.fps}
          style={{
            width: '100%',
            aspectRatio: '16/9',
          }}
          controls={true}
          autoPlay={false}
          loop={false}
          showVolumeControls={true}
          allowFullscreen={true}
        />
      </div>
    </div>
  );
};
