# Deployment Guide

## Environment Variables

When deploying to production platforms like Leapcell, Vercel, or others, make sure to configure the following environment variables:

### Required Environment Variables

#### CORS Configuration
```bash
CORS_ORIGINS=https://atlasnotes-eta.vercel.app,https://yourdomain.com
```
- **Multiple origins**: Separate with commas
- **No trailing slash**: Use `https://domain.com` not `https://domain.com/`
- **Include protocol**: Always use `https://` for production
- **Local development**: `http://localhost:5173` is automatically included

#### Database Configuration
```bash
# PostgreSQL connection
DB_HOST=your-db-host
DB_PORT=5432
DB_USER=your-db-user
DB_PASSWORD=your-db-password
DB_NAME=your-db-name
```

#### Authentication
```bash
# Clerk Authentication
CLERK_SECRET_KEY=your-clerk-secret-key
```

#### AI Services (Optional)
```bash
# OpenAI
OPENAI_API_KEY=your-openai-key

# Anthropic Claude
ANTHROPIC_API_KEY=your-anthropic-key

# Google AI
GOOGLE_AI_API_KEY=your-google-ai-key
```

#### Calendar Integration (Optional)
```bash
# Google Calendar
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_CALLBACK_URL=https://your-backend-url/api/calendar/google/callback

# Microsoft Calendar
MICROSOFT_CLIENT_ID=your-microsoft-client-id
MICROSOFT_CLIENT_SECRET=your-microsoft-client-secret
MICROSOFT_CALLBACK_URL=https://your-backend-url/api/calendar/microsoft/callback
```

#### WhatsApp Integration (Optional)
```bash
WHATSAPP_BUSINESS_ACCOUNT_ID=your-account-id
WHATSAPP_ACCESS_TOKEN=your-access-token
WHATSAPP_PHONE_NUMBER_ID=your-phone-number-id
WHATSAPP_VERIFY_TOKEN=your-verify-token
WHATSAPP_WEBHOOK_SECRET=your-webhook-secret
```

## Deploying to Leapcell

### 1. Connect Your Repository
- Log in to [Leapcell](https://leapcell.io)
- Create a new service from your GitHub repository
- Select the `backend` directory as the root

### 2. Configure Build Settings
- **Root Directory:** `backend/`
- **Build Command:** `go mod tidy && go build -tags netgo -ldflags '-s -w' -o app ./cmd`
- **Start Command:** `./app`
- **Port:** `8080`

### 3. Set Environment Variables
In your Leapcell service settings, add all required environment variables listed above.

**Important for CORS:**
```bash
CORS_ORIGINS=https://atlasnotes-eta.vercel.app
```
Or for multiple frontends:
```bash
CORS_ORIGINS=https://atlasnotes-eta.vercel.app,https://yourdomain.com,https://staging.yourdomain.com
```

### 4. Deploy
- Commit your changes to the main branch
- Leapcell will automatically detect changes and rebuild
- Check the deployment logs for any errors

### 5. Verify Deployment
Test your CORS configuration:
```bash
curl -H "Origin: https://atlasnotes-eta.vercel.app" \
     -H "Access-Control-Request-Method: POST" \
     -H "Access-Control-Request-Headers: Content-Type,Authorization" \
     -X OPTIONS \
     https://your-backend-url/api/notebooks
```

You should see CORS headers in the response:
```
Access-Control-Allow-Origin: https://atlasnotes-eta.vercel.app
Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS
Access-Control-Allow-Credentials: true
```

## Deploying Frontend to Vercel

### 1. Configure Environment Variables
In your Vercel project settings, add:
```bash
VITE_API_BASE_URL=https://your-backend-url.apn.leapcell.dev
```

### 2. Build Configuration
Vercel should auto-detect your Vite configuration. If needed:
- **Framework Preset:** Vite
- **Root Directory:** `frontend/`
- **Build Command:** `npm run build` or `bun run build`
- **Output Directory:** `dist`

### 3. Deploy
```bash
# Using Vercel CLI
cd frontend
vercel --prod
```

Or connect your GitHub repository to Vercel for automatic deployments.

## Common Issues

### CORS Errors
**Problem:** Getting CORS errors even with `CORS_ORIGINS` set

**Solution:**
1. Verify the environment variable is set correctly in Leapcell
2. Ensure no trailing slashes: `https://domain.com` ✅ not `https://domain.com/` ❌
3. Include the protocol: `https://` not just `domain.com`
4. Redeploy the backend after changing environment variables
5. Check backend logs to confirm the origins are loaded:
   ```
   CORS configured with multiple origins origins=[https://atlasnotes-eta.vercel.app http://localhost:5173]
   ```

### Database Connection Issues
**Problem:** Cannot connect to database

**Solution:**
1. Verify all database credentials are correct
2. Ensure your hosting platform allows outbound connections
3. Check if your database requires SSL/TLS
4. Verify the database is accessible from your hosting platform's IP

### Build Failures
**Problem:** Build fails on Leapcell

**Solution:**
1. Ensure `go.mod` and `go.sum` are committed
2. Check build logs for specific error messages
3. Verify Go version compatibility (requires Go 1.21+)
4. Make sure all imports are available

## Health Check

After deployment, verify your backend is running:

```bash
# Check if server is up
curl https://your-backend-url/health

# Check authentication (should return 401 without token)
curl https://your-backend-url/api/notebooks
```

## Monitoring

Monitor your deployment:
- Check Leapcell logs for errors
- Monitor response times
- Set up alerts for downtime
- Track API usage and rate limits

## Scaling

Leapcell automatically scales based on traffic. For high-traffic applications:
- Consider upgrading to persistent servers
- Implement caching strategies
- Optimize database queries
- Use connection pooling

## Support

If you encounter issues:
1. Check the deployment logs in Leapcell dashboard
2. Verify all environment variables are set correctly
3. Test CORS configuration with curl
4. Join the Leapcell Discord for support

