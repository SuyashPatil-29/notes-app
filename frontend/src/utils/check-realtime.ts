/**
 * Utility to check if real-time features are properly configured
 */

export const isRealtimeConfigured = (): boolean => {
  const supabaseUrl = import.meta.env.VITE_SUPABASE_URL;
  const supabaseKey = import.meta.env.VITE_SUPABASE_PUBLISHABLE_OR_ANON_KEY;
  
  return !!(supabaseUrl && supabaseKey && 
    supabaseUrl !== 'undefined' && 
    supabaseKey !== 'undefined');
};

export const getRealtimeStatus = () => {
  const supabaseUrl = import.meta.env.VITE_SUPABASE_URL;
  const supabaseKey = import.meta.env.VITE_SUPABASE_PUBLISHABLE_OR_ANON_KEY;
  
  return {
    configured: isRealtimeConfigured(),
    hasUrl: !!(supabaseUrl && supabaseUrl !== 'undefined'),
    hasKey: !!(supabaseKey && supabaseKey !== 'undefined'),
    message: isRealtimeConfigured() 
      ? 'Real-time features enabled' 
      : 'Real-time features disabled: Configure VITE_SUPABASE_URL and VITE_SUPABASE_PUBLISHABLE_OR_ANON_KEY in .env'
  };
};

// Log realtime status in development
if (import.meta.env.DEV) {
  const status = getRealtimeStatus();
  if (status.configured) {
    console.log('✅ Real-time collaboration enabled');
  } else {
    console.warn('⚠️ Real-time collaboration disabled');
    console.warn('📝 To enable:');
    console.warn('   1. Create a Supabase project at https://supabase.com');
    console.warn('   2. Add VITE_SUPABASE_URL and VITE_SUPABASE_PUBLISHABLE_OR_ANON_KEY to .env');
    console.warn('   3. Restart dev server');
  }
}

