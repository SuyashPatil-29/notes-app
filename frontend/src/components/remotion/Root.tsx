import React from 'react';
import {Composition} from 'remotion';
import {MyComposition} from './Composition';
import {NoteVideoComposition} from './NoteVideoComposition';

export const RemotionRoot: React.FC = () => {
  return (
    <>
      <Composition
        id="Empty"
        component={MyComposition}
        durationInFrames={60}
        fps={30}
        width={1280}
        height={720}
      />
      <Composition
        id="NoteVideo"
        component={NoteVideoComposition}
        durationInFrames={180}
        fps={30}
        width={1920}
        height={1080}
        defaultProps={{
          title: "Sample Note Title",
          content: "This is sample content for the video composition.",
          durationInFrames: 180,
          fps: 30,
          theme: "light" as const,
          themeColors: {
            primary: '#3b82f6',
            secondary: '#8b5cf6',
          },
          isDark: false,
          backgroundStyle: 'gradient',
          transitionStyle: 'slide',
        }}
      />
    </>
  );
};