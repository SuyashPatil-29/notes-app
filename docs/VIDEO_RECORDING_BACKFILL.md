# Video Recording Backfill Feature

## Overview

The video recording backfill feature allows users to fetch video recordings for existing meetings that were recorded before the video feature was implemented. This is useful for meetings that have a `recall_recording_id` in the database but are missing the `video_download_url`.

## How It Works

### Backend Implementation

1. **New Endpoint**: `POST /meetings/backfill-videos`
   - Authenticates the user
   - Finds all completed meetings with `recall_recording_id` but missing `video_download_url`
   - For each meeting:
     - Fetches bot details from Recall.ai API
     - Extracts the video download URL from the bot's recordings
     - Updates the database with the video URL
   - Returns a summary of the operation including:
     - Total meetings processed
     - Successfully updated count
     - Failed count
     - Error details (if any)

2. **Controller Function**: `BackfillVideoURLs` in `meeting.controller.go`
   - Located at: `/Users/suyash/Desktop/work/notes-app/backend/internal/controllers/meeting.controller.go`
   - Handles the backfill logic
   - Provides detailed logging for debugging

3. **API Route**: Registered in `cmd/main.go`
   ```go
   protected.POST("/meetings/backfill-videos", controllers.BackfillVideoURLs)
   ```

### Frontend Implementation

1. **Utility Function**: `backfillVideoURLs` in `meeting.ts`
   - Located at: `/Users/suyash/Desktop/work/notes-app/frontend/src/utils/meeting.ts`
   - Makes a POST request to the backfill endpoint
   - Returns a promise with the operation results

2. **UI Component**: Banner in `MeetingsList.tsx`
   - Automatically detects meetings missing video URLs
   - Shows a banner with the count of meetings needing backfill
   - Provides a "Fetch Videos" button to trigger the backfill
   - Displays loading state during the operation
   - Shows success/error toasts with operation results
   - Automatically refreshes the meetings list after successful backfill

## User Experience

1. **Detection**: When a user visits the meetings page, the app automatically checks for meetings that have recordings but are missing video URLs.

2. **Notification**: If any meetings need video backfill, a blue banner appears at the top of the meetings list:
   ```
   [Download Icon] X meeting(s) missing video recordings
                   Fetch video recordings for your previous meetings
                   [Fetch Videos Button]
   ```

3. **Action**: User clicks "Fetch Videos" button
   - Button shows loading state: "Fetching..."
   - Backend processes each meeting
   - Success toast appears with the number of updated meetings
   - Banner disappears if all videos were fetched successfully

4. **Results**:
   - **All up to date**: Shows message that no updates were needed
   - **Success**: Shows count of successfully updated meetings
   - **Partial failure**: Shows both success count and error count
   - **Complete failure**: Shows error message

## API Response Format

```typescript
{
  message: string;           // Human-readable message
  meetings_count: number;    // Total meetings processed
  updated_count: number;     // Successfully updated meetings
  failed_count: number;      // Failed meetings
  errors?: string[];         // Array of error messages (if any failures)
}
```

## Database Schema

The `video_download_url` field was added to the `meeting_recordings` table:

```sql
ALTER TABLE meeting_recordings
ADD COLUMN IF NOT EXISTS video_download_url TEXT;
```

**Migration Files**:
- Forward: `/Users/suyash/Desktop/work/notes-app/backend/db/migrations/005_add_video_download_url.sql`
- Rollback: `/Users/suyash/Desktop/work/notes-app/backend/db/migrations/005_add_video_download_url_rollback.sql`

## Technical Details

### Recall.ai API Integration

The backfill uses the existing Recall.ai client to fetch bot details:

```go
botDetails, err := recallClient.GetBot(meeting.BotID)
videoURL := botDetails.Recordings[0].MediaShortcuts.VideoMixed.Data.DownloadURL
```

### Error Handling

The system gracefully handles various error scenarios:
- Bot not found
- No recordings available
- Video URL not available
- Database save failures

Each error is logged with relevant context and included in the response.

### Performance Considerations

- Backfill processes meetings sequentially to avoid rate limiting
- Each meeting requires one API call to Recall.ai
- Database updates use GORM's `Save` method for safety
- The operation runs synchronously, so users see immediate results

## Testing

To test the backfill feature:

1. **Create a test meeting** with a valid `recall_recording_id` but no `video_download_url`:
   ```sql
   UPDATE meeting_recordings 
   SET video_download_url = NULL 
   WHERE recall_recording_id IS NOT NULL 
   LIMIT 1;
   ```

2. **Visit the meetings page** in the frontend - you should see the backfill banner

3. **Click "Fetch Videos"** and observe:
   - Loading state
   - Success toast
   - Updated meeting with video player
   - Banner disappears

## Future Enhancements

Potential improvements for the backfill feature:

1. **Batch Processing**: Process multiple meetings in parallel
2. **Scheduled Backfill**: Automatic background job to check for missing videos
3. **Retry Logic**: Automatic retry for failed meetings
4. **Progress Indicator**: Show progress for large backfill operations
5. **Selective Backfill**: Allow users to choose specific meetings to backfill

## Troubleshooting

### "No meetings need video URL updates"
- All meetings already have video URLs
- Or no meetings have `recall_recording_id`

### "Video URL not available for this recording"
- Recall.ai may not have generated the video yet
- Video recording might not have been enabled for that bot
- Try again later or check Recall.ai dashboard

### "Failed to get bot details from Recall.ai"
- Check Recall.ai API credentials
- Verify network connectivity
- Check Recall.ai API status
- Bot may have been deleted

## Related Files

**Backend**:
- `/Users/suyash/Desktop/work/notes-app/backend/internal/controllers/meeting.controller.go`
- `/Users/suyash/Desktop/work/notes-app/backend/cmd/main.go`
- `/Users/suyash/Desktop/work/notes-app/backend/pkg/recallai/client.go`

**Frontend**:
- `/Users/suyash/Desktop/work/notes-app/frontend/src/components/MeetingsList.tsx`
- `/Users/suyash/Desktop/work/notes-app/frontend/src/utils/meeting.ts`
- `/Users/suyash/Desktop/work/notes-app/frontend/src/components/MeetingDetail.tsx`

**Database**:
- `/Users/suyash/Desktop/work/notes-app/backend/db/migrations/005_add_video_download_url.sql`
- `/Users/suyash/Desktop/work/notes-app/backend/internal/models/meeting.model.go`

